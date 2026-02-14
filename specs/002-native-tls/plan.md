# Implementation Plan: Native TLS Support

**Branch**: `002-native-tls` | **Date**: 2026-02-14 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-native-tls/spec.md`

## Summary

Add native HTTPS support to the readwise-mcp-server, removing the nginx sidecar dependency. When TLS is configured via environment variables, the server runs two listeners: HTTPS on `TLS_PORT` for MCP traffic and HTTP on `PORT` for health/readiness probes. When TLS is not configured, behavior is unchanged (single HTTP listener). The Kubernetes deployment is simplified to a single-container Pod with the TLS Secret mounted directly.

## Technical Context

**Language/Version**: Go 1.25+ (current: go 1.25.7)
**Primary Dependencies**: `github.com/modelcontextprotocol/go-sdk` v1.3.0, Go stdlib `crypto/tls`, `net/http`
**Storage**: N/A
**Testing**: `go test` with `net/http/httptest`, `crypto/tls` test certificates
**Target Platform**: Linux/arm64 (k3s cluster), also darwin/amd64 for development
**Project Type**: Single Go binary
**Performance Goals**: Server startup < 2 seconds in both TLS and non-TLS modes
**Constraints**: No new external dependencies (stdlib only for TLS)
**Scale/Scope**: 4 source files modified, 2 new test files, 3 deployment files modified, 1 deployment file deleted

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Stateless MCP Server | PASS | TLS config via env vars, no state stored |
| II. Profile-Based Tool Exposure | N/A | No impact on tool profiles |
| III. Go Implementation | PASS | Using Go stdlib `crypto/tls`, standard project layout |
| IV. Test-First Development | PASS | Unit tests for config validation, TLS listener, cert expiry warning |
| V. Caching for Performance | N/A | No impact on caching |
| VI. Transparent Error Handling | PASS | Descriptive errors for missing/invalid certs, port conflicts |
| Technology Stack: TLS | AMENDMENT | Changes "nginx sidecar for TLS termination" to "Native Go TLS (nginx sidecar optional)" |

**Gate result**: PASS. The TLS amendment is documented in the spec and will be applied to the constitution after implementation.

## Project Structure

### Documentation (this feature)

```text
specs/002-native-tls/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/readwise-mcp/
└── main.go                          # Signal handling for graceful shutdown

internal/
├── types/
│   ├── common.go                    # Add TLS config fields + validation
│   └── common_test.go               # Tests for TLS config loading + validation
└── server/
    ├── server.go                    # Dual-listener logic, TLS setup, cert expiry check
    └── server_test.go               # Tests for TLS listener, health-only mux

deploy/
├── readwise-mcp.yml                 # Remove nginx sidecar, mount TLS secret to Go container
├── kustomization.yml                # Remove nginx ConfigMap generator
└── nginx.conf                       # DELETE this file

Containerfile                        # Add EXPOSE 8443
```

**Structure Decision**: Existing Go project layout (`cmd/` + `internal/`) is unchanged. All modifications are within existing packages. No new packages needed.

## Complexity Tracking

No constitution violations requiring justification. The TLS amendment is a planned evolution documented in the spec.
