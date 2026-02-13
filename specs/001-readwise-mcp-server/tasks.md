# Tasks: Readwise MCP Server

**Input**: Design documents from `/specs/001-readwise-mcp-server/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per constitution principle IV (Test-First Development). TDD is mandatory.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and Go module setup

- [ ] T001 Initialize Go module with `go mod init github.com/rhuss/readwise-mcp-server` and add dependency `github.com/modelcontextprotocol/go-sdk@v1.3.0` in go.mod
- [ ] T002 Create project directory structure per plan.md: cmd/readwise-mcp/, internal/server/, internal/api/, internal/cache/, internal/tools/, internal/types/, deploy/
- [ ] T003 Create Makefile with targets: build, test, lint, clean, container, deploy

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [ ] T004 [P] Define Config struct and environment variable loading (READWISE_PROFILES, PORT, LOG_LEVEL, CACHE_MAX_SIZE_MB, CACHE_TTL_SECONDS, CACHE_ENABLED) in internal/types/common.go
- [ ] T005 [P] Define ErrorResponse struct with Type, Code, Message, RetryAfter fields and constructor functions for validation_error, auth_error, api_error, internal_error in internal/api/errors.go
- [ ] T006 [P] Define Source struct with JSON tags mapping to Readwise v2 API fields (user_book_id, title, author, category, etc.) in internal/types/source.go
- [ ] T007 [P] Define Highlight struct with JSON tags mapping to v2 API fields (id, text, note, location, book_id, etc.) in internal/types/highlight.go
- [ ] T008 [P] Define Document struct with JSON tags mapping to Reader v3 API fields (id, url, title, category, location, reading_progress, etc.) in internal/types/document.go
- [ ] T009 [P] Define Tag struct (ID int64, Name string) and pagination types (PageResponse, CursorResponse) in internal/types/tag.go and internal/types/common.go
- [ ] T010 Implement HTTP client wrapper with Authorization header injection (Token format), base URL configuration for v2/v3, and response error handling (429 passthrough with Retry-After) in internal/api/client.go
- [ ] T011 [P] Write tests for Config loading (defaults, env overrides, profile parsing) in internal/types/common_test.go
- [ ] T012 [P] Write tests for ErrorResponse construction and JSON serialization in internal/api/errors_test.go
- [ ] T013 [P] Write tests for HTTP client (auth header injection, 429 handling, error parsing) using httptest.NewServer in internal/api/client_test.go
- [ ] T014 Implement API key extraction middleware that reads the Readwise API key from MCP client request headers and injects into context in internal/server/middleware.go
- [ ] T015 Write tests for API key extraction middleware (present, missing, invalid cases) in internal/server/middleware_test.go
- [ ] T016 Implement MCP server skeleton using official go-sdk: create mcp.Server, configure Streamable HTTP transport on configurable port, wire up middleware in internal/server/server.go
- [ ] T017 Implement entry point that loads config, creates server, and starts listening in cmd/readwise-mcp/main.go

**Checkpoint**: Foundation ready. Server starts, accepts MCP connections, extracts API keys. No tools registered yet.

---

## Phase 3: User Story 1 - Browse Readwise Highlights (Priority: P1) MVP

**Goal**: AI assistant users can browse Readwise highlights and sources via 9 MCP tools

**Independent Test**: Connect to server with valid API key, list sources, list highlights, get daily review, list tags

### Tests for User Story 1

> **Write tests FIRST, ensure they FAIL before implementation**

- [ ] T018 [P] [US1] Write tests for Readwise v2 API methods (ListBooks, GetBook, ListHighlights, GetHighlight, ExportHighlights, GetDailyReview, ListBookTags, ListHighlightTags) using httptest mock server in internal/api/v2_test.go
- [ ] T019 [P] [US1] Write tests for list_sources, get_source, list_highlights, get_highlight tool handlers (parameter validation, response mapping) in internal/tools/readwise_test.go
- [ ] T020 [P] [US1] Write tests for search_highlights (query matching, relevance scoring, result limiting, source_id filtering) in internal/tools/search_test.go

### Implementation for User Story 1

- [ ] T021 [P] [US1] Implement Readwise v2 API methods: ListBooks (page-number pagination), GetBook, ListHighlights (page-number pagination), GetHighlight in internal/api/v2.go
- [ ] T022 [P] [US1] Implement Readwise v2 API methods: ExportHighlights (cursor pagination, loop until nextPageCursor null), GetDailyReview, ListBookTags, ListHighlightTags in internal/api/v2.go
- [ ] T023 [US1] Implement tool handlers for list_sources (page_size, page, category, updated_after params), get_source (id param), list_highlights (page_size, page, source_id, updated_after params), get_highlight (id param) in internal/tools/readwise.go
- [ ] T024 [US1] Implement tool handlers for export_highlights (updated_after param), get_daily_review (no params), list_source_tags (source_id param), list_highlight_tags (highlight_id param) in internal/tools/readwise.go
- [ ] T025 [US1] Implement search_highlights tool handler: fetch/retrieve cached export data, case-insensitive substring search across text/notes/source titles, relevance scoring (exact > partial), configurable limit (default 50, max 200) in internal/tools/search.go
- [ ] T026 [US1] Register all 9 readwise profile tools with the MCP server using go-sdk AddTool with parameter schemas in internal/tools/registry.go

**Checkpoint**: User Story 1 complete. 9 readwise tools functional. Can list sources, highlights, tags, export, daily review, and search.

---

## Phase 4: User Story 2 - Browse Reader Documents (Priority: P1)

**Goal**: AI assistant users can browse Reader documents, filter by location/category, and search

**Independent Test**: Connect with valid API key, list documents filtered by location, get document with content, search documents

### Tests for User Story 2

- [ ] T027 [P] [US2] Write tests for Reader v3 API methods (ListDocuments with cursor pagination, GetDocument with/without content, ListReaderTags) using httptest mock server in internal/api/v3_test.go
- [ ] T028 [P] [US2] Write tests for list_documents, get_document, list_reader_tags tool handlers (parameter validation, filtering, response mapping) in internal/tools/reader_test.go
- [ ] T029 [P] [US2] Write tests for search_documents (query matching across title/author/summary/notes, location/category filtering, relevance scoring) in internal/tools/search_test.go

### Implementation for User Story 2

- [ ] T030 [P] [US2] Implement Reader v3 API methods: ListDocuments (cursor pagination, location/category/updatedAfter/limit params), GetDocument (id param, withHtmlContent option) in internal/api/v3.go
- [ ] T031 [P] [US2] Implement Reader v3 API methods: ListReaderTags in internal/api/v3.go
- [ ] T032 [US2] Implement tool handlers for list_documents (location, category, updated_after, limit params), get_document (id, include_content params), list_reader_tags (no params) in internal/tools/reader.go
- [ ] T033 [US2] Implement search_documents tool handler: fetch/retrieve cached document list, apply location/category filters, case-insensitive search across title/author/summary/notes, relevance scoring, configurable limit (default 50, max 200) in internal/tools/search.go
- [ ] T034 [US2] Register all 4 reader profile tools with the MCP server in internal/tools/registry.go

**Checkpoint**: User Stories 1 and 2 complete. All 13 read-only tools functional.

---

## Phase 5: User Story 4 - Profile-Based Tool Configuration (Priority: P2)

**Goal**: Server operators can control which tools are exposed via configurable profiles

**Independent Test**: Start server with different READWISE_PROFILES values, verify correct tools are registered

### Tests for User Story 4

- [ ] T035 [P] [US4] Write tests for profile definitions (5 base profiles with correct tool counts and types) in internal/tools/profiles_test.go
- [ ] T036 [P] [US4] Write tests for shortcut expansion ("basic" -> reader+write, "all" -> all 5 profiles) in internal/tools/profiles_test.go
- [ ] T037 [P] [US4] Write tests for dependency validation (write without read fails, video without reader fails, destructive without read fails, basic self-resolves) in internal/tools/profiles_test.go
- [ ] T038 [P] [US4] Write tests for tool filtering (only tools from active profiles registered, no duplicates across profiles) in internal/tools/profiles_test.go

### Implementation for User Story 4

- [ ] T039 [US4] Define Profile struct (Name, Type, Dependencies, Tools), base profile definitions (readwise:read, reader:read, write:modifier, video:modifier, destructive:modifier), and shortcut map in internal/tools/profiles.go
- [ ] T040 [US4] Implement profile resolution: parse READWISE_PROFILES, expand shortcuts, deduplicate, validate dependencies (fail fast with clear error), collect matching tools in internal/tools/profiles.go
- [ ] T041 [US4] Refactor tool registration in internal/tools/registry.go to use profile system: only register tools from active profiles, support video tool sub-dependencies (3 read-only need reader, 2 write-requiring need reader+write)
- [ ] T042 [US4] Wire profile resolution into server startup in cmd/readwise-mcp/main.go: load profiles from config, resolve, register filtered tools

**Checkpoint**: Profile system complete. Server exposes only tools matching active profiles. Dependency validation prevents unsafe configurations.

---

## Phase 6: User Story 7 - Efficient Caching (Priority: P2)

**Goal**: Server-side LRU cache reduces API calls, search operates over cached data, write operations invalidate caches

**Independent Test**: Make repeated read calls, verify reduced API calls. Make write call, verify cache invalidation. Exceed memory limit, verify LRU eviction.

### Tests for User Story 7

- [ ] T043 [P] [US7] Write tests for CacheEntry (TTL expiry, size tracking, last access update) in internal/cache/entry_test.go
- [ ] T044 [P] [US7] Write tests for LRU cache (insert, get, evict LRU on size limit, min 30s lifetime, concurrent access safety) in internal/cache/lru_test.go
- [ ] T045 [P] [US7] Write tests for CacheManager (key construction with SHA-256 hashed API key, per-user isolation, endpoint-specific TTLs, invalidation map) in internal/cache/manager_test.go

### Implementation for User Story 7

- [ ] T046 [P] [US7] Implement CacheEntry struct with Key, Data, Size, CreatedAt, TTL, LastAccessedAt fields and IsExpired/Touch methods in internal/cache/entry.go
- [ ] T047 [US7] Implement LRU cache: doubly-linked list + map, Get (update access time, check TTL), Put (add/update entry, evict if over size limit), size tracking, min 30s entry lifetime, sync.RWMutex for concurrency in internal/cache/lru.go
- [ ] T048 [US7] Implement CacheManager: key construction (SHA256 of api_key_hash + endpoint + sorted query params), Get/Put with endpoint-specific TTLs (export 5m, books 5m, list 5m, tags 10m), Invalidate by endpoint+user, total size reporting in internal/cache/manager.go
- [ ] T049 [US7] Integrate cache into API client: wrap cacheable endpoints (export, books list, document list, tags) with cache-through, invalidate on write/destructive operations per invalidation map in data-model.md in internal/api/client.go
- [ ] T050 [US7] Integrate cache into search tools: search_highlights reads from export cache, search_documents reads from document list cache, trigger cache population on miss in internal/tools/search.go
- [ ] T051 [US7] Wire CacheManager into server startup with config (enabled flag, max size MB, default TTL) in cmd/readwise-mcp/main.go

**Checkpoint**: Cache system complete. Repeated reads return cached data. Search operates over cache. Write invalidates relevant entries. Memory bounded by LRU eviction.

---

## Phase 7: User Story 3 - Create and Modify Content (Priority: P2)

**Goal**: AI assistant users can save documents, create highlights, and manage tags via 7 write tools

**Independent Test**: Save a URL to Reader, create a highlight, add a tag. Verify via subsequent read operations.

### Tests for User Story 3

- [ ] T052 [P] [US3] Write tests for v2 write API methods (CreateHighlight, UpdateHighlight, CreateHighlights batch, AddBookTag, AddHighlightTag) using httptest mock in internal/api/v2_test.go
- [ ] T053 [P] [US3] Write tests for v3 write API methods (SaveDocument, UpdateDocument) using httptest mock in internal/api/v3_test.go
- [ ] T054 [P] [US3] Write tests for write tool handlers (parameter validation, cache invalidation triggers, response mapping) in internal/tools/write_test.go

### Implementation for User Story 3

- [ ] T055 [P] [US3] Implement v2 write API methods: CreateHighlight (POST /highlights/), UpdateHighlight (PATCH /highlights/{id}/), CreateHighlights batch (POST /highlights/ with array), AddBookTag (POST /books/{id}/tags/), AddHighlightTag (POST /highlights/{id}/tags/) in internal/api/v2.go
- [ ] T056 [P] [US3] Implement v3 write API methods: SaveDocument (POST /save/), UpdateDocument (PATCH /update/{id}/) in internal/api/v3.go
- [ ] T057 [US3] Implement tool handlers for save_document, update_document, create_highlight, update_highlight, add_source_tag, add_highlight_tag, bulk_create_highlights with parameter validation and cache invalidation calls in internal/tools/write.go
- [ ] T058 [US3] Register all 7 write profile tools (gated by write profile, requires readwise or reader dependency) in internal/tools/registry.go

**Checkpoint**: Write tools complete. Users can save, create, update content. Cache invalidation keeps search results fresh.

---

## Phase 8: User Story 5 - Video Document Management (Priority: P3)

**Goal**: AI assistant users can browse videos, view transcripts, manage playback position, and create timestamped highlights

**Independent Test**: List videos, get video with transcript, update playback position, create timestamped highlight

### Tests for User Story 5

- [ ] T059 [P] [US5] Write tests for video tool handlers: list_videos (category=video filter), get_video (transcript content), get_video_position (read-only), update_video_position (requires write), create_video_highlight (timestamp metadata, requires write) in internal/tools/video_test.go

### Implementation for User Story 5

- [ ] T060 [US5] Implement tool handlers: list_videos (list_documents filtered to category=video), get_video (get_document with content), get_video_position (reading_progress extraction) as read-only requiring reader profile in internal/tools/video.go
- [ ] T061 [US5] Implement tool handlers: update_video_position (PATCH via update_document with reading_progress), create_video_highlight (create_highlight with timestamp in location/location_type=time_offset) requiring reader+write profiles in internal/tools/video.go
- [ ] T062 [US5] Register 5 video profile tools: 3 read-only gated by video+reader, 2 write-requiring gated by video+reader+write in internal/tools/registry.go

**Checkpoint**: Video tools complete. Read-only video tools work with reader profile only. Write video tools require write profile.

---

## Phase 9: User Story 6 - Destructive Operations (Priority: P3)

**Goal**: AI assistant users can delete highlights, tags, and documents with explicit opt-in via destructive profile

**Independent Test**: Enable destructive profile, delete a highlight, verify it no longer appears. Disable destructive profile, verify delete tools unavailable.

### Tests for User Story 6

- [ ] T063 [P] [US6] Write tests for v2 destructive API methods (DeleteHighlight, DeleteHighlightTag, DeleteBookTag) and v3 (DeleteDocument) using httptest mock in internal/api/v2_test.go and internal/api/v3_test.go
- [ ] T064 [P] [US6] Write tests for destructive tool handlers (parameter validation, cache invalidation triggers, profile gating) in internal/tools/destructive_test.go

### Implementation for User Story 6

- [ ] T065 [P] [US6] Implement v2 destructive API methods: DeleteHighlight (DELETE /highlights/{id}/), DeleteHighlightTag (DELETE /highlights/{id}/tags/{tag_id}), DeleteBookTag (DELETE /books/{id}/tags/{tag_id}) in internal/api/v2.go
- [ ] T066 [P] [US6] Implement v3 destructive API method: DeleteDocument (DELETE /delete/{id}/) in internal/api/v3.go
- [ ] T067 [US6] Implement tool handlers for delete_highlight, delete_highlight_tag, delete_source_tag, delete_document with parameter validation and cache invalidation in internal/tools/destructive.go
- [ ] T068 [US6] Register 4 destructive profile tools (gated by destructive profile, requires corresponding read profile) in internal/tools/registry.go

**Checkpoint**: All 29 MCP tools complete. Destructive tools only available with explicit destructive profile opt-in.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Health endpoints, deployment, and final validation

- [ ] T069 Implement /health endpoint (returns 200 if server is running) and /ready endpoint (returns 200 if profiles resolved and server accepting connections) in internal/server/server.go
- [ ] T070 Write tests for /health and /ready endpoints in internal/server/server_test.go
- [ ] T071 [P] Create Containerfile with multi-stage build: Go 1.22 builder stage, alpine runtime, CGO_ENABLED=0 GOOS=linux GOARCH=arm64, expose port 8080 in Containerfile
- [ ] T072 [P] Create Kustomize deployment manifests: StatefulSet (1 replica, readwise-mcp + nginx sidecar, resource limits, liveness/readiness probes), Service (LoadBalancer), VPA in deploy/readwise-mcp.yml, deploy/kustomization.yml, deploy/vpa.yml
- [ ] T073 [P] Create nginx sidecar config for TLS termination (listen 8443, proxy_pass to localhost:8080) in deploy/nginx.conf
- [ ] T074 Run full test suite (`go test ./...`), verify all tests pass, validate SC-007 (100% FR coverage)
- [ ] T075 Validate quickstart.md steps: build, run, configure profiles, run tests

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion, BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2
- **US2 (Phase 4)**: Depends on Phase 2, can run in parallel with US1
- **US4 (Phase 5)**: Depends on Phase 2 (refactors registry from Phase 3/4)
- **US7 (Phase 6)**: Depends on Phase 2, enhances tools from Phase 3/4
- **US3 (Phase 7)**: Depends on Phases 2, 5 (profiles), 6 (cache invalidation)
- **US5 (Phase 8)**: Depends on Phases 2, 4 (reader API), 5 (profiles), 7 (write API for 2 tools)
- **US6 (Phase 9)**: Depends on Phases 2, 5 (profiles), 6 (cache invalidation)
- **Polish (Phase 10)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Independent after foundational. MVP deliverable.
- **US2 (P1)**: Independent after foundational. Can parallel with US1.
- **US4 (P2)**: Independent after foundational. Refactors tool registration.
- **US7 (P2)**: Independent after foundational. Enhances existing tools with caching.
- **US3 (P2)**: Depends on US4 (profiles gate write access) and US7 (cache invalidation).
- **US5 (P3)**: Depends on US2 (reader API), US4 (profiles), US3 (write API for 2 tools).
- **US6 (P3)**: Depends on US4 (profiles) and US7 (cache invalidation).

### Parallel Opportunities

**Phase 2** (all [P] tasks): T004-T009 can run in parallel (different files). T011-T013 and T015 can run in parallel (test files).

**Phase 3** US1: T018-T020 tests in parallel, then T021-T022 API methods in parallel.

**Phase 4** US2: T027-T029 tests in parallel, then T030-T031 API methods in parallel.

**Phase 5** US4: T035-T038 all tests in parallel.

**Phase 6** US7: T043-T045 tests in parallel, then T046 can parallel with T047.

**Phase 7** US3: T052-T054 tests in parallel, then T055-T056 API methods in parallel.

**Phase 9** US6: T063-T064 tests in parallel, then T065-T066 API methods in parallel.

---

## Parallel Example: User Story 1

```text
# Launch all tests in parallel:
Task T018: "Tests for v2 API methods in internal/api/v2_test.go"
Task T019: "Tests for readwise tool handlers in internal/tools/readwise_test.go"
Task T020: "Tests for search_highlights in internal/tools/search_test.go"

# After tests written and failing, launch API methods in parallel:
Task T021: "v2 API list/get methods in internal/api/v2.go"
Task T022: "v2 API export/review/tags methods in internal/api/v2.go"

# Then implement tools sequentially (same file):
Task T023: "list_sources, get_source, list_highlights, get_highlight handlers"
Task T024: "export_highlights, daily_review, tag handlers"
Task T025: "search_highlights handler"
Task T026: "Register 9 readwise tools"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 (9 readwise tools)
4. **STOP and VALIDATE**: Connect MCP client, list sources, search highlights
5. Deploy if ready (functional with readwise profile only)

### Incremental Delivery

1. Setup + Foundational -> Server starts, accepts connections
2. Add US1 (readwise) -> 9 tools, browsing + search -> Deploy MVP
3. Add US2 (reader) -> +4 tools, document browsing -> Deploy
4. Add US4 (profiles) -> Tool gating, shortcuts -> Deploy
5. Add US7 (cache) -> Performance, search enhancement -> Deploy
6. Add US3 (write) -> +7 tools, content creation -> Deploy
7. Add US5 (video) -> +5 tools, video management -> Deploy
8. Add US6 (destructive) -> +4 tools, deletion -> Deploy
9. Polish -> Health, container, k3s deployment -> Production

---

## Notes

- [P] tasks = different files, no dependencies between them
- [Story] label maps task to specific user story
- Constitution mandates TDD: write tests FIRST, verify they FAIL, then implement
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Total: 29 MCP tools across 5 profiles, matching SC-001
