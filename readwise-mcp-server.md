# Spec: Readwise MCP Server

## Overview

A Go-based Model Context Protocol (MCP) server providing AI assistants access to Readwise and Reader APIs. The server operates statelessly with API key passthrough, supports configurable tool profiles, and includes server-side caching for search operations.

## Goals

1. Provide MCP tools for browsing and managing Readwise highlights and Reader documents
2. Support flexible profile-based tool exposure for different use cases
3. Deploy to k3s home cluster following established patterns
4. Enable API key passthrough (client provides credentials)
5. Implement efficient caching for search and export operations

## Non-Goals

- Server-side credential storage
- Server-side rate limiting (pass through 429s)
- Bulk delete operations
- Custom authentication beyond API key passthrough

---

## Architecture

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

---

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

#### Profile Resolution

Profiles are resolved in order:
1. Expand shortcuts (basic → reader,write)
2. Deduplicate
3. Validate dependencies (write requires readwise or reader)
4. Register matching tools

---

## Tool Specifications

### Profile: readwise (9 tools)

#### list_sources
List highlight sources (books, articles, PDFs, etc.) with pagination and filtering.

**Parameters:**
- `page_size` (int, optional): Results per page (1-1000, default 100)
- `page` (int, optional): Page number (default 1)
- `category` (string, optional): Filter by category (books, articles, tweets, supplementals, podcasts)
- `updated_after` (string, optional): ISO 8601 datetime filter

**Returns:** Paginated list of sources with id, title, author, category, highlight_count, tags.

**API:** `GET /api/v2/books/`

---

#### get_source
Get details of a single highlight source.

**Parameters:**
- `id` (string, required): Source ID

**Returns:** Source details including title, author, category, source_url, highlight_count, tags.

**API:** `GET /api/v2/books/{id}/`

---

#### list_highlights
List highlights with pagination and filtering.

**Parameters:**
- `page_size` (int, optional): Results per page (1-1000, default 100)
- `page` (int, optional): Page number (default 1)
- `source_id` (string, optional): Filter by source
- `updated_after` (string, optional): ISO 8601 datetime filter

**Returns:** Paginated list of highlights with id, text, note, source_id, location, tags, highlighted_at.

**API:** `GET /api/v2/highlights/`

---

#### get_highlight
Get a single highlight by ID.

**Parameters:**
- `id` (string, required): Highlight ID

**Returns:** Highlight details including text, note, source, location, color, tags.

**API:** `GET /api/v2/highlights/{id}/`

---

#### export_highlights
Bulk export all highlights. Results are cached for efficient search.

**Parameters:**
- `updated_after` (string, optional): ISO 8601 datetime, export only items updated after this time

**Returns:** Complete highlight export grouped by source.

**API:** `GET /api/v2/export/`

**Caching:** Results cached with 5-minute TTL, keyed by API key hash + updated_after.

---

#### get_daily_review
Get today's daily review highlights.

**Parameters:** None

**Returns:** List of highlights selected for today's review.

**API:** `GET /api/v2/review/`

---

#### list_source_tags
List tags on a source.

**Parameters:**
- `source_id` (string, required): Source ID

**Returns:** List of tags with id and name.

**API:** `GET /api/v2/books/{id}/tags`

---

#### list_highlight_tags
List tags on a highlight.

**Parameters:**
- `highlight_id` (string, required): Highlight ID

**Returns:** List of tags with id and name.

**API:** `GET /api/v2/highlights/{id}/tags`

---

#### search_highlights
Search highlights by query string. Uses cached export data.

**Parameters:**
- `query` (string, required): Search query
- `source_id` (string, optional): Limit search to specific source
- `limit` (int, optional): Max results (default 50)

**Returns:** Matching highlights with relevance ranking.

**Implementation:** Client-side search over cached export data. Searches text, notes, and source titles.

---

### Profile: reader (4 tools)

#### list_documents
List Reader documents with filtering.

**Parameters:**
- `location` (string, optional): Filter by location (new, later, shortlist, archive, feed)
- `category` (string, optional): Filter by category (article, email, rss, highlight, note, pdf, epub, tweet, video)
- `updated_after` (string, optional): ISO 8601 datetime filter
- `limit` (int, optional): Results limit (1-100, default 100)

**Returns:** Paginated list of documents with id, title, author, url, category, location, tags, reading_progress.

**API:** `GET /api/v3/list/`

**Caching:** Results cached with 5-minute TTL.

---

#### get_document
Get a single Reader document with optional content.

**Parameters:**
- `id` (string, required): Document ID
- `include_content` (bool, optional): Include HTML content (default false)

**Returns:** Document details including title, author, url, summary, content (if requested), reading_progress.

**API:** `GET /api/v3/list/?id={id}`

---

#### list_reader_tags
List all tags in Reader.

**Parameters:** None

**Returns:** List of all tags with key and name.

**API:** `GET /api/v3/tags/`

**Caching:** Results cached with 10-minute TTL.

---

#### search_documents
Search Reader documents by query string. Uses cached list data.

**Parameters:**
- `query` (string, required): Search query
- `location` (string, optional): Limit to location
- `category` (string, optional): Limit to category
- `limit` (int, optional): Max results (default 50)

**Returns:** Matching documents with relevance ranking.

**Implementation:** Client-side search over cached document list.

---

### Profile: write (7 tools)

Tools in this profile require a corresponding read profile to be active.

#### save_document
Save a URL to Reader.

**Requires:** reader

**Parameters:**
- `url` (string, required): URL to save
- `title` (string, optional): Override title
- `author` (string, optional): Override author
- `summary` (string, optional): Custom summary
- `tags` ([]string, optional): Tags to apply
- `location` (string, optional): Save location (new, later, shortlist, archive)
- `category` (string, optional): Document category

**Returns:** Saved document with id and url.

**API:** `POST /api/v3/save/`

**Cache Invalidation:** Clears document list cache for this API key.

---

#### update_document
Update Reader document metadata.

**Requires:** reader

**Parameters:**
- `id` (string, required): Document ID
- `title` (string, optional): New title
- `author` (string, optional): New author
- `summary` (string, optional): New summary
- `location` (string, optional): New location
- `tags` ([]string, optional): Replace tags
- `seen` (bool, optional): Mark as seen/unseen

**Returns:** Updated document.

**API:** `PATCH /api/v3/update/{id}/`

**Cache Invalidation:** Clears document list cache for this API key.

---

#### create_highlight
Create a new highlight.

**Requires:** readwise

**Parameters:**
- `text` (string, required): Highlight text
- `source_id` (string, optional): Existing source ID
- `source_title` (string, optional): Title for new source (if source_id not provided)
- `source_author` (string, optional): Author for new source
- `source_url` (string, optional): URL for new source
- `note` (string, optional): Note to attach
- `location` (int, optional): Location in source
- `location_type` (string, optional): Location type (page, order, time_offset)
- `highlighted_at` (string, optional): ISO 8601 datetime

**Returns:** Created highlight with id.

**API:** `POST /api/v2/highlights/`

**Cache Invalidation:** Clears export cache for this API key.

---

#### update_highlight
Update an existing highlight.

**Requires:** readwise

**Parameters:**
- `id` (string, required): Highlight ID
- `text` (string, optional): New text
- `note` (string, optional): New note
- `location` (int, optional): New location
- `color` (string, optional): Highlight color

**Returns:** Updated highlight.

**API:** `PATCH /api/v2/highlights/{id}/`

**Cache Invalidation:** Clears export cache for this API key.

---

#### add_source_tag
Add a tag to a source.

**Requires:** readwise

**Parameters:**
- `source_id` (string, required): Source ID
- `name` (string, required): Tag name

**Returns:** Created tag with id.

**API:** `POST /api/v2/books/{id}/tags/`

---

#### add_highlight_tag
Add a tag to a highlight.

**Requires:** readwise

**Parameters:**
- `highlight_id` (string, required): Highlight ID
- `name` (string, required): Tag name

**Returns:** Created tag with id.

**API:** `POST /api/v2/highlights/{id}/tags/`

---

#### bulk_create_highlights
Create multiple highlights in a single request.

**Requires:** readwise

**Parameters:**
- `highlights` ([]object, required): Array of highlight objects, each with:
  - `text` (string, required): Highlight text
  - `source_title` (string, required): Source title
  - `source_author` (string, optional): Source author
  - `source_url` (string, optional): Source URL
  - `note` (string, optional): Note
  - `location` (int, optional): Location
  - `highlighted_at` (string, optional): ISO 8601 datetime

**Returns:** Array of created highlight IDs.

**API:** `POST /api/v2/highlights/` (batch format)

**Cache Invalidation:** Clears export cache for this API key.

---

### Profile: video (5 tools)

All video tools require the `reader` profile.

#### list_videos
List video documents in Reader.

**Requires:** reader

**Parameters:**
- `location` (string, optional): Filter by location
- `limit` (int, optional): Results limit (default 50)

**Returns:** List of video documents.

**API:** `GET /api/v3/list/?category=video`

---

#### get_video
Get video details including transcript if available.

**Requires:** reader

**Parameters:**
- `id` (string, required): Video document ID

**Returns:** Video details with transcript (if available).

**API:** `GET /api/v3/list/?id={id}`

---

#### get_video_position
Get playback position for a video.

**Requires:** reader

**Parameters:**
- `id` (string, required): Video document ID

**Returns:** Current playback position and duration.

**API:** `GET /api/v3/list/?id={id}` (extract from reading_progress)

---

#### update_video_position
Update playback position for a video.

**Requires:** reader, write

**Parameters:**
- `id` (string, required): Video document ID
- `position` (float, required): Position in seconds
- `duration` (float, optional): Total duration in seconds

**Returns:** Updated position.

**API:** `PATCH /api/v3/update/{id}/`

---

#### create_video_highlight
Create a highlight at a specific video timestamp.

**Requires:** reader, write

**Parameters:**
- `id` (string, required): Video document ID
- `text` (string, required): Highlight text (transcript excerpt)
- `timestamp` (float, required): Start timestamp in seconds
- `end_timestamp` (float, optional): End timestamp in seconds
- `note` (string, optional): Note to attach

**Returns:** Created highlight.

**API:** Custom implementation using highlight API with timestamp metadata.

---

### Profile: destructive (4 tools)

All destructive tools require corresponding read profiles.

#### delete_highlight
Delete a highlight.

**Requires:** readwise

**Parameters:**
- `id` (string, required): Highlight ID

**Returns:** Confirmation of deletion.

**API:** `DELETE /api/v2/highlights/{id}/`

**Cache Invalidation:** Clears export cache for this API key.

---

#### delete_highlight_tag
Remove a tag from a highlight.

**Requires:** readwise

**Parameters:**
- `highlight_id` (string, required): Highlight ID
- `tag_id` (string, required): Tag ID

**Returns:** Confirmation of deletion.

**API:** `DELETE /api/v2/highlights/{id}/tags/{tag_id}`

---

#### delete_source_tag
Remove a tag from a source.

**Requires:** readwise

**Parameters:**
- `source_id` (string, required): Source ID
- `tag_id` (string, required): Tag ID

**Returns:** Confirmation of deletion.

**API:** `DELETE /api/v2/books/{id}/tags/{tag_id}`

---

#### delete_document
Delete a Reader document.

**Requires:** reader

**Parameters:**
- `id` (string, required): Document ID

**Returns:** Confirmation of deletion.

**API:** `DELETE /api/v3/delete/{id}/`

**Cache Invalidation:** Clears document list cache for this API key.

---

## Caching

### Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Cache Manager                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              LRU Cache (per API key hash)            │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────────┐   │  │
│  │  │ highlights │ │ documents  │ │  reader_tags   │   │  │
│  │  │   export   │ │    list    │ │                │   │  │
│  │  │  TTL: 5m   │ │  TTL: 5m   │ │   TTL: 10m     │   │  │
│  │  └────────────┘ └────────────┘ └────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
│  Memory limit: configurable (default: 128 MB)              │
│  Eviction: LRU when limit exceeded                         │
└────────────────────────────────────────────────────────────┘
```

### Cache Keys

```
key = sha256(api_key_hash + endpoint + sorted_query_params)
```

API keys are hashed (SHA-256) before use in cache keys to prevent exposure.

### Cached Endpoints

| Endpoint | TTL | Invalidated By |
|----------|-----|----------------|
| `GET /api/v2/export/` | 5 min | create_highlight, update_highlight, delete_highlight, bulk_create_highlights |
| `GET /api/v2/books/` | 5 min | (none, source metadata rarely changes) |
| `GET /api/v3/list/` | 5 min | save_document, update_document, delete_document |
| `GET /api/v3/tags/` | 10 min | (none, tags change rarely) |

### Memory Management

- Each cache entry tracks its serialized size
- Total cache size is summed across all entries
- When limit exceeded, LRU entries are evicted until under limit
- Minimum entry lifetime: 30 seconds (prevent thrashing)

---

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

### Rate Limit Handling

429 responses from Readwise are passed through directly to the client. The `Retry-After` header value is included in the error response.

---

## Project Structure

```
readwise-mcp/
├── cmd/
│   └── readwise-mcp/
│       └── main.go                 # Entry point, config loading
├── internal/
│   ├── server/
│   │   ├── server.go               # MCP server setup
│   │   ├── transport.go            # SSE/HTTP transport
│   │   └── middleware.go           # Auth extraction, logging
│   ├── api/
│   │   ├── client.go               # HTTP client wrapper
│   │   ├── v2.go                   # Readwise v2 API methods
│   │   ├── v3.go                   # Reader v3 API methods
│   │   └── errors.go               # API error types
│   ├── cache/
│   │   ├── manager.go              # Cache orchestration
│   │   ├── lru.go                  # LRU implementation
│   │   └── entry.go                # Entry with size tracking
│   ├── tools/
│   │   ├── registry.go             # Tool registration
│   │   ├── profiles.go             # Profile definitions and resolution
│   │   ├── readwise.go             # Readwise profile tools
│   │   ├── reader.go               # Reader profile tools
│   │   ├── write.go                # Write profile tools
│   │   ├── video.go                # Video profile tools
│   │   ├── destructive.go          # Destructive profile tools
│   │   └── search.go               # Search implementations
│   └── types/
│       ├── source.go               # Source (book) types
│       ├── highlight.go            # Highlight types
│       ├── document.go             # Reader document types
│       └── common.go               # Shared types (pagination, errors)
├── deploy/
│   ├── kustomization.yml           # Kustomize config
│   ├── readwise-mcp.yml            # StatefulSet + Service
│   ├── vpa.yml                     # VerticalPodAutoscaler
│   └── nginx.conf                  # nginx sidecar config
├── docs/
│   ├── README.md                   # Project overview
│   ├── TOOLS.md                    # Tool reference
│   ├── PROFILES.md                 # Profile documentation
│   └── DEPLOYMENT.md               # Deployment guide
├── Containerfile                   # Podman build (multi-stage)
├── Makefile                        # Build, test, deploy targets
├── go.mod
└── go.sum
```

---

## Deployment

### Kubernetes Resources

**StatefulSet** with:
- 1 replica
- 2 containers: readwise-mcp (Go) + nginx (TLS termination)
- Resource requests: 128Mi memory, 100m CPU
- Resource limits: 256Mi memory, 500m CPU
- Liveness probe: HTTP GET /health
- Readiness probe: HTTP GET /ready

**Service** (LoadBalancer):
- Static IP from MetalLB range (10.9.11.XXX)
- Port 8443 (HTTPS via nginx)

**VPA** for resource recommendations.

### Container Build

```dockerfile
# Containerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o readwise-mcp ./cmd/readwise-mcp

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/readwise-mcp /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["readwise-mcp"]
```

### Build Command

```bash
podman build --platform linux/arm64 -t readwise-mcp:latest -f Containerfile .
```

---

## Testing Strategy

### Unit Tests

- All tools: parameter validation, response mapping
- API client: request building, response parsing, error handling
- Cache: LRU eviction, TTL expiry, size tracking
- Profile resolution: expansion, deduplication, dependency validation

### Integration Tests

- Mock Readwise API server
- End-to-end tool execution
- Cache behavior with real data shapes
- Profile combinations

### Test Structure

```
internal/
├── api/
│   ├── client_test.go
│   ├── v2_test.go
│   └── v3_test.go
├── cache/
│   ├── manager_test.go
│   └── lru_test.go
└── tools/
    ├── profiles_test.go
    ├── readwise_test.go
    ├── reader_test.go
    └── search_test.go
```

---

## Acceptance Criteria

1. **Profile System**
   - [ ] All base profiles (readwise, reader, write, video, destructive) work independently
   - [ ] Shortcut profiles (basic, all) expand correctly
   - [ ] Multiple profiles can be combined
   - [ ] Dependency validation works (write requires read profile)
   - [ ] Tool filtering based on active profiles works

2. **API Integration**
   - [ ] All Readwise v2 endpoints work correctly
   - [ ] All Reader v3 endpoints work correctly
   - [ ] API key passed through from client headers
   - [ ] 429 responses passed through with Retry-After

3. **Caching**
   - [ ] Export and list endpoints are cached
   - [ ] Cache respects configured memory limit
   - [ ] LRU eviction works correctly
   - [ ] Write operations invalidate relevant caches
   - [ ] Cache keys are isolated per API key

4. **Search**
   - [ ] search_highlights searches cached export data
   - [ ] search_documents searches cached document list
   - [ ] Results are relevance-ranked

5. **Deployment**
   - [ ] Container builds for linux/arm64 with Podman
   - [ ] Deploys to k3s via kustomize
   - [ ] nginx sidecar terminates TLS
   - [ ] Health and readiness probes work

6. **Documentation**
   - [ ] README covers setup and configuration
   - [ ] TOOLS.md documents all tools with examples
   - [ ] PROFILES.md explains profile system
   - [ ] DEPLOYMENT.md covers k3s deployment

---

## Open Questions

1. **MCP SDK choice**: Should we use `github.com/mark3labs/mcp-go` or another Go MCP implementation?
2. **TLS certificates**: Use Let's Encrypt via cert-manager, or self-signed for internal use?
3. **Metrics**: Should we expose Prometheus metrics for cache hits/misses, API latency?

---

## References

- [Readwise API Documentation](https://readwise.io/api_deets)
- [Reader API Documentation](https://readwise.io/reader_api)
- [Reference Implementation (TypeScript)](https://github.com/IAmAlexander/readwise-mcp)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
