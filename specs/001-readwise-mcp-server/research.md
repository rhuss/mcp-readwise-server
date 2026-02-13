# Research: Readwise MCP Server

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13

## Decision 1: MCP SDK Selection

**Decision**: Use `github.com/modelcontextprotocol/go-sdk` (official Go MCP SDK)

**Rationale**:
- Official SDK maintained by the Go team at Google and Anthropic jointly
- Stable at v1.3.0 with backward compatibility guarantee since v1.0.0
- Reflection-based tool registration reduces boilerplate for 29 tools
- Built-in support for Streamable HTTP (recommended) and SSE (legacy)
- Future-proof: tracks spec changes as the reference implementation
- Industry trend is toward the official SDK (GitHub's MCP server migrating from mcp-go)

**Alternatives considered**:
- `github.com/mark3labs/mcp-go` (v0.29.0): Most popular community SDK with ~1,307 importers. Builder-pattern API is more ergonomic for prototyping. However, pre-1.0 with breaking changes between minor versions (v0.29.0 changed argument access patterns). Explicit HTTP header passthrough in tool handlers is a plus, but the official SDK supports this via transport APIs.
- `github.com/metoro-io/mcp-golang`: Struct-based auto-schema generation. Smallest community (~84 importers), less transport flexibility, slower development pace.

## Decision 2: Transport Protocol

**Decision**: Use Streamable HTTP as the primary transport

**Rationale**:
- SSE transport was deprecated in MCP specification version 2025-03-26
- Streamable HTTP uses a single endpoint, works well with existing HTTP infrastructure (proxies, load balancers, nginx sidecar)
- Does not require persistent connections, reducing resource usage
- The official SDK supports it natively via built-in transport handlers
- Compatible with the k3s + nginx sidecar deployment model

**Alternatives considered**:
- SSE: Still supported for backward compatibility but deprecated. Requires two separate endpoints and persistent connections.
- Stdio: Not applicable for a network-accessible server deployment.

## Decision 3: Readwise API Authentication

**Decision**: Extract API key from MCP client request headers and pass as `Authorization: Token XXX` to Readwise APIs

**Rationale**:
- Both Readwise v2 and Reader v3 use the same token format: `Authorization: Token XXX`
- Single token works across both API surfaces
- Token validation available via `GET /api/v2/auth/` (returns 204 on success)
- No OAuth flow or other authentication mechanisms exist

**Alternatives considered**: None viable. The Readwise API only supports token-based auth.

## Decision 4: Pagination Strategy

**Decision**: Support both page-number and cursor-based pagination depending on the upstream API endpoint

**Rationale**:
- Readwise v2 list endpoints (`/books/`, `/highlights/`) use page-number pagination (`page`, `page_size`)
- Readwise v2 export endpoint (`/export/`) uses cursor-based pagination (`pageCursor`, `nextPageCursor`)
- Reader v3 list endpoint (`/list/`) uses cursor-based pagination exclusively
- The MCP tools must abstract this difference for clients

**Alternatives considered**: Normalizing to a single pagination style would add complexity and break alignment with upstream API behavior.

## Decision 5: Rate Limit Handling

**Decision**: Pass through 429 responses with `Retry-After` header value

**Rationale**:
- Readwise v2 default: 240 req/min, restricted endpoints (`/books/`, `/highlights/`): 20 req/min
- Reader v3 default: 20 req/min, save/update: 50 req/min
- Both APIs return HTTP 429 with `Retry-After` header (seconds to wait)
- Constitution mandates transparent error handling with no server-side retry or rate limiting

**Alternatives considered**: Server-side retry with backoff. Rejected per constitution principle VI (Transparent Error Handling).

## Decision 6: API Key Passthrough Mechanism

**Decision**: MCP clients pass the Readwise API key via a custom header in the HTTP request to the MCP server. The server extracts it from the request context per-call.

**Rationale**:
- The official Go SDK provides access to transport-level request context
- API key must be available per-request since different clients may use different keys
- No server-side credential storage per constitution principle I
- The MCP server hashes the API key (SHA-256) for use in cache keys per FR-018

**Alternatives considered**:
- Environment variable: Would limit the server to a single API key, preventing multi-user support.
- MCP resource/prompt parameters: Non-standard and would require client-side changes.

## Decision 7: Search Implementation

**Decision**: Client-side (server-side from MCP perspective) search over cached data using case-insensitive substring matching with relevance scoring

**Rationale**:
- Readwise API does not provide a search endpoint
- Export and list endpoints return all data, suitable for local search
- Cached data enables fast repeated searches without API calls
- Relevance ranking: exact matches scored higher than partial matches, with matches across text, notes, and source titles
- Result limit: configurable, default 50, maximum 200 per FR-016/FR-017

**Alternatives considered**:
- Full-text search engine (Bleve, etc.): Over-engineering for the scale (up to 10,000 highlights per SC-003). Simple substring matching with scoring is sufficient.
- Fuzzy matching: Could be added later but not required by spec.

## API Reference Summary

### Readwise v2 Endpoints Used

| Endpoint | Method | Rate Limit | Pagination |
| -------- | ------ | ---------- | ---------- |
| `/api/v2/auth/` | GET | 240/min | N/A |
| `/api/v2/books/` | GET | 20/min | Page-number |
| `/api/v2/books/{id}/` | GET | 240/min | N/A |
| `/api/v2/books/{id}/tags` | GET | 240/min | N/A |
| `/api/v2/books/{id}/tags/` | POST | 240/min | N/A |
| `/api/v2/books/{id}/tags/{tag_id}` | DELETE | 240/min | N/A |
| `/api/v2/highlights/` | GET | 20/min | Page-number |
| `/api/v2/highlights/` | POST | 240/min | N/A |
| `/api/v2/highlights/{id}/` | GET | 240/min | N/A |
| `/api/v2/highlights/{id}/` | PATCH | 240/min | N/A |
| `/api/v2/highlights/{id}/` | DELETE | 240/min | N/A |
| `/api/v2/highlights/{id}/tags` | GET | 240/min | N/A |
| `/api/v2/highlights/{id}/tags/` | POST | 240/min | N/A |
| `/api/v2/highlights/{id}/tags/{tag_id}` | DELETE | 240/min | N/A |
| `/api/v2/export/` | GET | 240/min | Cursor |
| `/api/v2/review/` | GET | 240/min | N/A |

### Reader v3 Endpoints Used

| Endpoint | Method | Rate Limit | Pagination |
| -------- | ------ | ---------- | ---------- |
| `/api/v3/list/` | GET | 20/min | Cursor |
| `/api/v3/save/` | POST | 50/min | N/A |
| `/api/v3/update/{id}/` | PATCH | 50/min | N/A |
| `/api/v3/delete/{id}/` | DELETE | 20/min | N/A |
| `/api/v3/tags/` | GET | 20/min | N/A |

### Auth Header Format

```
Authorization: Token <access_token>
```

Single token works for both v2 and v3. Validate via `GET /api/v2/auth/` (returns HTTP 204).
