# Research: Fix Search Root Cause and Highlights ID Mismatch

**Date**: 2026-05-17
**Branch**: 004-fix-search-root-cause

## Root Cause: search_documents Rate Limiting

### Investigation

The `search_documents` handler (`internal/tools/search.go:93-124`) calls `client.ListDocuments(ctx, apiKey, "", "", "", 0)` with limit=0, which fetches ALL documents from the Reader v3 API via cursor-based pagination (100 per page). Each page requires a separate HTTP request, rate-limited to 1 per 3 seconds by the client-side token bucket.

For a library with 500 documents, this means 5 paginated requests, taking at least 15 seconds. Combined with the 30-second HTTP timeout, large libraries will either hit the upstream rate limit or timeout.

The cache manager (`internal/cache/Manager`) exists and has TTL-configured endpoints, but search tools are not wired to use it. The `RegisterSearchDocumentsTool` and `RegisterSearchHighlightsTool` calls in `registry.go` pass only the `api.Client`, not the `cache.Manager`.

### Decision

Wire the cache manager into search tools. On cache hit, return cached data immediately. On cache miss, fetch from upstream, cache the result, then search locally. The cache already has TTLs for the relevant endpoints (`/api/v3/list/` = 5 min, `/api/v2/export/` = 5 min).

### Alternatives Considered

- **Server-side search API**: The Readwise/Reader API does not offer a search endpoint. Local search is the only option.
- **Incremental cache warming**: Background prefetch of documents. Over-engineered for this use case since the cache manager already handles TTL-based freshness.
- **Pass-through to list_documents cache**: Share cache with `list_documents` by using the same endpoint key. Chosen, since `list_documents` and `search_documents` fetch from the same `/api/v3/list/` endpoint.

## ULID vs Numeric ID Detection for list_highlights

### Investigation

The `list_highlights` handler (`internal/tools/readwise.go:145-169`) passes `input.SourceID` directly to `client.ListHighlights()`, which calls the v2 API at `/highlights/?book_id=<id>`. The v2 API only accepts numeric IDs for `book_id`.

Reader documents use ULID-style IDs (alphanumeric, 26 characters, e.g., `01krk43zgd57fcsj3tg7xn3skz`). These are deterministic, time-ordered identifiers that cannot be converted to the classic Readwise numeric book_id.

### Decision

Detect ULID-style IDs in `list_highlights` by checking if the string is purely numeric. If not numeric, use the export API to find highlights for the matching document, filtering by source. This leverages the existing `ExportHighlights` method and cache.

### Alternatives Considered

- **Separate tool (`list_document_highlights`)**: Creates tool sprawl, confusing for users. Rejected.
- **Reader API document-specific highlights endpoint**: The Reader v3 API doesn't expose a highlights-by-document endpoint. Not available.
- **Mapping ULID to numeric ID**: No known mapping exists between Reader ULIDs and classic Readwise book IDs. Rejected.
- **Filter export data by document ID**: Use `ExportHighlights` and filter by `user_book_id` matching. But `user_book_id` is numeric, not a ULID. The Reader document ID is a separate identifier.

Revised approach: For ULID source IDs, fetch the Reader document via `GetDocument` to get its metadata, then use the export API filtered or search by matching title/URL. However, the simplest approach is to just use the `/list/` endpoint with the document ID as a filter parameter, since the v3 API accepts the ULID.

Let me verify: the Reader v3 `/list/` endpoint returns documents. Each document has an `id` field that is the ULID. There is no v3 highlights-per-document endpoint. The best approach is to use the export API and match highlights by checking each source's `source_url` or other cross-referenceable field against the Reader document.

**Final decision**: For ULID source IDs, fetch the document via `GetDocument(id)` to get its metadata (title, URL), then use the already-cached export data to find highlights from that source by matching on title or URL. This avoids adding a new API call pattern and leverages the cache.

## Cache Wiring Architecture

### How search tools get the cache

The `RegisterSearchDocumentsTool` and `RegisterSearchHighlightsTool` functions need an additional `*cache.Manager` parameter. The registry already passes `cm` to write/video/destructive tools, so the pattern is established.

### Cache key strategy

- Document search uses endpoint `/api/v3/list/` with no params (fetches all), same key as `list_documents` with no filters.
- Highlight search uses endpoint `/api/v2/export/` with no params (fetches all), same key as `export_highlights`.
- This means data cached by `list_documents` is reusable by `search_documents`, and vice versa.

### Cache flow for search_documents

1. Check cache for `/api/v3/list/` with no params
2. If hit: unmarshal, search locally, return results
3. If miss: call `ListDocuments(ctx, apiKey, "", "", "", 0)`, cache raw bytes, search locally, return results

### Cache flow for search_highlights

1. Check cache for `/api/v2/export/` with no params
2. If hit: unmarshal, search locally, return results
3. If miss: call `ExportHighlights(ctx, apiKey, "")`, cache raw bytes, search locally, return results
