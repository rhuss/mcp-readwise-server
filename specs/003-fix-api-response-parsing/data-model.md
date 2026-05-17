# Data Model: Fix API Response Parsing Bugs

**Date**: 2026-05-17
**Branch**: 003-fix-api-response-parsing

## Modified Entities

### CursorResponse[T] (internal/types/common.go)

**Change**: `NextPageCursor` field type from `string` to `json.Number`

| Field | Before | After | Reason |
|-------|--------|-------|--------|
| NextPageCursor | `string` | `json.Number` | Readwise API returns numeric cursor values |

**Impact**: All callers that check `page.NextPageCursor == ""` for pagination termination continue to work because `json.Number` zero value is `""`.

Affected callers:
- `internal/api/v3.go:61` - `ListDocuments` pagination loop
- `internal/api/v2.go:140` - `ExportHighlights` pagination loop

### Client (internal/api/client.go)

**Change**: Add rate limiter and retry configuration fields

| Field | Type | Purpose |
|-------|------|---------|
| rateLimiter | `*rate.Limiter` | Token bucket for client-side rate limiting |
| maxRetries | `int` | Maximum retry attempts for 429 responses |

## New Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `golang.org/x/time/rate` | latest | Token bucket rate limiter |
