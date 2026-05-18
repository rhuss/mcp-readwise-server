# Data Model: Fix Search Root Cause and Highlights ID Mismatch

**Date**: 2026-05-17
**Branch**: 004-fix-search-root-cause

## Modified Entities

### Search Tool Registration (internal/tools/registry.go)

**Change**: Pass `*cache.Manager` to search tool registration functions.

| Function | Before | After |
|----------|--------|-------|
| `RegisterSearchDocumentsTool` | `(s, client)` | `(s, client, cm)` |
| `RegisterSearchHighlightsTool` | `(s, client)` | `(s, client, cm)` |

### Search Tool Handlers (internal/tools/search.go)

**Change**: Add cache lookup/store in search handlers.

| Handler | Before | After |
|---------|--------|-------|
| `makeSearchDocumentsHandler` | Fetches all docs from API every call | Checks cache first, fetches on miss, caches result |
| `makeSearchHighlightsHandler` | Fetches all exports from API every call | Checks cache first, fetches on miss, caches result |

### List Highlights Handler (internal/tools/readwise.go)

**Change**: Detect ULID vs numeric source IDs, route accordingly.

| Scenario | Before | After |
|----------|--------|-------|
| Numeric source_id | Calls v2 `/highlights/?book_id=<id>` | Same (unchanged) |
| ULID source_id | Calls v2 `/highlights/?book_id=<id>` (fails with 400) | Uses export data + document lookup to find matching highlights |

## Cache Key Mapping

| Tool | Cache Endpoint Key | Params | Shared With |
|------|--------------------|--------|-------------|
| search_documents | `/api/v3/list/` | (none) | list_documents (no filters) |
| search_highlights | `/api/v2/export/` | (none) | export_highlights (no filters) |

## No New Dependencies

All changes use existing packages. No new imports required beyond the existing `cache` package.
