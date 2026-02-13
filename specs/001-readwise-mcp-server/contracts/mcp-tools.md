# MCP Tool Contracts: Readwise MCP Server

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13

This document defines the 29 MCP tool contracts organized by profile. Each tool lists its input parameters, return shape, cache behavior, and error cases.

---

## Profile: readwise (9 tools)

### list_sources

List highlight sources with pagination and filtering.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `page_size` | int | no | 100 | 1-1000 |
| `page` | int | no | 1 | >= 1 |
| `category` | string | no | - | books, articles, tweets, supplementals, podcasts |
| `updated_after` | string | no | - | ISO 8601 datetime |

**Output**: `{ count, next, previous, results: [Source] }`

**Cache**: Reads from cache (`/api/v2/books/`, TTL 5 min). Populates cache on miss.

**Errors**: `validation_error` (invalid params), `auth_error` (bad key), `api_error` (upstream failure)

---

### get_source

Get details of a single source.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |

**Output**: `Source` object

**Cache**: No dedicated cache. Single-item fetch.

**Errors**: `validation_error` (missing id), `auth_error`, `api_error` (404 if not found)

---

### list_highlights

List highlights with pagination and filtering.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `page_size` | int | no | 100 | 1-1000 |
| `page` | int | no | 1 | >= 1 |
| `source_id` | string | no | - | non-empty if set |
| `updated_after` | string | no | - | ISO 8601 datetime |

**Output**: `{ count, next, previous, results: [Highlight] }`

**Cache**: No dedicated cache. Per-request fetch.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### get_highlight

Get a single highlight by ID.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |

**Output**: `Highlight` object

**Cache**: No dedicated cache.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### export_highlights

Bulk export all highlights, grouped by source. Primary data source for search.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `updated_after` | string | no | ISO 8601 datetime |

**Output**: `{ count, results: [{ source fields..., highlights: [Highlight] }] }`

Note: Internally paginates through all cursor pages and returns the complete result.

**Cache**: Cached at `/api/v2/export/`, TTL 5 min, keyed by api_key_hash + updated_after.

**Errors**: `auth_error`, `api_error`

---

### get_daily_review

Get today's daily review highlights.

**Input**: None

**Output**: `{ review_id, review_url, review_completed, highlights: [ReviewHighlight] }`

**Cache**: No cache (daily review changes frequently).

**Errors**: `auth_error`, `api_error`

---

### list_source_tags

List tags on a source.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `source_id` | string | yes | non-empty |

**Output**: `[Tag]`

**Cache**: No dedicated cache.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### list_highlight_tags

List tags on a highlight.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `highlight_id` | string | yes | non-empty |

**Output**: `[Tag]`

**Cache**: No dedicated cache.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### search_highlights

Search highlights by query. Operates over cached export data.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `query` | string | yes | - | non-empty |
| `source_id` | string | no | - | non-empty if set |
| `limit` | int | no | 50 | 1-200 |

**Output**: `[{ highlight: Highlight, source_title: string, relevance_score: float }]`

**Behavior**:
1. Fetch or retrieve cached export data
2. Search across highlight text, notes, and source titles (case-insensitive)
3. Score results: exact matches ranked above partial matches
4. Return up to `limit` results sorted by relevance score descending

**Cache**: Reads from export cache. Triggers export cache population if empty.

**Errors**: `validation_error` (empty query), `auth_error`, `api_error` (if cache population fails)

---

## Profile: reader (4 tools)

### list_documents

List Reader documents with filtering.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `location` | string | no | - | new, later, shortlist, archive, feed |
| `category` | string | no | - | article, email, rss, highlight, note, pdf, epub, tweet, video |
| `updated_after` | string | no | - | ISO 8601 datetime |
| `limit` | int | no | 100 | 1-100 |

**Output**: `{ count, results: [Document] }`

Note: Internally paginates through cursor pages up to `limit` results.

**Cache**: Cached at `/api/v3/list/`, TTL 5 min.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### get_document

Get a single Reader document with optional content.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `id` | string | yes | - | non-empty |
| `include_content` | bool | no | false | - |

**Output**: `Document` (with `Content` field populated if `include_content` is true)

**Cache**: No dedicated cache. Single-item fetch.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### list_reader_tags

List all tags in Reader.

**Input**: None

**Output**: `[Tag]`

**Cache**: Cached at `/api/v3/tags/`, TTL 10 min.

**Errors**: `auth_error`, `api_error`

---

### search_documents

Search Reader documents by query. Operates over cached list data.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `query` | string | yes | - | non-empty |
| `location` | string | no | - | new, later, shortlist, archive, feed |
| `category` | string | no | - | article, email, rss, highlight, note, pdf, epub, tweet, video |
| `limit` | int | no | 50 | 1-200 |

**Output**: `[{ document: Document, relevance_score: float }]`

**Behavior**:
1. Fetch or retrieve cached document list
2. Apply location/category filters if specified
3. Search across title, author, summary, and notes (case-insensitive)
4. Score results: exact matches ranked above partial matches
5. Return up to `limit` results sorted by relevance score descending

**Cache**: Reads from document list cache. Triggers cache population if empty.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

## Profile: write (7 tools)

**Dependency**: Requires `readwise` or `reader` profile.

### save_document

Save a URL to Reader.

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `url` | string | yes | - | valid URL |
| `title` | string | no | - | - |
| `author` | string | no | - | - |
| `summary` | string | no | - | - |
| `tags` | []string | no | - | - |
| `location` | string | no | - | new, later, shortlist, archive |
| `category` | string | no | - | - |

**Output**: `{ id, url }`

**Cache Invalidation**: Clears `/api/v3/list/` cache for this API key.

**Errors**: `validation_error` (invalid URL), `auth_error`, `api_error`

---

### update_document

Update Reader document metadata.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |
| `title` | string | no | - |
| `author` | string | no | - |
| `summary` | string | no | - |
| `location` | string | no | new, later, shortlist, archive |
| `tags` | []string | no | - |
| `seen` | bool | no | - |

**Output**: `Document`

**Cache Invalidation**: Clears `/api/v3/list/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### create_highlight

Create a new highlight.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `text` | string | yes | non-empty, max 8191 chars |
| `source_id` | string | no | - |
| `source_title` | string | no | required if source_id not set |
| `source_author` | string | no | - |
| `source_url` | string | no | - |
| `note` | string | no | - |
| `location` | int | no | - |
| `location_type` | string | no | page, order, time_offset |
| `highlighted_at` | string | no | ISO 8601 datetime |

**Output**: `Highlight`

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### update_highlight

Update an existing highlight.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |
| `text` | string | no | max 8191 chars |
| `note` | string | no | - |
| `location` | int | no | - |
| `color` | string | no | - |

**Output**: `Highlight`

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### add_source_tag

Add a tag to a source.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `source_id` | string | yes | non-empty |
| `name` | string | yes | non-empty |

**Output**: `Tag`

**Cache Invalidation**: Clears `/api/v2/books/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### add_highlight_tag

Add a tag to a highlight.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `highlight_id` | string | yes | non-empty |
| `name` | string | yes | non-empty |

**Output**: `Tag`

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### bulk_create_highlights

Create multiple highlights in a single request.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `highlights` | []object | yes | non-empty array, each with required `text` and `source_title` |

Each highlight object:
| Field | Type | Required | Validation |
| ----- | ---- | -------- | ---------- |
| `text` | string | yes | non-empty, max 8191 chars |
| `source_title` | string | yes | non-empty |
| `source_author` | string | no | - |
| `source_url` | string | no | - |
| `note` | string | no | - |
| `location` | int | no | - |
| `highlighted_at` | string | no | ISO 8601 datetime |

**Output**: `[{ id }]`

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

## Profile: video (5 tools)

**Dependency**: Requires `reader` profile. Write-requiring tools additionally require `write` profile.

### list_videos

List video documents in Reader. **Read-only** (requires reader only).

**Input**:
| Parameter | Type | Required | Default | Validation |
| --------- | ---- | -------- | ------- | ---------- |
| `location` | string | no | - | new, later, shortlist, archive, feed |
| `limit` | int | no | 50 | 1-100 |

**Output**: `{ count, results: [Document] }` (filtered to category=video)

**Cache**: Uses document list cache filtered to video category.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### get_video

Get video details including transcript. **Read-only** (requires reader only).

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |

**Output**: `Document` with transcript content if available.

**Cache**: No dedicated cache.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### get_video_position

Get playback position for a video. **Read-only** (requires reader only).

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |

**Output**: `{ reading_progress: float, last_opened_at: datetime }`

**Cache**: No cache (position changes frequently).

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### update_video_position

Update playback position for a video. **Requires write profile**.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |
| `position` | float | yes | >= 0.0 |
| `duration` | float | no | >= 0.0 |

**Output**: `{ reading_progress: float }`

**Cache Invalidation**: Clears `/api/v3/list/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

### create_video_highlight

Create a highlight at a specific video timestamp. **Requires write profile**.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty (video document ID) |
| `text` | string | yes | non-empty, max 8191 chars |
| `timestamp` | float | yes | >= 0.0 (seconds) |
| `end_timestamp` | float | no | >= timestamp |
| `note` | string | no | - |

**Output**: `Highlight` with timestamp metadata.

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error`

---

## Profile: destructive (4 tools)

**Dependency**: Requires corresponding read profile (readwise for highlight/tag ops, reader for document ops).

### delete_highlight

Delete a highlight.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |

**Output**: `{ deleted: true }`

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### delete_highlight_tag

Remove a tag from a highlight.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `highlight_id` | string | yes | non-empty |
| `tag_id` | string | yes | non-empty |

**Output**: `{ deleted: true }`

**Cache Invalidation**: Clears `/api/v2/export/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### delete_source_tag

Remove a tag from a source.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `source_id` | string | yes | non-empty |
| `tag_id` | string | yes | non-empty |

**Output**: `{ deleted: true }`

**Cache Invalidation**: Clears `/api/v2/books/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)

---

### delete_document

Delete a Reader document.

**Input**:
| Parameter | Type | Required | Validation |
| --------- | ---- | -------- | ---------- |
| `id` | string | yes | non-empty |

**Output**: `{ deleted: true }`

**Cache Invalidation**: Clears `/api/v3/list/` cache for this API key.

**Errors**: `validation_error`, `auth_error`, `api_error` (404)
