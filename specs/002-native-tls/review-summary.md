# Review Summary: Native TLS Support

**Spec:** specs/002-native-tls/spec.md | **Plan:** specs/002-native-tls/plan.md
**Generated:** 2026-02-14

> Distilled decision points for reviewers. See full spec/plan for details.

---

## Feature Overview

The readwise-mcp-server currently relies on an nginx sidecar container for TLS termination. This feature adds native HTTPS support directly in the Go binary, eliminating the nginx dependency. When TLS is configured (via `TLS_CERT_FILE` and `TLS_KEY_FILE` env vars), the server runs two listeners: HTTPS for MCP traffic and HTTP for health/readiness probes. When TLS is not configured, behavior is identical to today. The Kubernetes deployment is simplified to a single-container Pod.

## Scope Boundaries

- **In scope:** Native TLS via Go stdlib, dual-listener mode, deployment manifest updates, certificate expiry warning, graceful shutdown, constitution amendment
- **Out of scope:** Live certificate reload (follow-up feature), Let's Encrypt / autocert, custom cipher suite configuration, mutual TLS (mTLS)
- **Why these boundaries:** The goal is minimal, dependency-free TLS. Live reload and mTLS add complexity that isn't needed for the current deployment (home k3s cluster with manually provisioned certs).

## Critical Decisions

### Dual listeners vs. single listener with mixed protocol
- **Choice:** Two separate `http.Server` instances (HTTPS on 8443, HTTP on 8080)
- **Alternatives:** Connection multiplexing with `cmux` (single port), TLS-optional handler
- **Trade-off:** Two goroutines and shutdown coordination vs. adding a dependency (`cmux`) or custom protocol sniffing
- **Feedback:** Is the dual-listener approach acceptable, or would you prefer a single-port solution?

### Go stdlib TLS vs. external library
- **Choice:** `crypto/tls` from Go stdlib with `tls.LoadX509KeyPair`
- **Alternatives:** `certmagic` (auto-renewal), custom TLS manager
- **Trade-off:** No automatic cert renewal, but zero new dependencies
- **Feedback:** Is manual cert rotation (restart required) acceptable for this deployment?

### Certificate rotation strategy
- **Choice:** Restart required for cert rotation (no file watching)
- **Alternatives:** `fsnotify`-based watcher for automatic reload
- **Trade-off:** Simpler implementation now, operator must restart on cert change. Follow-up feature identified.
- **Feedback:** Should cert reload be prioritized as a near-term follow-up?

## Areas of Potential Disagreement

### Removing nginx entirely vs. keeping it optional
- **Decision:** Deployment manifests remove nginx completely. The Go server handles TLS.
- **Why this might be controversial:** nginx provides battle-tested TLS termination, request buffering, and rate limiting. Removing it loses those capabilities.
- **Alternative view:** Keep nginx as an optional deployment variant for operators who want its features.
- **Seeking input on:** Is the current deployment the only consumer? If other operators use this server, should we maintain an nginx variant?

### Health probes on plain HTTP in TLS mode
- **Decision:** `/health` and `/ready` stay on plain HTTP (port 8080) even when TLS is enabled.
- **Why this might be controversial:** Some security policies require all traffic to be encrypted.
- **Alternative view:** Serve probes over HTTPS too, or make it configurable.
- **Seeking input on:** Are there environments where HTTP health probes would violate security policy?

## Naming Decisions

| Item | Name | Context |
|------|------|---------|
| Cert path env var | `TLS_CERT_FILE` | Matches common Go conventions (`TLS_` prefix) |
| Key path env var | `TLS_KEY_FILE` | Pairs with `TLS_CERT_FILE` |
| HTTPS port env var | `TLS_PORT` | Short, clear. Alternative was `HTTPS_PORT` |
| Mount path in K8s | `/etc/tls` | Short. Alternative was `/etc/tls/certs` |
| Config method | `TLSEnabled()` | Returns bool. Alternative was `HasTLS()` |

## Schema Definitions

### Config (extended)

```go
type Config struct {
    // ... existing fields ...
    TLSCertFile string  // from TLS_CERT_FILE
    TLSKeyFile  string  // from TLS_KEY_FILE
    TLSPort     int     // from TLS_PORT, default 8443
}
```

## Architecture Choices

- **Pattern:** Conditional dual-listener (two `http.Server` instances coordinated via goroutines)
- **Components:** `types.Config` (validation), `server.Server` (listener management), `main.go` (signal handling)
- **Integration:** No changes to MCP handler, auth middleware, or tool registration. TLS is purely a transport-layer addition.

## Open Questions

- [ ] Should the HTTP health listener bind to `127.0.0.1` only or `0.0.0.0`? (Kubernetes probes need `0.0.0.0`, but `127.0.0.1` is more secure)

## Risk Areas

| Risk | Impact | Mitigation |
|------|--------|------------|
| SSE streaming over TLS | Med | Go's stdlib handles this natively. Existing nginx config disabled buffering; Go has no buffering by default. |
| Cert/key permission errors in distroless container | Low | Distroless runs as non-root. TLS Secret volumes are readable by default. |
| Breaking existing deployments | High | TLS is opt-in (disabled by default). No behavior change without env vars. |

---
*Share this with reviewers. Full context in linked spec and plan.*
