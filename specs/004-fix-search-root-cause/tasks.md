# Tasks: Fix Search Root Cause and Highlights ID Mismatch

**Input**: Design documents from `specs/004-fix-search-root-cause/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No new dependencies needed. This phase wires the cache manager into the tool registration.

- [x] T001 Update `RegisterSearchDocumentsTool` signature in `internal/tools/search.go` to accept `*cache.Manager` as a third parameter. Update `makeSearchDocumentsHandler` to accept and close over the cache manager.
- [x] T002 Update `RegisterSearchHighlightsTool` signature in `internal/tools/search.go` to accept `*cache.Manager` as a third parameter. Update `makeSearchHighlightsHandler` to accept and close over the cache manager.
- [x] T003 Update `RegisterReadwiseTools` signature in `internal/tools/readwise.go` to accept `*cache.Manager` as a second parameter. Pass it through to handler factories that need it (specifically `makeListHighlightsHandler`).
- [x] T004 Update call sites in `internal/tools/registry.go` to pass `cm` to `RegisterSearchDocumentsTool(s, client, cm)`, `RegisterSearchHighlightsTool(s, client, cm)`, and `RegisterReadwiseTools(s, client, cm)`.
- [x] T005 Verify the project still compiles after the signature changes: `go build ./...`.

**Checkpoint**: All search and readwise tools now have access to the cache manager. No behavior change yet.

---

## Phase 2: User Story 1 - Search Documents Uses Cache (Priority: P1) MVP

**Goal**: `search_documents` uses cached document data when available, avoiding upstream API calls on every search

**Independent Test**: Call `search_documents` twice. First call fetches from upstream. Second call should use cache (no upstream request).

### Tests for User Story 1

- [x] T006 [P] [US1] Add test `TestSearchDocumentsCacheHit` in `internal/tools/search_test.go`: create a mock server that counts requests, a cache manager, call search handler twice with the same query. Assert the mock server received only one set of paginated requests (cache hit on second call).
- [x] T007 [P] [US1] Add test `TestSearchDocumentsCacheMiss` in `internal/tools/search_test.go`: create a mock server, a cache manager with no data, call search handler once. Assert the mock server was called and results are correct.

### Implementation for User Story 1

- [x] T008 [US1] In `makeSearchDocumentsHandler` in `internal/tools/search.go`: before calling `client.ListDocuments`, check `cm.Get(apiKey, "/api/v3/list/", nil)`. If hit, unmarshal the cached bytes into `types.CursorResponse[types.Document]` and use `Results` for local search. If miss, call `client.ListDocuments` as before, then marshal the full result and call `cm.Put(apiKey, "/api/v3/list/", nil, data)` to cache it.
- [x] T009 [US1] Run tests to verify: `go test ./internal/tools/ -run TestSearchDocuments -v`.

**Checkpoint**: search_documents uses cache. Repeated searches are instant.

---

## Phase 3: User Story 2 - List Highlights for Reader Documents (Priority: P2)

**Goal**: `list_highlights` works with ULID-style Reader document IDs

**Independent Test**: Call `list_highlights` with a ULID source_id and verify highlights are returned.

### Tests for User Story 2

- [x] T010 [P] [US2] Add test `TestListHighlightsWithULID` in `internal/tools/readwise_test.go` (create file if needed): mock both v2 export and v3 list endpoints. Call `list_highlights` with a ULID source_id. Assert highlights matching the document are returned.
- [x] T011 [P] [US2] Add test `TestListHighlightsWithNumericID` in `internal/tools/readwise_test.go`: call `list_highlights` with a numeric source_id. Assert it still calls the v2 `/highlights/` endpoint (backward compatible).

### Implementation for User Story 2

- [x] T012 [US2] Add a helper function `isNumericID(id string) bool` in `internal/tools/readwise.go` that returns true if the string consists entirely of digits.
- [x] T013 [US2] In `makeListHighlightsHandler` in `internal/tools/readwise.go`: when `input.SourceID` is non-empty and `isNumericID` returns false (ULID detected), use a different code path: (1) call `client.GetDocument(ctx, apiKey, input.SourceID, false)` to get the document metadata, (2) call `client.ExportHighlights(ctx, apiKey, "")` to get all highlights (using cache via the cache manager if available), (3) filter export sources by matching `SourceURL` or `Title` against the retrieved document, (4) return the matching highlights. If numeric, use the existing v2 API path.
- [x] T014 [US2] Run tests to verify: `go test ./internal/tools/ -run TestListHighlights -v`.

**Checkpoint**: list_highlights works with both numeric and ULID source IDs.

---

## Phase 4: User Story 3 - Search Highlights Uses Cache (Priority: P3)

**Goal**: `search_highlights` uses cached export data when available

**Independent Test**: Call `search_highlights` twice. Second call should use cache.

### Tests for User Story 3

- [x] T015 [P] [US3] Add test `TestSearchHighlightsCacheHit` in `internal/tools/search_test.go`: same pattern as T006 but for highlight search. Assert mock server receives only one set of requests.

### Implementation for User Story 3

- [x] T016 [US3] In `makeSearchHighlightsHandler` in `internal/tools/search.go`: before calling `client.ExportHighlights`, check `cm.Get(apiKey, "/api/v2/export/", nil)`. If hit, unmarshal cached bytes into `types.CursorResponse[types.ExportSource]` and use `Results` for local search. If miss, call `client.ExportHighlights`, cache the result with `cm.Put(apiKey, "/api/v2/export/", nil, data)`.
- [x] T017 [US3] Run tests to verify: `go test ./internal/tools/ -run TestSearchHighlights -v`.

**Checkpoint**: search_highlights uses cache. Repeated searches are instant.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and build verification

- [x] T018 Run full test suite `go test ./...` and verify all tests pass.
- [x] T019 Verify build succeeds: `go build ./cmd/readwise-mcp/`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately. BLOCKS all user stories.
- **User Stories (Phase 2-4)**: All depend on Phase 1 completion.
  - US1, US2, US3 can proceed in parallel after Phase 1.
- **Polish (Phase 5)**: Depends on all user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Depends on T001, T002, T004. No dependency on other stories.
- **User Story 2 (P2)**: Depends on T003, T004. No dependency on other stories.
- **User Story 3 (P3)**: Depends on T002, T004. No dependency on other stories.

### Parallel Opportunities

- T001, T002, T003 modify the same file or closely related files, so run sequentially
- T006, T007 can run in parallel (different test functions)
- T010, T011 can run in parallel (different test functions)
- After Phase 1: US1, US2, US3 can proceed in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (wire cache manager)
2. Complete Phase 2: User Story 1 (cached document search)
3. **STOP and VALIDATE**: Test search_documents with warm/cold cache
4. Deploy if ready

### Incremental Delivery

1. Setup -> Foundation ready
2. Add User Story 1 -> search_documents uses cache -> Deploy (MVP!)
3. Add User Story 2 -> list_highlights works with ULIDs -> Deploy
4. Add User Story 3 -> search_highlights uses cache -> Deploy

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Total tasks: 19
- US1: 4 tasks (T006-T009), US2: 5 tasks (T010-T014), US3: 3 tasks (T015-T017)
- The cache endpoint keys (`/api/v3/list/`, `/api/v2/export/`) match the keys already configured in `internal/cache/manager.go` defaultTTLs
