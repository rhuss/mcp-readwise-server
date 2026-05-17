# Research: Fix API Response Parsing Bugs

**Date**: 2026-05-17
**Branch**: 003-fix-api-response-parsing

## Bug 1: search_documents Spurious Rate Limit Error

### Investigation

The `search_documents` tool handler in `internal/tools/search.go:93-124` calls `client.ListDocuments()` which fetches ALL documents from the Reader v3 `/list/` endpoint with cursor-based pagination (`internal/api/v3.go:14-71`). It then performs local in-memory search/scoring.

The `doRequest` method in `internal/api/client.go:78` checks for HTTP 429 status before any other error handling. If the upstream API returns 429 during any paginated fetch, the error propagates up as a `RateLimitError`.

**Root cause hypothesis**: When a user has a large library, `ListDocuments` makes multiple paginated requests (100 documents per page). Each page fetch hits the Readwise API, and with a 20 req/min rate limit, a library with 300+ documents would require 3+ requests in rapid succession. Combined with any other recent API calls, this easily triggers genuine 429 responses.

**Alternative hypothesis**: The v3 `/list/` endpoint returns a non-200 status code that is being misinterpreted. However, the `doRequest` logic correctly checks status codes in order: 429, 401/403, 204, then generic non-2xx. A misinterpretation seems unlikely unless the search endpoint uses a different URL pattern.

### Decision

The fix requires two complementary changes:
1. Add retry-with-backoff in `doRequest` for 429 responses instead of failing immediately
2. Add client-side rate limiting (token bucket) to prevent hitting the upstream limit

### Alternatives Considered

- **API-side search**: The Readwise Reader API does not expose a server-side search endpoint. The current approach of fetching all documents and filtering locally is the only option.
- **Cache-first search**: The cache layer exists but isn't wired into search. A cache-warm approach would reduce API calls, but is a separate enhancement.

## Bug 2: search_highlights JSON Parsing Error

### Investigation

The `ExportHighlights` method in `internal/api/v2.go:108-150` uses `CursorResponse[ExportSource]`. The `CursorResponse` struct in `internal/types/common.go:142-146` defines `NextPageCursor` as `string`. The Readwise v2 export API returns `nextPageCursor` as an integer (e.g., `1234567`) not a quoted string.

When `json.Unmarshal` encounters a JSON number where a Go `string` is expected, it fails with: `json: cannot unmarshal number into Go struct field CursorResponse[...].nextPageCursor of type string`.

### Decision

Change `NextPageCursor` from `string` to `json.Number`. The `json.Number` type handles both quoted strings and bare numbers in JSON. Downstream code that compares `NextPageCursor` to empty string for pagination termination will use `page.NextPageCursor == ""` which works with `json.Number` since its zero value is `""`.

### Alternatives Considered

- **`interface{}`**: Too permissive, requires type assertions everywhere.
- **Custom unmarshaler**: Correct but over-engineered for a single field.
- **`json.Number`**: Simplest solution, handles both string and number, zero value is empty string. Chosen.

## Rate Limiting Strategy

### Client-Side Rate Limiting

- **Algorithm**: Token bucket using Go's `golang.org/x/time/rate` package
- **Configuration**: 20 tokens/minute (matching upstream limit), burst of 1 (strict enforcement)
- **Scope**: Per-Client instance (single rate limiter since all requests go through `doRequest`)
- **Behavior**: `Wait()` before each request, blocking until a token is available

### Retry with Backoff

- **Trigger**: HTTP 429 response from upstream
- **Strategy**: Exponential backoff with jitter, respecting Retry-After header when present
- **Max retries**: 3
- **Default backoff**: 60 seconds when Retry-After header is missing
- **Jitter**: Random 0-25% added to prevent thundering herd

### Constitution Impact

Constitution Principle VI states "No server-side retry or rate limiting." This feature deliberately amends that principle. The constitution should be updated to reflect that client-side rate limiting and retry logic for upstream 429 responses is now permitted.
