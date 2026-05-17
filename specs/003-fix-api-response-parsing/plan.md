# Implementation Plan: Fix API Response Parsing Bugs

**Branch**: `003-fix-api-response-parsing` | **Date**: 2026-05-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/003-fix-api-response-parsing/spec.md`

## Summary

Fix two API response parsing bugs (spurious rate-limit error in `search_documents`, JSON type mismatch in `search_highlights`) and add client-side rate limiting with retry-and-backoff for genuine 429 responses. The `CursorResponse.NextPageCursor` field changes from `string` to `json.Number`, and the `Client` struct gains a token bucket rate limiter and retry logic in `doRequest`.

## Technical Context

**Language/Version**: Go 1.25.7
**Primary Dependencies**: `github.com/modelcontextprotocol/go-sdk` v1.3.0, `golang.org/x/time/rate` (new)
**Storage**: N/A (stateless server)
**Testing**: Go stdlib `testing` with `net/http/httptest` mock servers
**Target Platform**: Linux arm64 (k3s home cluster)
**Project Type**: Single Go module
**Performance Goals**: All API calls complete within 30s timeout (existing), rate limiter configured at 20 req/min
**Constraints**: Must not break existing tool behavior, must be backward compatible with both string and numeric cursor values
**Scale/Scope**: Single MCP server binary, ~15 source files affected

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Stateless MCP Server | PASS | No state changes; rate limiter is per-process, not persisted |
| II. Profile-Based Tool Exposure | PASS | No tool surface changes |
| III. Go Implementation | PASS | Pure Go, no new external services |
| IV. Test-First Development | PASS | Tests written for all changes |
| V. Caching for Performance | PASS | Cache layer unchanged |
| VI. Transparent Error Handling | AMENDMENT | Adding retry/rate-limiting contradicts "no server-side retry." Constitution amendment needed |

**Violation justification**: Principle VI's "no retry" rule causes user-facing failures on every burst usage. The amendment permits client-side retry for 429 responses only, with transparent error surfacing after retries are exhausted.

## Project Structure

### Documentation (this feature)

```text
specs/003-fix-api-response-parsing/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/readwise-mcp/
└── main.go              # No changes expected

internal/
├── api/
│   ├── client.go        # Add rate limiter, retry logic in doRequest
│   ├── client_test.go   # Tests for retry, rate limiting, error handling
│   ├── errors.go        # No changes
│   ├── v2.go            # Update NextPageCursor usage (string -> json.Number)
│   ├── v2_test.go       # Add test for numeric cursor in export response
│   ├── v3.go            # Update NextPageCursor usage (string -> json.Number)
│   └── v3_test.go       # Add test for numeric cursor in list response
├── types/
│   └── common.go        # Change NextPageCursor type to json.Number
└── tools/
    ├── search.go         # No changes (bug fix is in lower layers)
    └── search_test.go    # Add integration-style test for search with mock server
```

**Structure Decision**: Standard Go project layout, all changes within existing `internal/` packages. No new packages needed.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Retry logic in doRequest | Users hit 429 on every search due to pagination burst | No retry = broken search for any library with 100+ documents |
| New dependency (x/time/rate) | Token bucket rate limiting | Hand-rolled rate limiter would be error-prone and less tested |
