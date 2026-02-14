# Tasks: Native TLS Support

**Input**: Design documents from `/specs/002-native-tls/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests are included per constitution Principle IV (Test-First Development).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: No new project setup needed. This feature modifies existing files only.

- [ ] T001 Generate self-signed test certificates for use in unit/integration tests in `internal/server/testdata/`

---

## Phase 2: Foundational (TLS Configuration)

**Purpose**: Add TLS config fields and validation to `types.Config`. MUST complete before user story implementation.

- [ ] T002 [P] Add TLS fields (`TLSCertFile`, `TLSKeyFile`, `TLSPort`) to `Config` struct and `LoadConfig()` in `internal/types/common.go`. Add `TLSEnabled() bool`, `TLSAddr() string` methods. Add `ValidateTLS() error` method that checks: both or neither cert/key set, file existence, port conflict with `PORT`.
- [ ] T003 [P] Add unit tests for TLS config loading and validation in `internal/types/common_test.go`: test defaults (TLS disabled), test env var overrides (`TLS_CERT_FILE`, `TLS_KEY_FILE`, `TLS_PORT`), test partial config error (only cert or only key), test port conflict error, test `TLSEnabled()`, test `TLSAddr()`.

**Checkpoint**: TLS configuration is loadable and validated. No listener changes yet.

---

## Phase 3: User Story 1 - Operator enables TLS (Priority: P1) MVP

**Goal**: Server serves MCP traffic over HTTPS when TLS is configured, with health probes on plain HTTP.

**Independent Test**: Start server with TLS env vars, verify `/mcp` responds over HTTPS and `/health` responds over HTTP.

### Tests for User Story 1

- [ ] T004 [P] [US1] Write test for TLS listener startup in `internal/server/server_test.go`: create a server with TLS config pointing to test certs, verify HTTPS listener starts on `TLS_PORT` and responds to requests, verify HTTP health listener starts on `PORT` and responds to `/health` and `/ready`.
- [ ] T005 [P] [US1] Write test for certificate expiry warning in `internal/server/server_test.go`: create a certificate expiring in < 30 days, verify the server logs a warning at startup. Create a certificate expiring in > 30 days, verify no warning.
- [ ] T006 [P] [US1] Write test for non-TLS backward compatibility in `internal/server/server_test.go`: create a server without TLS config, verify single HTTP listener serves all endpoints (`/mcp`, `/health`, `/ready`).

### Implementation for User Story 1

- [ ] T007 [US1] Implement dual-listener logic in `internal/server/server.go`: split `ListenAndServe()` into two modes. When `TLSEnabled()`: create HTTPS `http.Server` with `tls.Config{MinVersion: tls.VersionTLS12}` serving `/mcp` on `TLSAddr()`, create HTTP `http.Server` serving only `/health` and `/ready` on `Addr()`. Start both in goroutines. When not `TLSEnabled()`: keep current single-listener behavior unchanged.
- [ ] T008 [US1] Add certificate expiry check in `internal/server/server.go`: after loading cert with `tls.LoadX509KeyPair`, parse leaf certificate with `x509.ParseCertificate`, check `NotAfter` against `time.Now().Add(30 * 24 * time.Hour)`, log warning if expiring soon.
- [ ] T009 [US1] Add graceful shutdown with signal handling in `cmd/readwise-mcp/main.go`: set up `os.Signal` handler for `SIGTERM` and `SIGINT`, pass `context.Context` to server, call `Shutdown(ctx)` on both HTTP servers with 10-second timeout. Update `Server.ListenAndServe()` signature to accept `context.Context`.
- [ ] T010 [US1] Add TLS validation call in `cmd/readwise-mcp/main.go`: call `cfg.ValidateTLS()` after `LoadConfig()`, exit with error if validation fails.

**Checkpoint**: Server can serve HTTPS with dual listeners. Backward compatibility preserved.

---

## Phase 4: User Story 2 - Backward Compatibility (Priority: P1)

**Goal**: Server with no TLS configuration behaves identically to current implementation.

**Independent Test**: Start server without TLS env vars, verify all endpoints respond over HTTP on single port.

This story is already covered by T006 (backward compatibility test) and the conditional logic in T007. No additional tasks needed, as the implementation preserves the existing code path when TLS is disabled.

**Checkpoint**: Backward compatibility verified through T006.

---

## Phase 5: User Story 3 - Kubernetes Deployment (Priority: P2)

**Goal**: Deployment manifest runs a single container per Pod with TLS, removing nginx sidecar.

**Independent Test**: Apply manifests, verify single-container Pod serves HTTPS on 8443.

### Implementation for User Story 3

- [ ] T011 [US3] Update `deploy/readwise-mcp.yml`: remove nginx sidecar container (lines 56-73), remove nginx-config volume (line 74-76), keep tls volume but mount it into Go container at `/etc/tls` (read-only). Add env vars `TLS_CERT_FILE=/etc/tls/tls.crt`, `TLS_KEY_FILE=/etc/tls/tls.key`, `TLS_PORT=8443`. Add `containerPort: 8443` named `https` to Go container. Keep health/readiness probes pointing to `http` port unchanged.
- [ ] T012 [US3] Update `deploy/kustomization.yml`: remove the `configMapGenerator` section for `readwise-mcp-nginx`.
- [ ] T013 [US3] Delete `deploy/nginx.conf` (no longer needed).
- [ ] T014 [US3] Update `Containerfile`: add `EXPOSE 8443` after existing `EXPOSE 8080`.

**Checkpoint**: Kubernetes deployment runs single-container Pod with native TLS.

---

## Phase 6: Polish and Cross-Cutting Concerns

**Purpose**: Constitution update and final validation.

- [ ] T015 Update `specs/constitution.md` Technology Stack section: change "TLS: nginx sidecar for TLS termination" to "TLS: Native Go TLS (nginx sidecar optional)".
- [ ] T016 Run all tests with `go test ./...` and verify passing.
- [ ] T017 Validate quickstart.md scenarios manually: test both TLS and non-TLS startup locally.

---

## Dependencies and Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on T001 (test certs for T003)
- **User Story 1 (Phase 3)**: Depends on Phase 2 completion
- **User Story 2 (Phase 4)**: Covered by Phase 3 implementation (no separate work)
- **User Story 3 (Phase 5)**: Can start after Phase 2 (independent of Phase 3 code changes)
- **Polish (Phase 6)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires foundational config (Phase 2). Core implementation.
- **User Story 2 (P1)**: No separate tasks. Verified by T006 within US1 implementation.
- **User Story 3 (P2)**: Requires Phase 2 for env var names. Can be done in parallel with US1 (different files: deploy/ vs internal/).

### Within Each User Story

- Tests written and failing before implementation
- Config changes before server changes
- Server changes before main.go changes

### Parallel Opportunities

- T002 and T003 can run in parallel (different aspects of same files, but T003 depends on T002's types)
- T004, T005, T006 can all run in parallel (independent test functions)
- T011, T012, T013, T014 can all run in parallel (different files in deploy/ and root)

---

## Parallel Example: User Story 1

```bash
# Launch all tests in parallel (T004, T005, T006):
Task: "Write TLS listener test in internal/server/server_test.go"
Task: "Write cert expiry warning test in internal/server/server_test.go"
Task: "Write backward compat test in internal/server/server_test.go"
```

## Parallel Example: User Story 3

```bash
# Launch all deploy changes in parallel (T011, T012, T013, T014):
Task: "Update deploy/readwise-mcp.yml"
Task: "Update deploy/kustomization.yml"
Task: "Delete deploy/nginx.conf"
Task: "Update Containerfile"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Generate test certs
2. Complete Phase 2: TLS config fields + validation + tests
3. Complete Phase 3: Dual-listener implementation + tests
4. **STOP and VALIDATE**: Test TLS locally with self-signed certs
5. Working TLS without any deployment changes

### Incremental Delivery

1. Setup + Foundational -> Config ready
2. User Story 1 -> TLS works locally -> Test independently (MVP!)
3. User Story 3 -> Deployment updated -> Single-container Pod
4. Polish -> Constitution updated, all tests pass

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently


<!-- SDD-TRAIT:beads -->
## Beads Task Management

This project uses beads (`bd`) for persistent task tracking across sessions:
- `bd ready --json` returns unblocked tasks (dependencies resolved)
- `bd done <id>` marks a task complete
- `bd sync` persists state to git
- `bd create --title "DISCOVERED: ..." --labels discovered` tracks new work
- `bd list --status open` shows remaining work
Tasks from this file are bootstrapped as beads issues via `{Skill: sdd:beads-execute}`.
