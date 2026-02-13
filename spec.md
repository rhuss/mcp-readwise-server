# Feature Specification: Readwise MCP Server

**Feature Branch**: `main`
**Created**: 2026-02-13
**Status**: Draft
**Input**: User description: "A Go-based MCP server providing AI assistants access to Readwise and Reader APIs with configurable tool profiles and server-side caching."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse Readwise Highlights (Priority: P1)

An AI assistant user wants to browse their Readwise highlights and sources. They connect their MCP client (e.g. Claude Desktop) to the server, provide their Readwise API key, and use tools to list sources, view highlights, and search across their library.

**Why this priority**: Core read functionality is the foundation. Without browsing highlights, no other feature has value.

**Independent Test**: Can be fully tested by connecting to the server with a valid API key and listing sources/highlights. Delivers immediate value for knowledge retrieval.

**Acceptance Scenarios**:

1. **Given** a configured MCP client with a valid Readwise API key, **When** the user calls `list_sources`, **Then** a paginated list of sources (books, articles, etc.) is returned with id, title, author, category, and highlight count.
2. **Given** a source with highlights, **When** the user calls `list_highlights` with a `source_id` filter, **Then** only highlights for that source are returned.
3. **Given** a large highlight library, **When** the user calls `search_highlights` with a query, **Then** matching highlights are returned ranked by relevance, searching across text, notes, and source titles.
4. **Given** a valid API key, **When** the user calls `get_daily_review`, **Then** today's review highlights are returned.

---

### User Story 2 - Browse Reader Documents (Priority: P1)

An AI assistant user wants to browse their Reader library. They use tools to list documents, filter by location (new, later, archive) or category (article, PDF, email), and view document details including content.

**Why this priority**: Reader is the second major API surface and equally important for users who use Reader alongside Readwise.

**Independent Test**: Can be fully tested by connecting with a valid API key and listing/searching Reader documents. Delivers value for document discovery.

**Acceptance Scenarios**:

1. **Given** a valid API key with Reader documents, **When** the user calls `list_documents` with a `location` filter of "later", **Then** only documents in the "later" list are returned.
2. **Given** a document ID, **When** the user calls `get_document` with `include_content: true`, **Then** the document details include the HTML content.
3. **Given** documents in Reader, **When** the user calls `search_documents` with a query, **Then** matching documents are returned ranked by relevance.
4. **Given** tags in Reader, **When** the user calls `list_reader_tags`, **Then** all tags are returned.

---

### User Story 3 - Create and Modify Content (Priority: P2)

An AI assistant user wants to save new documents, create highlights, and manage tags. They use write profile tools to save URLs to Reader, create highlights with notes, and organize content with tags.

**Why this priority**: Write operations extend value by allowing the assistant to help organize and capture knowledge, but read access must work first.

**Independent Test**: Can be tested by saving a URL to Reader, creating a highlight, and adding tags. Verifiable by subsequent read operations.

**Acceptance Scenarios**:

1. **Given** the `write` profile is enabled alongside `reader`, **When** the user calls `save_document` with a URL, **Then** the document is saved to Reader and its ID is returned.
2. **Given** the `write` profile is enabled alongside `readwise`, **When** the user calls `create_highlight` with text and a source, **Then** the highlight is created and visible in subsequent list calls.
3. **Given** a source ID and tag name, **When** the user calls `add_source_tag`, **Then** the tag is applied to the source.
4. **Given** multiple highlights to create, **When** the user calls `bulk_create_highlights`, **Then** all highlights are created in a single request and IDs are returned.

---

### User Story 4 - Profile-Based Tool Configuration (Priority: P2)

A server operator wants to control which tools are exposed to MCP clients. They configure the `READWISE_PROFILES` environment variable to select specific profiles, and only tools matching active profiles are registered.

**Why this priority**: Profiles are the primary configuration mechanism. They gate access to write and destructive operations and must work correctly for safe operation.

**Independent Test**: Can be tested by starting the server with different profile configurations and verifying which tools are advertised.

**Acceptance Scenarios**:

1. **Given** `READWISE_PROFILES=readwise`, **When** the server starts, **Then** only the 9 readwise profile tools are registered.
2. **Given** `READWISE_PROFILES=basic`, **When** the server starts, **Then** the shortcut expands to `reader,write` and the corresponding tools are registered.
3. **Given** `READWISE_PROFILES=write`, **When** the server starts, **Then** validation fails because `write` requires `readwise` or `reader` as a dependency.
4. **Given** `READWISE_PROFILES=readwise,reader,write`, **When** the server starts, **Then** tools from all three profiles are registered with no duplicates.

---

### User Story 5 - Video Document Management (Priority: P3)

An AI assistant user wants to browse video documents in Reader, view transcripts, and create timestamped highlights.

**Why this priority**: Video is a specialized subset of Reader functionality. Valuable for video-heavy users but not a core requirement.

**Independent Test**: Can be tested by listing videos, getting a video with transcript, and creating a timestamped highlight.

**Acceptance Scenarios**:

1. **Given** the `video` profile is enabled with `reader`, **When** the user calls `list_videos`, **Then** only video-category documents are returned.
2. **Given** a video document ID, **When** the user calls `get_video`, **Then** video details including transcript (if available) are returned.
3. **Given** a video ID and the `write` profile, **When** the user calls `create_video_highlight` with text and a timestamp, **Then** a highlight is created with timestamp metadata.

---

### User Story 6 - Destructive Operations (Priority: P3)

An AI assistant user wants to delete highlights, documents, or tags. These operations require explicit opt-in via the `destructive` profile.

**Why this priority**: Delete operations are important for library management but require the highest care. Explicit opt-in prevents accidental data loss.

**Independent Test**: Can be tested by creating and then deleting a highlight, verifying it no longer appears in list calls.

**Acceptance Scenarios**:

1. **Given** the `destructive` profile is NOT enabled, **When** the client attempts to call `delete_highlight`, **Then** the tool is not available.
2. **Given** the `destructive` profile IS enabled with `readwise`, **When** the user calls `delete_highlight` with a valid ID, **Then** the highlight is deleted and the export cache is invalidated.
3. **Given** the `destructive` profile IS enabled with `reader`, **When** the user calls `delete_document` with a valid ID, **Then** the document is deleted and the document list cache is invalidated.

---

### User Story 7 - Efficient Caching (Priority: P2)

The server caches read-heavy API responses to reduce Readwise API calls and speed up search operations. Write operations invalidate relevant caches. Cache memory usage is bounded.

**Why this priority**: Caching is essential for the search tools (which operate over cached data) and for staying within Readwise API rate limits.

**Independent Test**: Can be tested by making repeated read calls and verifying reduced API calls, then making a write call and verifying cache invalidation.

**Acceptance Scenarios**:

1. **Given** a call to `export_highlights`, **When** the same call is made within 5 minutes, **Then** the cached result is returned without hitting the Readwise API.
2. **Given** a cached export, **When** the user calls `create_highlight`, **Then** the export cache for that API key is invalidated.
3. **Given** cache memory exceeds the configured limit, **When** a new entry is added, **Then** the least recently used entries are evicted until memory is under the limit.
4. **Given** two different API keys, **When** both call `export_highlights`, **Then** each gets their own cache entry (cache keys are isolated by API key hash).

---

### Edge Cases

- What happens when the Readwise API returns a 429 (rate limit)? The server passes it through to the client with the `Retry-After` value.
- What happens when the API key is missing or invalid? The server returns an `auth_error` with a clear message.
- What happens when a write profile is enabled without a corresponding read profile? The server rejects the configuration at startup.
- What happens when the cache memory limit is set to 0? Caching is effectively disabled, but the server still functions.
- What happens when a client provides invalid parameters (e.g., `page_size: -1`)? The server returns a `validation_error`.
- What happens when the Readwise API is unreachable? The server returns an `api_error` and does not cache the failure.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose Readwise v2 API operations as MCP tools: `list_sources`, `get_source`, `list_highlights`, `get_highlight`, `export_highlights`, `get_daily_review`, `list_source_tags`, `list_highlight_tags`, `search_highlights`
- **FR-002**: System MUST expose Reader v3 API operations as MCP tools: `list_documents`, `get_document`, `list_reader_tags`, `search_documents`
- **FR-003**: System MUST expose write operations as MCP tools: `save_document`, `update_document`, `create_highlight`, `update_highlight`, `add_source_tag`, `add_highlight_tag`, `bulk_create_highlights`
- **FR-004**: System MUST expose video operations as MCP tools: `list_videos`, `get_video`, `get_video_position`, `update_video_position`, `create_video_highlight`
- **FR-005**: System MUST expose destructive operations as MCP tools: `delete_highlight`, `delete_highlight_tag`, `delete_source_tag`, `delete_document`
- **FR-006**: System MUST support configurable tool profiles via `READWISE_PROFILES` environment variable
- **FR-007**: System MUST resolve profile shortcuts (`basic` -> `reader,write`; `all` -> all profiles)
- **FR-008**: System MUST validate profile dependencies (write requires a read profile)
- **FR-009**: System MUST extract the Readwise API key from client request headers (passthrough, no server-side storage)
- **FR-010**: System MUST cache responses for `export_highlights`, `list_sources`, `list_documents`, and `list_reader_tags` with configurable TTL
- **FR-011**: System MUST invalidate relevant caches when write/destructive operations succeed
- **FR-012**: System MUST enforce configurable memory limits on the cache with LRU eviction
- **FR-013**: System MUST pass through Readwise API rate limit responses (429) to the client with `Retry-After` information
- **FR-014**: System MUST return structured error responses with type, code, and message fields
- **FR-015**: System MUST provide health (`/health`) and readiness (`/ready`) endpoints
- **FR-016**: System MUST support `search_highlights` via client-side search over cached export data
- **FR-017**: System MUST support `search_documents` via client-side search over cached document list data
- **FR-018**: System MUST isolate cache entries per API key using hashed keys (SHA-256)

### Key Entities

- **Source**: A book, article, PDF, tweet, or podcast in Readwise. Has title, author, category, tags, and a collection of highlights.
- **Highlight**: A text excerpt from a source with optional note, location, color, and tags. The core unit of captured knowledge.
- **Document**: A Reader item (article, email, RSS entry, PDF, epub, tweet, video) with title, author, URL, location, reading progress, and optional content.
- **Tag**: A label applied to sources, highlights, or documents for organization.
- **Profile**: A named group of MCP tools that can be enabled/disabled as a unit. Profiles have type (read, modifier) and optional dependencies.
- **Cache Entry**: A stored API response keyed by API key hash + endpoint + query parameters, with TTL and size tracking.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 29 MCP tools (9 readwise + 4 reader + 7 write + 5 video + 4 destructive) function correctly against the live Readwise API
- **SC-002**: Profile system correctly enables/disables tools for all valid profile combinations
- **SC-003**: Cached search operations (`search_highlights`, `search_documents`) return relevant results ranked by relevance
- **SC-004**: Cache memory stays within configured limits under sustained load
- **SC-005**: Server deploys successfully to k3s cluster via Kustomize with TLS termination
- **SC-006**: Health and readiness probes respond correctly
- **SC-007**: Unit test coverage for tools, API client, cache, and profile resolution
- **SC-008**: Integration tests pass against a mock Readwise API server

---

## Architecture Overview

### System Context

```
┌─────────────────────────────────────────────────────────────────┐
│                      Claude Desktop / Client                     │
│                    (provides READWISE_API_KEY)                   │
└─────────────────────────────────────────────────────────────────┘
                              │ SSE/HTTP
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    k3s Cluster (mcp namespace)                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                   readwise-mcp Pod                        │   │
│  │  ┌────────────────────┐  ┌─────────────────────────────┐ │   │
│  │  │   nginx (sidecar)  │  │  readwise-mcp (Go binary)   │ │   │
│  │  │   :8443 (TLS)      │──│  :8080 (HTTP)               │ │   │
│  │  └────────────────────┘  └─────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                    LoadBalancer: 10.9.11.XXX                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │   Readwise API    │
                    │  readwise.io/api  │
                    └───────────────────┘
```

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        readwise-mcp                              │
│                                                                  │
│  ┌────────────────┐    ┌────────────────┐    ┌──────────────┐  │
│  │   MCP Server   │───▶│  Tool Registry │───▶│    Tools     │  │
│  │   (SSE/HTTP)   │    │                │    │              │  │
│  └────────────────┘    └────────────────┘    └──────┬───────┘  │
│          │                                          │          │
│          │                                          ▼          │
│          │             ┌────────────────┐    ┌──────────────┐  │
│          │             │  Cache Manager │◀───│  API Client  │  │
│          │             │   (LRU, 128MB) │    │  (v2 + v3)   │  │
│          │             └────────────────┘    └──────────────┘  │
│          │                                          │          │
│          └──────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `READWISE_PROFILES` | Comma-separated profiles to enable | `readwise` | No |
| `PORT` | HTTP port for MCP server | `8080` | No |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` | No |
| `CACHE_MAX_SIZE_MB` | Maximum cache memory in MB | `128` | No |
| `CACHE_TTL_SECONDS` | Cache entry time-to-live | `300` | No |
| `CACHE_ENABLED` | Enable/disable caching | `true` | No |

### Profile System

#### Base Profiles

| Profile | Type | Description |
|---------|------|-------------|
| `readwise` | Read | Readwise v2 API (sources, highlights, export, daily review) |
| `reader` | Read | Reader v3 API (documents, tags) |
| `write` | Modifier | Create/update operations (requires read profile) |
| `video` | Modifier | Video-specific tools (requires reader) |
| `destructive` | Modifier | Delete operations (explicit opt-in) |

#### Shortcut Profiles

| Shortcut | Expands To |
|----------|------------|
| `basic` | `reader,write` |
| `all` | `readwise,reader,write,video,destructive` |

## Tool Inventory

### Profile: readwise (9 tools)

| Tool | Description | API Endpoint |
|------|-------------|--------------|
| `list_sources` | List highlight sources with pagination/filtering | `GET /api/v2/books/` |
| `get_source` | Get single source details | `GET /api/v2/books/{id}/` |
| `list_highlights` | List highlights with pagination/filtering | `GET /api/v2/highlights/` |
| `get_highlight` | Get single highlight | `GET /api/v2/highlights/{id}/` |
| `export_highlights` | Bulk export highlights (cached) | `GET /api/v2/export/` |
| `get_daily_review` | Get today's review highlights | `GET /api/v2/review/` |
| `list_source_tags` | List tags on a source | `GET /api/v2/books/{id}/tags` |
| `list_highlight_tags` | List tags on a highlight | `GET /api/v2/highlights/{id}/tags` |
| `search_highlights` | Search highlights via cached export data | Client-side search |

### Profile: reader (4 tools)

| Tool | Description | API Endpoint |
|------|-------------|--------------|
| `list_documents` | List Reader documents with filtering | `GET /api/v3/list/` |
| `get_document` | Get document with optional content | `GET /api/v3/list/?id={id}` |
| `list_reader_tags` | List all Reader tags | `GET /api/v3/tags/` |
| `search_documents` | Search documents via cached list | Client-side search |

### Profile: write (7 tools)

| Tool | Description | API Endpoint |
|------|-------------|--------------|
| `save_document` | Save URL to Reader | `POST /api/v3/save/` |
| `update_document` | Update document metadata | `PATCH /api/v3/update/{id}/` |
| `create_highlight` | Create a highlight | `POST /api/v2/highlights/` |
| `update_highlight` | Update a highlight | `PATCH /api/v2/highlights/{id}/` |
| `add_source_tag` | Add tag to source | `POST /api/v2/books/{id}/tags/` |
| `add_highlight_tag` | Add tag to highlight | `POST /api/v2/highlights/{id}/tags/` |
| `bulk_create_highlights` | Batch create highlights | `POST /api/v2/highlights/` |

### Profile: video (5 tools)

| Tool | Description | API Endpoint |
|------|-------------|--------------|
| `list_videos` | List video documents | `GET /api/v3/list/?category=video` |
| `get_video` | Get video with transcript | `GET /api/v3/list/?id={id}` |
| `get_video_position` | Get playback position | `GET /api/v3/list/?id={id}` |
| `update_video_position` | Update playback position | `PATCH /api/v3/update/{id}/` |
| `create_video_highlight` | Create timestamped highlight | Custom implementation |

### Profile: destructive (4 tools)

| Tool | Description | API Endpoint |
|------|-------------|--------------|
| `delete_highlight` | Delete a highlight | `DELETE /api/v2/highlights/{id}/` |
| `delete_highlight_tag` | Remove tag from highlight | `DELETE /api/v2/highlights/{id}/tags/{tag_id}` |
| `delete_source_tag` | Remove tag from source | `DELETE /api/v2/books/{id}/tags/{tag_id}` |
| `delete_document` | Delete a Reader document | `DELETE /api/v3/delete/{id}/` |

## Error Handling

### Error Response Format

```json
{
  "error": {
    "type": "api_error",
    "code": "rate_limited",
    "message": "Readwise API rate limit exceeded",
    "retry_after": 60
  }
}
```

### Error Types

| Type | Description |
|------|-------------|
| `validation_error` | Invalid parameters |
| `auth_error` | Invalid or missing API key |
| `api_error` | Readwise API error (including 429) |
| `internal_error` | Server error |

## Open Questions

1. **MCP SDK choice**: Should we use `github.com/mark3labs/mcp-go` or another Go MCP implementation?
2. **TLS certificates**: Use Let's Encrypt via cert-manager, or self-signed for internal use?
3. **Metrics**: Should we expose Prometheus metrics for cache hits/misses, API latency?

## References

- [Readwise API Documentation](https://readwise.io/api_deets)
- [Reader API Documentation](https://readwise.io/reader_api)
- [Reference Implementation (TypeScript)](https://github.com/IAmAlexander/readwise-mcp)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
