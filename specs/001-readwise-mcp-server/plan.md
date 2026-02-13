# Implementation Plan: Readwise MCP Server

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-readwise-mcp-server/spec.md`

## Summary

Build a Go MCP server that exposes 29 tools for interacting with the Readwise v2 and Reader v3 APIs. The server uses the official Go MCP SDK with Streamable HTTP transport, implements a profile-based tool registration system with dependency validation, and includes an LRU cache with per-user isolation and memory-bounded eviction. API keys are passed through from MCP clients without server-side storage.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `github.com/modelcontextprotocol/go-sdk` (v1.3.0, official MCP SDK)
**Storage**: In-memory LRU cache only (no persistent storage)
**Testing**: `go test` with `testing` stdlib, `net/http/httptest` for mock API server
**Target Platform**: Linux/arm64 (k3s cluster), containerized via Podman
**Project Type**: Single server binary
**Performance Goals**: Search results within 1 second for 10,000 highlights (SC-003), 10 concurrent clients (SC-008)
**Constraints**: Cache memory bounded at configurable limit (default 128 MB), startup under 5 seconds (SC-005)
**Scale/Scope**: Single-instance deployment, up to 10 concurrent MCP client connections

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
| --------- | ------ | ----- |
| I. Stateless MCP Server | PASS | API key passthrough via request headers (FR-009), no server-side credential storage |
| II. Profile-Based Tool Exposure | PASS | 5 profiles with dependency validation (FR-006 through FR-008) |
| III. Go Implementation | PASS | Go 1.22+, standard `cmd/` and `internal/` layout, Podman + Kustomize |
| IV. Test-First Development | PASS | TDD with `go test`, mock API server for integration tests (SC-007) |
| V. Caching for Performance | PASS | LRU cache with per-API-key isolation, memory limits, TTL (FR-010 through FR-012, FR-018) |
| VI. Transparent Error Handling | PASS | 429 passthrough with Retry-After (FR-013), structured errors (FR-014) |

**Gate result**: PASS. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-readwise-mcp-server/
├── plan.md              # This file
├── research.md          # Phase 0 output - SDK and API research
├── data-model.md        # Phase 1 output - entity definitions
├── quickstart.md        # Phase 1 output - getting started guide
├── contracts/           # Phase 1 output - API contracts
│   ├── readwise-v2.md   # Readwise v2 API contract
│   ├── reader-v3.md     # Reader v3 API contract
│   └── mcp-tools.md     # MCP tool definitions contract
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── readwise-mcp/
    └── main.go                 # Entry point, config loading, server startup

internal/
├── server/
│   ├── server.go               # MCP server setup, tool registration
│   └── middleware.go           # API key extraction from request context
├── api/
│   ├── client.go               # HTTP client wrapper with auth header injection
│   ├── v2.go                   # Readwise v2 API methods
│   ├── v3.go                   # Reader v3 API methods
│   └── errors.go               # API error types and 429 passthrough
├── cache/
│   ├── manager.go              # Cache orchestration, invalidation logic
│   ├── lru.go                  # LRU implementation with size tracking
│   └── entry.go                # Cache entry with TTL and size
├── tools/
│   ├── registry.go             # Tool registration with profile filtering
│   ├── profiles.go             # Profile definitions, shortcuts, dependency validation
│   ├── readwise.go             # 9 readwise profile tool handlers
│   ├── reader.go               # 4 reader profile tool handlers
│   ├── write.go                # 7 write profile tool handlers
│   ├── video.go                # 5 video profile tool handlers (3 read-only, 2 write-requiring)
│   ├── destructive.go          # 4 destructive profile tool handlers
│   └── search.go               # Search over cached data (highlights + documents)
└── types/
    ├── source.go               # Source (book) types matching v2 API
    ├── highlight.go            # Highlight types matching v2 API
    ├── document.go             # Reader document types matching v3 API
    ├── tag.go                  # Tag types (shared)
    └── common.go               # Pagination, error response, config types

deploy/
├── kustomization.yml           # Kustomize config
├── readwise-mcp.yml            # StatefulSet + Service
├── vpa.yml                     # VerticalPodAutoscaler
└── nginx.conf                  # nginx sidecar config for TLS termination

Containerfile                   # Podman multi-stage build (linux/arm64)
Makefile                        # Build, test, lint, deploy targets
go.mod
go.sum
```

**Structure Decision**: Standard Go project layout with `cmd/` for the binary entry point and `internal/` for all private packages. The `internal/` tree mirrors the component diagram from the spec: server, api, cache, tools, types. Each tool profile gets its own file for maintainability. Deploy directory follows the k3s + Kustomize pattern from the constitution.

## Complexity Tracking

No constitution violations to justify. All design choices align with the 6 constitutional principles.
