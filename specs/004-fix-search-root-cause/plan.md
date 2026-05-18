# Implementation Plan: Fix Search Root Cause and Highlights ID Mismatch

**Branch**: `004-fix-search-root-cause` | **Date**: 2026-05-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/004-fix-search-root-cause/spec.md`

## Summary

Fix the root cause of `search_documents` rate limiting by wiring search tools through the existing cache manager, and fix the `list_highlights` ULID ID mismatch by detecting non-numeric source IDs and routing through the export API instead of the classic v2 highlights endpoint.

## Technical Context

**Language/Version**: Go 1.25.7
**Primary Dependencies**: `github.com/modelcontextprotocol/go-sdk` v1.3.0, `golang.org/x/time/rate`
**Storage**: In-memory LRU cache (existing `internal/cache` package)
**Testing**: Go stdlib `testing` with `net/http/httptest` mock servers
**Target Platform**: Linux arm64 (k3s home cluster)
**Project Type**: Single Go module
**Performance Goals**: Cached searches complete in under 1 second
**Constraints**: Must not break existing tool behavior; cache keys must be compatible with existing tools
**Scale/Scope**: 4 files modified, ~150 lines changed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Stateless MCP Server | PASS | Cache is per-process, not persisted |
| II. Profile-Based Tool Exposure | PASS | No tool surface changes |
| III. Go Implementation | PASS | Pure Go |
| IV. Test-First Development | PASS | Tests for cache integration and ULID routing |
| V. Caching for Performance | PASS | Directly implements this principle for search |
| VI. Transparent Error Handling | PASS | Amended in 003 to permit retry/rate-limiting |

No violations.

## Project Structure

### Documentation (this feature)

```text
specs/004-fix-search-root-cause/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── cache/
│   └── manager.go       # No changes (existing infrastructure sufficient)
├── tools/
│   ├── registry.go      # Pass cache manager to search + readwise registration
│   ├── search.go        # Add cache Get/Put in search handlers
│   ├── search_test.go   # Tests for cached search and ULID routing
│   └── readwise.go      # ULID detection in list_highlights, pass cache to registration
└── types/
    └── common.go        # No changes
```

**Structure Decision**: All changes within existing `internal/tools/` package. The cache manager is already instantiated in `server.go` and passed to `RegisterAllTools` in `registry.go`.

## Complexity Tracking

No constitution violations. No complexity concerns.
