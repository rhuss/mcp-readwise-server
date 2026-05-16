# Readwise MCP Server: API Response Parsing Bugs

*Observed: 2026-05-16*
*Server: readwise-mcp-server (Go), deployed on k3s home cluster*
*Source: /Users/rhuss/Development/ai/mcp/readwise-mcp-server*

## Bug 1: `search_documents` returns fake "rate_limited" error

### Symptoms

- Every call to `search_documents` returns `api_error: rate_limited: Rate limited by upstream API. Retry after N seconds` regardless of actual API usage
- Other endpoints (`list_documents`, `get_document`) work fine at the same time
- Server logs (`kubectl -n mcp logs readwise-mcp-0`) show no corresponding request reaching the server when `search_documents` is called. Only session connect/disconnect events visible
- Waiting the suggested retry period (50-60 seconds) makes no difference
- The error persists even when it's the first and only request in a session

### Hypothesis

The Reader search API (`/reader/v3/list/?query=...` or similar) may return a different response format or status code than what the Go client expects. The server might be misinterpreting a non-429 response as a rate limit error. Alternatively, the Readwise Reader search endpoint itself might have a separate, stricter rate limit than the list endpoint, but the "retry after 50+ seconds" on a first-ever request argues against this.

### Where to look

- `internal/api/client.go`: how HTTP responses are parsed, specifically error handling for non-200 responses
- `internal/api/errors.go`: the `NewRateLimitError` function and where it's called
- Check if the Reader search API uses a different base URL or endpoint that might be returning an unexpected response
- Check if the `Retry-After` header parsing is correct

### Reproduction

```
# Any search_documents call via MCP triggers this:
mcp__readwise__search_documents(query="anything", limit=5)
# Returns: api_error: rate_limited: Rate limited by upstream API. Retry after 54 seconds
```

---

## Bug 2: `search_highlights` JSON parsing error

### Symptoms

- Calling `search_highlights` returns: `internal_error: internal: failed to parse export response: json: cannot unmarshal number into Go struct field CursorResponse[...ExportSource].nextPageCursor of type string`
- The Readwise API returns `nextPageCursor` as a number (integer), but the Go struct defines it as `string`

### Root cause

Type mismatch in the response struct. The Readwise export API changed (or always used) a numeric cursor, but the Go struct expects a string.

### Where to look

- `internal/types/` or wherever `CursorResponse` and `ExportSource` are defined
- The `nextPageCursor` field needs to accept both string and number (use `json.Number` or `interface{}` and convert, or use a custom unmarshaler)

### Fix

Change the `nextPageCursor` field type. Options:

1. **Quick fix**: Change field type to `interface{}` and convert to string after unmarshaling
2. **Proper fix**: Use a custom JSON unmarshaler that handles both string and number
3. **Simplest**: Change to `json.Number` which accepts both

```go
// Before:
type CursorResponse[T any] struct {
    NextPageCursor string `json:"nextPageCursor"`
    // ...
}

// After (option 3):
type CursorResponse[T any] struct {
    NextPageCursor json.Number `json:"nextPageCursor"`
    // ...
}
```

### Reproduction

```
# Any search_highlights call triggers this:
mcp__readwise__search_highlights(query="test", limit=1)
# Returns: internal_error: internal: failed to parse export response: json: cannot unmarshal number...
```

---

## Priority

Bug 1 is higher priority since `search_documents` is the primary way to find articles by title/content. Bug 2 affects highlight search which is less commonly used but still broken.

## Additional observation: rate limiting in general

The Readwise API has a 20 requests/minute rate limit. The server has `CACHE_ENABLED=true` configured but no client-side request queuing or throttling. When genuine rate limits are hit (after a burst of requests), the server surfaces the 429 directly to the MCP client with no retry logic. Consider adding:

- Client-side rate limiting (token bucket, max N requests per minute)
- Automatic retry with backoff for 429 responses
- Cache TTL tuning (`CACHE_TTL_SECONDS` env var, currently using default)
