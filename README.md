# <img src="assets/logo.png" alt="" width="96" valign="middle"> &nbsp;&nbsp; Readwise MCP Server

A Go-based [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server that gives AI assistants access to your [Readwise](https://readwise.io) highlights and [Reader](https://read.readwise.io) library.

The server acts as a stateless proxy: clients pass their Readwise API key in every request, and the server forwards it to the Readwise API. No credentials are stored.

## Features

- **29 MCP tools** covering highlights, documents, tags, videos, and search
- **Profile system** to control which tools are exposed
- **Native TLS** with dual-listener mode (HTTPS for MCP, HTTP for health probes)
- **In-memory LRU cache** with per-user isolation and automatic invalidation
- **Distroless container image** for minimal attack surface
- **Kubernetes-ready** with health/readiness probes and VPA support

## Quick Start

### Run with Podman

```bash
# Build the container image
podman build -t readwise-mcp-server -f Containerfile .

# Run with default settings (readwise profile, no TLS)
podman run -p 8080:8080 readwise-mcp-server

# Run with multiple profiles
podman run -p 8080:8080 \
  -e READWISE_PROFILES=basic \
  readwise-mcp-server
```

### Build from Source

```bash
make build
READWISE_PROFILES=basic ./build/readwise-mcp-server
```

## Client Configuration

Point your MCP client at the server's `/mcp` endpoint, passing your Readwise API key in the `Authorization` header.

**Claude Code** (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "readwise": {
      "type": "streamable-http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token YOUR_READWISE_API_KEY"
      }
    }
  }
}
```

Get your API key at [readwise.io/access_token](https://readwise.io/access_token).

## Profiles

Profiles control which tools the server exposes. Set them via the `READWISE_PROFILES` environment variable as a comma-separated list.

### Base Profiles

| Profile | Type | Tools | Dependencies |
|---------|------|-------|--------------|
| `readwise` | read | 9 tools for Readwise highlights API (v2) | none |
| `reader` | read | 4 tools for Reader documents API (v3) | none |
| `write` | modifier | 7 tools for creating/updating content | `readwise` or `reader` |
| `video` | modifier | 5 tools for video documents and playback | `reader` |
| `destructive` | modifier | 4 tools for deleting content | `readwise` or `reader` |

### Shortcuts

| Shortcut | Expands to |
|----------|------------|
| `basic` | `reader`, `write` |
| `all` | `readwise`, `reader`, `write`, `video`, `destructive` |

### Profile Examples

```bash
# Default: only Readwise highlights (read-only)
READWISE_PROFILES=readwise

# Reader with write access (same as "basic")
READWISE_PROFILES=reader,write

# Everything except destructive operations
READWISE_PROFILES=readwise,reader,write,video

# All tools enabled
READWISE_PROFILES=all
```

Dependencies are validated at startup. The server will refuse to start if a modifier profile is enabled without its required read profile (e.g., `write` without `readwise` or `reader`).

## Tools

### Readwise Profile (9 tools)

| Tool | Description |
|------|-------------|
| `list_sources` | List highlight sources (books, articles, etc.) with pagination and filtering |
| `get_source` | Get details of a single source by ID |
| `list_highlights` | List highlights with pagination and filtering |
| `get_highlight` | Get a single highlight by ID |
| `export_highlights` | Bulk export all highlights grouped by source |
| `get_daily_review` | Get today's daily review highlights |
| `list_source_tags` | List all tags on a specific source |
| `list_highlight_tags` | List all tags on a specific highlight |
| `search_highlights` | Search highlights by query across text, notes, and titles |

### Reader Profile (4 tools)

| Tool | Description |
|------|-------------|
| `list_documents` | List Reader documents with filtering by location/category |
| `get_document` | Get a single document, optionally with full content |
| `list_reader_tags` | List all tags in Reader |
| `search_documents` | Search documents across title, author, summary, and notes |

### Write Profile (7 tools)

| Tool | Description |
|------|-------------|
| `save_document` | Save a URL to Reader |
| `update_document` | Update document metadata (title, author, summary, location, tags) |
| `create_highlight` | Create a new highlight on a source |
| `update_highlight` | Update an existing highlight's text, note, location, or color |
| `add_source_tag` | Add a tag to a source |
| `add_highlight_tag` | Add a tag to a highlight |
| `bulk_create_highlights` | Create multiple highlights in a single request |

### Video Profile (5 tools)

| Tool | Description |
|------|-------------|
| `list_videos` | List video documents from Reader |
| `get_video` | Get a video document with transcript |
| `get_video_position` | Get the current playback position |
| `update_video_position` | Update the playback position (requires `write`) |
| `create_video_highlight` | Create a timestamped highlight on a video (requires `write`) |

### Destructive Profile (4 tools)

| Tool | Description |
|------|-------------|
| `delete_highlight` | Delete a highlight permanently |
| `delete_highlight_tag` | Remove a tag from a highlight |
| `delete_source_tag` | Remove a tag from a source |
| `delete_document` | Delete a Reader document permanently |

## Configuration

All configuration is via environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `READWISE_PROFILES` | `readwise` | Comma-separated profile names |
| `PORT` | `8080` | HTTP port (health probes, or MCP when TLS is off) |
| `TLS_CERT_FILE` | | Path to TLS certificate PEM file |
| `TLS_KEY_FILE` | | Path to TLS private key PEM file |
| `TLS_PORT` | `8443` | HTTPS port for MCP endpoint (when TLS is on) |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `CACHE_ENABLED` | `true` | Enable in-memory response cache |
| `CACHE_MAX_SIZE_MB` | `128` | Maximum cache size in MB |
| `CACHE_TTL_SECONDS` | `300` | Default cache TTL in seconds |

## TLS

The server supports native TLS with a dual-listener architecture:

- **HTTPS** on `TLS_PORT` (default 8443): serves the `/mcp` endpoint
- **HTTP** on `PORT` (default 8080): serves `/health` and `/ready` probes

This lets Kubernetes probe the container over plain HTTP while MCP traffic is encrypted.

### Enabling TLS

Set both `TLS_CERT_FILE` and `TLS_KEY_FILE`. The server validates that both files exist and are readable, and that `TLS_PORT` differs from `PORT`.

```bash
podman run -p 8080:8080 -p 8443:8443 \
  -v ./certs:/etc/tls:ro \
  -e TLS_CERT_FILE=/etc/tls/tls.crt \
  -e TLS_KEY_FILE=/etc/tls/tls.key \
  -e READWISE_PROFILES=basic \
  readwise-mcp-server
```

When TLS is not configured, all endpoints are served over HTTP on `PORT`.

## Health Endpoints

| Endpoint | Port | Description |
|----------|------|-------------|
| `/health` | HTTP | Liveness probe, always returns 200 |
| `/ready` | HTTP | Readiness probe, always returns 200 |
| `/mcp` | HTTPS (or HTTP) | MCP protocol endpoint |

## Caching

The server caches API responses per user (keyed by a hash of the API key) with LRU eviction.

- Export and list endpoints use a 5-minute TTL
- Tag listing uses a 10-minute TTL
- Write and delete operations automatically invalidate affected cache entries
- Disable caching with `CACHE_ENABLED=false`

## Deployment on Kubernetes

The following example deploys the server as a StatefulSet with native TLS on a Kubernetes cluster.

### Generate TLS Certificates

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt \
  -subj "/CN=mcp-readwise.example.com" \
  -addext "subjectAltName=DNS:mcp-readwise.example.com,IP:10.0.0.100"
```

### Create the TLS Secret

Use a kustomize secret generator or create the Secret directly:

```yaml
# auth/kustomization.yml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

secretGenerator:
- name: readwise-mcp-tls
  type: kubernetes.io/tls
  files:
  - tls.crt
  - tls.key

generatorOptions:
  disableNameSuffixHash: true
```

### StatefulSet and Service

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: readwise-mcp
spec:
  replicas: 1
  serviceName: readwise-mcp-pods
  selector:
    matchLabels:
      app: readwise-mcp
  template:
    metadata:
      labels:
        app: readwise-mcp
    spec:
      containers:
      - name: mcp
        image: readwise-mcp-server:latest
        ports:
        - name: http
          containerPort: 8080
        - name: https
          containerPort: 8443
        env:
        - name: READWISE_PROFILES
          value: "basic"
        - name: PORT
          value: "8080"
        - name: TLS_CERT_FILE
          value: "/etc/tls/tls.crt"
        - name: TLS_KEY_FILE
          value: "/etc/tls/tls.key"
        - name: TLS_PORT
          value: "8443"
        - name: LOG_LEVEL
          value: "info"
        - name: CACHE_ENABLED
          value: "true"
        volumeMounts:
        - name: tls-certs
          mountPath: /etc/tls
          readOnly: true
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 5
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 3
          periodSeconds: 10
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
      volumes:
      - name: tls-certs
        secret:
          secretName: readwise-mcp-tls
---
apiVersion: v1
kind: Service
metadata:
  name: readwise-mcp
spec:
  type: LoadBalancer
  selector:
    app: readwise-mcp
  ports:
  - name: https
    port: 8443
    targetPort: https
```

### Kustomization

```yaml
# kustomization.yml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: mcp

resources:
- readwise-mcp.yml
- auth/

labels:
- includeSelectors: true
  pairs:
    app: readwise-mcp

images:
- name: readwise-mcp-server
  newName: docker.io/youruser/readwise-mcp-server
  newTag: latest
```

### Deploy

```bash
kustomize build . | kubectl apply -f -
```

### Optional: Vertical Pod Autoscaler

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: readwise-mcp
spec:
  targetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: readwise-mcp
  updatePolicy:
    updateMode: "Off"
  resourcePolicy:
    containerPolicies:
    - containerName: mcp
      controlledResources: ["cpu", "memory"]
      minAllowed:
        cpu: 50m
        memory: 64Mi
      maxAllowed:
        cpu: 1
        memory: 512Mi
```

## Development

```bash
make build       # Build binary to build/readwise-mcp-server
make test        # Run tests with race detector
make lint        # Run go vet and golangci-lint
make container   # Build container image with Podman
make clean       # Remove build artifacts
```

## License

MIT
