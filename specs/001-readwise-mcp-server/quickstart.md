# Quickstart: Readwise MCP Server

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13

## Prerequisites

- Go 1.22+
- Podman (for container builds)
- A Readwise account with an API access token ([readwise.io/access_token](https://readwise.io/access_token))

## Setup

### 1. Initialize the Go module

```bash
go mod init github.com/rhuss/readwise-mcp-server
go get github.com/modelcontextprotocol/go-sdk@v1.3.0
```

### 2. Build and run locally

```bash
go build -o readwise-mcp ./cmd/readwise-mcp
READWISE_PROFILES=readwise ./readwise-mcp
```

The server starts on port 8080 by default with Streamable HTTP transport.

### 3. Configure profiles

Set `READWISE_PROFILES` to control which tools are exposed:

```bash
# Read-only Readwise highlights
READWISE_PROFILES=readwise

# Reader documents + write access
READWISE_PROFILES=basic

# Everything
READWISE_PROFILES=all

# Custom combination
READWISE_PROFILES=readwise,reader,write
```

### 4. Connect an MCP client

Configure your MCP client (e.g., Claude Desktop) to connect to the server. The client must pass the Readwise API key in the HTTP request headers.

### 5. Run tests

```bash
# All tests
go test ./...

# With verbose output
go test -v ./...

# Specific package
go test ./internal/cache/...
go test ./internal/tools/...
```

### 6. Build container

```bash
podman build --platform linux/arm64 -t readwise-mcp:latest -f Containerfile .
```

### 7. Deploy to k3s

```bash
kustomize build deploy/ | kubectl apply -f -
```

## Environment Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `READWISE_PROFILES` | `readwise` | Comma-separated profiles to enable |
| `PORT` | `8080` | HTTP port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `CACHE_MAX_SIZE_MB` | `128` | Maximum cache memory in MB |
| `CACHE_TTL_SECONDS` | `300` | Default cache TTL in seconds |
| `CACHE_ENABLED` | `true` | Enable/disable caching |

## Development Order

The recommended implementation order follows the user story priorities:

1. **Foundation**: Config loading, API client, error types, MCP server skeleton
2. **P1 - Readwise read tools** (Story 1): list_sources, get_source, list_highlights, get_highlight, export_highlights, get_daily_review, list_source_tags, list_highlight_tags, search_highlights
3. **P1 - Reader read tools** (Story 2): list_documents, get_document, list_reader_tags, search_documents
4. **P2 - Profile system** (Story 4): Profile definitions, shortcuts, dependency validation, tool filtering
5. **P2 - Cache system** (Story 7): LRU cache, TTL, memory limits, per-user isolation, invalidation
6. **P2 - Write tools** (Story 3): save_document, update_document, create_highlight, update_highlight, add_source_tag, add_highlight_tag, bulk_create_highlights
7. **P3 - Video tools** (Story 5): list_videos, get_video, get_video_position, update_video_position, create_video_highlight
8. **P3 - Destructive tools** (Story 6): delete_highlight, delete_highlight_tag, delete_source_tag, delete_document
9. **Health endpoints** (FR-015): /health, /ready
10. **Deployment**: Containerfile, Kustomize manifests, nginx config
