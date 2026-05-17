# Tasks: Fix API Response Parsing Bugs

**Input**: Design documents from `specs/003-fix-api-response-parsing/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Add new dependency and prepare infrastructure

- [x] T001 Add `golang.org/x/time` dependency via `go get golang.org/x/time/rate`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core changes to the API client that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Change `NextPageCursor` field type from `string` to `json.Number` in `CursorResponse[T]` struct in `internal/types/common.go`. Add `"encoding/json"` import. Update the `json` struct tag to remain `json:"nextPageCursor"`.
- [x] T003 Update `ListDocuments` pagination loop in `internal/api/v3.go:61` to compare `page.NextPageCursor` with `json.Number("")` (empty `json.Number`) instead of empty string for pagination termination. Use `page.NextPageCursor == ""` which works because `json.Number` zero value is `""`. Convert cursor to string with `page.NextPageCursor.String()` when setting the `pageCursor` query parameter.
- [x] T004 Update `ExportHighlights` pagination loop in `internal/api/v2.go:140` with the same cursor comparison fix as T003. Use `page.NextPageCursor.String()` when setting the `pageCursor` query parameter.
- [x] T005 Add `rateLimiter *rate.Limiter` field to the `Client` struct in `internal/api/client.go`. Initialize it in `NewClient()` and `NewClientWithBaseURLs()` with `rate.NewLimiter(rate.Every(3*time.Second), 1)` (20 requests/minute = 1 request per 3 seconds, burst 1). Add import for `golang.org/x/time/rate`.
- [x] T006 Add retry-with-backoff logic to `doRequest` in `internal/api/client.go`. Wrap the HTTP request execution in a retry loop (max 3 attempts). On 429 response: parse Retry-After header (default 60s if missing), sleep with exponential backoff plus random jitter (0-25%), then retry. Call `c.rateLimiter.Wait(ctx)` before each request attempt to enforce client-side throttling. After all retries exhausted, return `NewRateLimitError(retryAfter)`.

**Checkpoint**: Foundation ready, all API calls now have rate limiting and retry support, cursor parsing handles both string and numeric values

---

## Phase 3: User Story 1 - Search Documents Returns Correct Results (Priority: P1) MVP

**Goal**: `search_documents` returns matching documents instead of spurious rate-limit errors

**Independent Test**: Issue a `search_documents` call and verify matching documents are returned

### Tests for User Story 1

- [x] T007 [P] [US1] Add test `TestDoRequestRetryOn429` in `internal/api/client_test.go`: mock server returns 429 with Retry-After header on first call, 200 on second call. Verify retry succeeds and returns data.
- [x] T008 [P] [US1] Add test `TestDoRequestRetryExhausted` in `internal/api/client_test.go`: mock server always returns 429. Verify `RateLimitError` is returned after 3 attempts.
- [x] T009 [P] [US1] Add test `TestDoRequestRateLimiter` in `internal/api/client_test.go`: send multiple rapid requests, verify they are spaced by the rate limiter (measure elapsed time).
- [x] T010 [P] [US1] Add test `TestListDocumentsNumericCursor` in `internal/api/v3_test.go`: mock server returns `nextPageCursor` as a JSON number (e.g., `12345`), verify pagination works without errors.

### Implementation for User Story 1

- [x] T011 [US1] Verify `search_documents` handler in `internal/tools/search.go` requires no changes (bug fix is in lower layers via T005, T006). Run existing tests to confirm no regressions: `go test ./internal/tools/ -run TestSearch`.

**Checkpoint**: search_documents should now return results instead of rate-limit errors

---

## Phase 4: User Story 2 - Search Highlights Returns Parsed Results (Priority: P2)

**Goal**: `search_highlights` returns matching highlights without JSON parsing errors

**Independent Test**: Issue a `search_highlights` call and verify matching highlights are returned

### Tests for User Story 2

- [x] T012 [P] [US2] Add test `TestExportHighlightsNumericCursor` in `internal/api/v2_test.go`: mock server returns `nextPageCursor` as a JSON number. Verify `ExportHighlights` parses correctly and paginates.
- [x] T013 [P] [US2] Add test `TestExportHighlightsNullCursor` in `internal/api/v2_test.go`: mock server returns `nextPageCursor` as null. Verify pagination stops correctly.

### Implementation for User Story 2

- [x] T014 [US2] Verify `search_highlights` handler in `internal/tools/search.go` requires no changes (bug fix is in T002, T004). Run: `go test ./internal/tools/ -run TestSearchHighlights`.

**Checkpoint**: search_highlights should now return results without JSON parsing errors

---

## Phase 5: User Story 3 - Resilient Rate Limit Handling (Priority: P3)

**Goal**: Server automatically retries rate-limited requests with backoff

**Independent Test**: Simulate burst requests and verify automatic retry behavior

### Tests for User Story 3

- [x] T015 [P] [US3] Add test `TestDoRequestRetryWithJitter` in `internal/api/client_test.go`: verify retry delays include jitter (non-deterministic but within expected range).
- [x] T016 [P] [US3] Add test `TestDoRequestRetryAfterHeaderMissing` in `internal/api/client_test.go`: mock server returns 429 without Retry-After header. Verify default 60s backoff is used.

### Implementation for User Story 3

- [x] T017 [US3] Implementation is complete via T005 and T006. Run full test suite to verify: `go test ./...`.

**Checkpoint**: Rate limiting and retry fully functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, constitution update, and cleanup

- [x] T018 Update constitution in `.specify/memory/constitution.md` to amend Principle VI: add note that client-side retry for 429 responses and client-side rate limiting are permitted.
- [x] T019 Run full test suite `go test ./...` and verify all tests pass with zero failures.
- [x] T020 Verify build succeeds: `go build ./cmd/readwise-mcp/`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion, BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Phase 2 completion
  - US1, US2, US3 can proceed in parallel after Phase 2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on T002, T003, T005, T006. No dependency on other stories.
- **User Story 2 (P2)**: Depends on T002, T004. No dependency on other stories.
- **User Story 3 (P3)**: Depends on T005, T006. No dependency on other stories.

### Within Each User Story

- Tests written first (verify they compile)
- Implementation verified via test execution
- Story checkpoint validates independently

### Parallel Opportunities

- T007, T008, T009, T010 can all run in parallel (different test functions, same file)
- T012, T013 can run in parallel
- T015, T016 can run in parallel
- After Phase 2: US1, US2, US3 can proceed in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for US1 together:
Task: "Add test TestDoRequestRetryOn429 in internal/api/client_test.go"
Task: "Add test TestDoRequestRetryExhausted in internal/api/client_test.go"
Task: "Add test TestDoRequestRateLimiter in internal/api/client_test.go"
Task: "Add test TestListDocumentsNumericCursor in internal/api/v3_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (add dependency)
2. Complete Phase 2: Foundational (cursor fix + rate limiter + retry)
3. Complete Phase 3: User Story 1 tests + verification
4. **STOP and VALIDATE**: Test search_documents independently
5. Deploy if ready

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
2. Add User Story 1 -> search_documents works -> Deploy (MVP!)
3. Add User Story 2 -> search_highlights works -> Deploy
4. Add User Story 3 -> Retry resilience validated -> Deploy
5. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- The foundational phase does the heavy lifting; user story phases mostly verify correctness
- Total tasks: 20
- US1: 5 tasks (T007-T011), US2: 3 tasks (T012-T014), US3: 3 tasks (T015-T017)
- Parallel opportunities: T007-T010, T012-T013, T015-T016, and all 3 user stories after Phase 2
