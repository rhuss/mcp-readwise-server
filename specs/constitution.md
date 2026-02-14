# Readwise MCP Server Constitution

## Core Principles

### I. Stateless MCP Server
The server operates statelessly with API key passthrough. No server-side credential storage. Clients provide their own Readwise API key per request. The server is a thin, secure proxy layer between MCP clients and the Readwise/Reader APIs.

### II. Profile-Based Tool Exposure
Tools are grouped into profiles (readwise, reader, write, video, destructive) that control which capabilities are available. Profiles are composable, have explicit dependencies, and support shortcuts for common configurations. This keeps the tool surface area intentional and controllable.

### III. Go Implementation
The server is written in Go, targeting linux/arm64 for deployment on a home k3s cluster. Build with Podman, deploy with Kustomize. Follow standard Go project layout with `cmd/` and `internal/` packages.

### IV. Test-First Development
TDD is mandatory. Unit tests cover tool parameter validation, API client behavior, cache logic, and profile resolution. Integration tests use a mock Readwise API server. Tests must pass before any merge.

### V. Caching for Performance
Server-side LRU caching for read-heavy endpoints (exports, document lists, tags). Write operations invalidate relevant caches. Cache keys are isolated per API key (hashed). Memory-bounded with configurable limits.

### VI. Transparent Error Handling
API errors (including rate limits) are passed through to the client. No server-side retry or rate limiting. Error responses use a consistent structured format with type, code, and message.

## Technology Stack

- **Language**: Go 1.22+
- **Container**: Podman (multi-stage builds, linux/arm64)
- **Orchestration**: k3s with Kustomize
- **TLS**: Native Go TLS (nginx sidecar optional)
- **Protocol**: MCP over SSE/HTTP

## Governance

Constitution supersedes ad-hoc decisions. Amendments require explicit discussion and documentation.

**Version**: 1.1.0 | **Ratified**: 2026-02-13 | **Last Amended**: 2026-02-14
