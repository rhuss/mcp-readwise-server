# Data Model: Readwise MCP Server

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13

## Entities

### Source

Represents a book, article, PDF, tweet, or podcast in the user's Readwise library. Maps to the Readwise v2 `/books/` API response.

| Field | Type | Description | Source |
| ----- | ---- | ----------- | ------ |
| `ID` | int64 | Unique source identifier | `user_book_id` from API |
| `Title` | string | Source title | `title` |
| `ReadableTitle` | string | Display-friendly title | `readable_title` |
| `Author` | string | Author name | `author` |
| `Category` | string | One of: books, articles, tweets, supplementals, podcasts | `category` |
| `Source` | string | Import source (e.g., "kindle", "raindrop") | `source` |
| `CoverImageURL` | string | Cover image URL | `cover_image_url` |
| `SourceURL` | string | Original source URL | `source_url` |
| `ReadwiseURL` | string | URL to Readwise page | `readwise_url` |
| `UniqueURL` | string | Unique URL for deduplication | `unique_url` |
| `HighlightCount` | int | Number of highlights | `num_highlights` |
| `Tags` | []Tag | Tags applied to this source | `book_tags` |
| `DocumentNote` | string | User's note about the source | `document_note` |
| `Summary` | string | Source summary | `summary` |
| `LastHighlightAt` | time.Time | When last highlight was made | `last_highlight_at` |
| `UpdatedAt` | time.Time | Last update timestamp | `updated` |

**Validation**:
- `ID` must be positive
- `Category` must be one of the allowed values

**Relationships**:
- Has many Highlights (via `book_id` on Highlight)
- Has many Tags

---

### Highlight

A text excerpt from a source with optional note, location, and tags. Maps to the Readwise v2 `/highlights/` API response.

| Field | Type | Description | Source |
| ----- | ---- | ----------- | ------ |
| `ID` | int64 | Unique highlight identifier | `id` |
| `Text` | string | Highlight text (max 8191 chars) | `text` |
| `Note` | string | User's note on the highlight | `note` |
| `Location` | int | Position in source | `location` |
| `LocationType` | string | One of: page, order, time_offset | `location_type` |
| `Color` | string | Highlight color | `color` |
| `HighlightedAt` | time.Time | When highlight was created | `highlighted_at` |
| `CreatedAt` | time.Time | Creation timestamp | `created_at` |
| `UpdatedAt` | time.Time | Last update timestamp | `updated_at` |
| `BookID` | int64 | Parent source ID | `book_id` |
| `URL` | string | Highlight URL | `url` |
| `ReadwiseURL` | string | Readwise page URL | `readwise_url` |
| `Tags` | []Tag | Tags applied to this highlight | `tags` |
| `IsFavorite` | bool | Favorited by user | `is_favorite` |
| `IsDiscard` | bool | Discarded by user | `is_discard` |
| `ExternalID` | string | External system ID | `external_id` |

**Validation**:
- `ID` must be positive
- `Text` is required and max 8191 characters
- `LocationType` must be one of the allowed values if set

**Relationships**:
- Belongs to one Source (via `BookID`)
- Has many Tags

---

### Document

A Reader item (article, email, RSS entry, PDF, epub, tweet, video). Maps to the Reader v3 `/list/` API response.

| Field | Type | Description | Source |
| ----- | ---- | ----------- | ------ |
| `ID` | string | Unique document identifier | `id` |
| `URL` | string | Document URL | `url` |
| `SourceURL` | string | Original source URL | `source_url` |
| `Title` | string | Document title | `title` |
| `Author` | string | Author name | `author` |
| `Source` | string | Import source | `source` |
| `Category` | string | One of: article, email, rss, highlight, note, pdf, epub, tweet, video | `category` |
| `Location` | string | One of: new, later, shortlist, archive, feed | `location` |
| `Tags` | map[string]Tag | Tags applied to this document | `tags` |
| `SiteName` | string | Website name | `site_name` |
| `WordCount` | int | Word count | `word_count` |
| `CreatedAt` | time.Time | Creation timestamp | `created_at` |
| `UpdatedAt` | time.Time | Last update timestamp | `updated_at` |
| `PublishedDate` | string | Publication date | `published_date` |
| `Summary` | string | Document summary | `summary` |
| `ImageURL` | string | Cover image URL | `image_url` |
| `Notes` | string | User notes | `notes` |
| `ParentID` | string | Parent document ID (for highlights/notes) | `parent_id` |
| `ReadingProgress` | float64 | Reading progress 0.0 to 1.0 | `reading_progress` |
| `FirstOpenedAt` | time.Time | First opened timestamp | `first_opened_at` |
| `LastOpenedAt` | time.Time | Last opened timestamp | `last_opened_at` |
| `LastMovedAt` | time.Time | Last moved timestamp | `last_moved_at` |
| `SavedAt` | time.Time | When saved | `saved_at` |
| `Content` | string | HTML content (optional, only when requested) | `html` |

**Validation**:
- `ID` is required
- `Category` must be one of the allowed values
- `Location` must be one of the allowed values
- `ReadingProgress` must be between 0.0 and 1.0

**Relationships**:
- May have a parent Document (via `ParentID`)
- Has many Tags

---

### Tag

A label applied to sources, highlights, or documents for organization.

| Field | Type | Description | Source |
| ----- | ---- | ----------- | ------ |
| `ID` | int64 | Tag identifier (v2 API) | `id` |
| `Name` | string | Tag name/label | `name` |

**Validation**:
- `Name` is required and non-empty

**Notes**: In the v2 API, tags have numeric IDs. In the v3 API, tags use string keys. The data model uses both representations internally.

---

### Profile

A named group of MCP tools that can be enabled or disabled as a unit. This is a configuration entity, not an API entity.

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Name` | string | Profile identifier (readwise, reader, write, video, destructive) |
| `Type` | string | One of: read, modifier |
| `Dependencies` | []string | Required profiles that must also be active |
| `Tools` | []ToolDef | Tools belonging to this profile |

**Shortcuts**:

| Shortcut | Expands To |
| -------- | ---------- |
| `basic` | reader, write |
| `all` | readwise, reader, write, video, destructive |

**Dependency Rules**:
- `write` requires `readwise` or `reader`
- `video` requires `reader`
- `destructive` requires `readwise` or `reader` (corresponding to its tools)

**State Transitions**: Profile resolution follows this sequence:
1. Parse comma-separated profile names from `READWISE_PROFILES` env var
2. Expand shortcuts (basic, all)
3. Deduplicate
4. Validate dependencies (fail fast if unsatisfied)
5. Collect and register matching tools

---

### CacheEntry

A stored API response with TTL and size tracking. Internal entity for cache management.

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Key` | string | SHA-256 hash of (api_key_hash + endpoint + sorted_query_params) |
| `Data` | []byte | Serialized response data |
| `Size` | int64 | Size in bytes of the serialized data |
| `CreatedAt` | time.Time | When entry was cached |
| `TTL` | time.Duration | Time-to-live for this entry |
| `LastAccessedAt` | time.Time | Last time entry was read (for LRU) |

**Cache Key Construction**:
```
key = SHA256(SHA256(api_key) + endpoint + sorted(query_params))
```

**TTL Defaults**:

| Cached Endpoint | Default TTL |
| --------------- | ----------- |
| `/api/v2/export/` | 5 minutes |
| `/api/v2/books/` | 5 minutes |
| `/api/v3/list/` | 5 minutes |
| `/api/v3/tags/` | 10 minutes |

**Invalidation Map** (write/destructive operation -> cached endpoints to invalidate):

| Operation | Invalidates |
| --------- | ----------- |
| `create_highlight`, `update_highlight`, `delete_highlight`, `bulk_create_highlights` | `/api/v2/export/` |
| `add_source_tag`, `delete_source_tag` | `/api/v2/books/` |
| `add_highlight_tag`, `delete_highlight_tag` | `/api/v2/export/` |
| `save_document`, `update_document`, `delete_document` | `/api/v3/list/` |

**Eviction**: LRU when total cache size exceeds configured memory limit (default 128 MB). Minimum entry lifetime of 30 seconds to prevent thrashing.

---

### ErrorResponse

Structured error returned to MCP clients. Internal entity for error formatting.

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Type` | string | One of: validation_error, auth_error, api_error, internal_error |
| `Code` | string | Machine-readable error code (e.g., rate_limited, invalid_param) |
| `Message` | string | Human-readable error description |
| `RetryAfter` | int | Seconds to wait before retrying (only for rate_limited) |

---

### Config

Server configuration loaded from environment variables. Internal entity.

| Field | Type | Env Var | Default |
| ----- | ---- | ------- | ------- |
| `Profiles` | []string | `READWISE_PROFILES` | ["readwise"] |
| `Port` | int | `PORT` | 8080 |
| `LogLevel` | string | `LOG_LEVEL` | "info" |
| `CacheMaxSizeMB` | int | `CACHE_MAX_SIZE_MB` | 128 |
| `CacheTTLSeconds` | int | `CACHE_TTL_SECONDS` | 300 |
| `CacheEnabled` | bool | `CACHE_ENABLED` | true |
