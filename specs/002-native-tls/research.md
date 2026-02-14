# Research: Native TLS Support

**Feature**: 002-native-tls | **Date**: 2026-02-14

## R1: Go stdlib TLS configuration

**Decision**: Use `crypto/tls.Config` with `http.Server.TLSConfig` and `ListenAndServeTLS`.

**Rationale**: Go's stdlib provides production-grade TLS support. No external library is needed. The `tls.Config` struct allows fine-grained control over TLS versions, cipher suites, and certificate loading. `tls.LoadX509KeyPair` handles PEM cert/key pair loading with clear error messages for format or mismatch issues.

**Alternatives considered**:
- `autocert` (Let's Encrypt): Requires public DNS and port 443 access. Not suitable for private k3s cluster.
- Third-party TLS libraries (e.g., `certmagic`): Adds dependency. Overkill for static cert file loading.

## R2: Dual-listener pattern in Go

**Decision**: Run two `http.Server` instances in separate goroutines, coordinated with `errgroup`.

**Rationale**: Each `http.Server` gets its own `ServeMux`. The HTTPS server mounts `/mcp`, the HTTP server mounts `/health` and `/ready`. Using `golang.org/x/sync/errgroup` (or a simple goroutine + channel pattern) allows clean startup/shutdown coordination. If either listener fails, both shut down.

**Alternatives considered**:
- Single listener with TLS-optional handler: Go's `net/http` does not support mixed TLS/plaintext on a single port. Would require custom `net.Listener` with TLS sniffing (e.g., `cmux`). Adds complexity and a dependency.
- `cmux` (connection multiplexer): Could serve both TLS and plain on one port, but adds a dependency and is harder to reason about for Kubernetes probes.

**Implementation note**: No new dependency needed. A goroutine with `sync.WaitGroup` or channels is sufficient. `errgroup` from `x/sync` is optional but clean. Given the constitution's preference for minimal dependencies, use stdlib `sync.WaitGroup` + channels.

## R3: Certificate expiry checking

**Decision**: Parse the certificate at startup, check `NotAfter`, and log a warning if it expires within 30 days.

**Rationale**: `tls.LoadX509KeyPair` returns a `tls.Certificate`. Call `x509.ParseCertificate` on `cert.Certificate[0]` (the leaf) to access `NotAfter`. Compare against `time.Now().Add(30 * 24 * time.Hour)`. This is a startup-only check, not continuous monitoring.

**Alternatives considered**:
- Periodic cert expiry check: Would require a background goroutine. Out of scope (live reload is a follow-up feature).
- Skip expiry check entirely: The spec requires FR-008.

## R4: Graceful shutdown with dual listeners

**Decision**: Use `context.Context` cancellation and `http.Server.Shutdown()` for both servers.

**Rationale**: `main.go` sets up a signal handler (`os.Signal` for `SIGTERM`, `SIGINT`). On signal, cancel the context. Both servers call `Shutdown(ctx)` with a timeout (e.g., 10 seconds) to drain inflight connections. This is the standard Go pattern for graceful HTTP shutdown.

**Alternatives considered**:
- `server.Close()` (immediate): Does not drain connections. Would drop inflight MCP requests.
- No signal handling: Current code does not handle signals (uses `http.ListenAndServe` which blocks). Adding signal handling is required for FR-010.

## R5: Kubernetes deployment changes

**Decision**: Remove nginx sidecar container, ConfigMap, and volume. Mount `readwise-mcp-tls` Secret directly into the Go container at `/etc/tls`.

**Rationale**: The existing `kubernetes.io/tls` Secret format produces `tls.crt` and `tls.key` files, which map directly to `TLS_CERT_FILE=/etc/tls/tls.crt` and `TLS_KEY_FILE=/etc/tls/tls.key`. The Service ports remain the same (8443 for HTTPS, 8080 for HTTP probes). Health probes continue targeting the `http` port unchanged.

**Alternatives considered**:
- Keep nginx as optional: Adds deployment complexity for no benefit when Go handles TLS natively.
- Use cert-manager's `csi-driver` instead of Secret mount: More complex setup. The existing Secret-based approach already works.

## R6: TLS version and cipher suite configuration

**Decision**: Configure `tls.Config` with `MinVersion: tls.VersionTLS12`. Let Go select cipher suites automatically.

**Rationale**: Go's default cipher suite selection is secure and regularly updated. Pinning cipher suites requires maintenance and can lock out clients. TLS 1.2 minimum matches the existing nginx configuration. TLS 1.3 cipher suites are not configurable in Go (by design, they are always strong).

**Alternatives considered**:
- TLS 1.3 only: Would break clients that only support TLS 1.2. The spec requires FR-009 support for both.
- Custom cipher suite list: Maintenance burden with no security benefit over Go's defaults.
