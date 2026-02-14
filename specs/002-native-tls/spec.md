# Feature Specification: Native TLS Support

**Feature Branch**: `002-native-tls`
**Created**: 2026-02-14
**Status**: Draft
**Input**: User description: "Native HTTPS support without nginx sidecar dependency, using dual listeners (HTTPS for MCP, HTTP for health probes)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Operator enables TLS on the MCP server (Priority: P1)

An operator deploying the readwise-mcp-server wants encrypted communication for MCP clients without running a separate nginx sidecar container. They configure TLS by providing certificate and key file paths through environment variables. The server starts serving MCP traffic over HTTPS while keeping health and readiness probes accessible over plain HTTP.

**Why this priority**: This is the core feature. Without it, the server requires an nginx sidecar for any TLS deployment, adding operational complexity (extra container, ConfigMap, and volume mounts).

**Independent Test**: Can be fully tested by starting the server with TLS environment variables set and verifying that the `/mcp` endpoint responds over HTTPS and `/health` responds over HTTP.

**Acceptance Scenarios**:

1. **Given** the server is configured with `TLS_CERT_FILE` and `TLS_KEY_FILE` pointing to valid PEM files, **When** the server starts, **Then** it listens on port 8443 (HTTPS) for MCP traffic and port 8080 (HTTP) for health probes.
2. **Given** the server is configured with TLS enabled, **When** an MCP client connects to the HTTPS port, **Then** the connection is encrypted using the provided certificate.
3. **Given** the server is configured with TLS enabled, **When** a Kubernetes liveness probe sends an HTTP request to `/health` on port 8080, **Then** it receives a successful response without needing TLS configuration.

---

### User Story 2 - Operator runs without TLS (backward compatibility) (Priority: P1)

An operator who does not need TLS (e.g., running behind a load balancer or in a trusted network) starts the server without setting any TLS environment variables. The server behaves exactly as it does today: all endpoints served over plain HTTP on a single port.

**Why this priority**: Equal to P1 because breaking backward compatibility would disrupt all existing deployments.

**Independent Test**: Can be tested by starting the server without TLS environment variables and verifying all endpoints respond over HTTP on the configured port.

**Acceptance Scenarios**:

1. **Given** no TLS environment variables are set, **When** the server starts, **Then** it listens on port 8080 (HTTP) and serves all endpoints (`/mcp`, `/health`, `/ready`) over plain HTTP.
2. **Given** no TLS environment variables are set, **When** the server starts, **Then** no TLS-related log messages or errors appear.

---

### User Story 3 - Kubernetes deployment without nginx sidecar (Priority: P2)

An operator deploys the readwise-mcp-server on Kubernetes using a standard `kubernetes.io/tls` Secret. The TLS Secret is mounted directly into the Go container, and the deployment manifest no longer includes an nginx sidecar container, its ConfigMap, or its volume mounts. The Service routes HTTPS traffic directly to the Go container.

**Why this priority**: This simplifies the deployment topology and removes the nginx dependency, but requires a working TLS implementation (P1) first.

**Independent Test**: Can be tested by applying the updated Kubernetes manifests and verifying the MCP endpoint is reachable over HTTPS through the Service, with only a single container running in the Pod.

**Acceptance Scenarios**:

1. **Given** a Kubernetes deployment with a `kubernetes.io/tls` Secret mounted at `/etc/tls`, **When** the Pod starts, **Then** the Go container reads `tls.crt` and `tls.key` from that path and serves HTTPS.
2. **Given** the updated deployment manifest, **When** `kubectl get pods` is run, **Then** each Pod shows only one container (no nginx sidecar).

---

### Edge Cases

- What happens when only one of `TLS_CERT_FILE` or `TLS_KEY_FILE` is set? The server should fail to start with a clear error message indicating both are required.
- What happens when the certificate or key file does not exist at the specified path? The server should fail to start with an error identifying the missing file.
- What happens when the certificate and key do not match? The server should fail to start with an error from TLS initialization.
- What happens when the TLS port conflicts with the HTTP port? The server should fail to start with an error indicating the port conflict.
- What happens when the certificate file contains an expired certificate? The server should still start (certificate validity is the operator's responsibility), but log a warning if the certificate expires within 30 days.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept TLS configuration through environment variables `TLS_CERT_FILE` (path to PEM certificate) and `TLS_KEY_FILE` (path to PEM private key).
- **FR-002**: System MUST accept a configurable HTTPS port through `TLS_PORT` (default: 8443).
- **FR-003**: When both `TLS_CERT_FILE` and `TLS_KEY_FILE` are set, system MUST start two listeners: HTTPS on `TLS_PORT` for `/mcp` and HTTP on `PORT` for `/health` and `/ready`.
- **FR-004**: When TLS is not configured, system MUST behave identically to the current implementation (single HTTP listener on `PORT` serving all endpoints).
- **FR-005**: System MUST fail to start with a descriptive error if only one of `TLS_CERT_FILE` or `TLS_KEY_FILE` is provided.
- **FR-006**: System MUST fail to start with a descriptive error if the specified certificate or key files cannot be read.
- **FR-007**: System MUST fail to start with a descriptive error if `TLS_PORT` equals `PORT`.
- **FR-008**: System MUST log a warning at startup if the TLS certificate expires within 30 days.
- **FR-009**: System MUST support TLS 1.2 and TLS 1.3.
- **FR-010**: System MUST shut down both listeners gracefully when receiving a termination signal.

### Key Entities

- **TLS Configuration**: Certificate file path, key file path, HTTPS port. Derived entirely from environment variables. Determines whether the server runs in single-listener (HTTP) or dual-listener (HTTP + HTTPS) mode.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Server starts and serves MCP traffic over HTTPS when TLS is configured, verified by a successful client connection using the server's certificate.
- **SC-002**: Health and readiness probes remain accessible over plain HTTP regardless of TLS configuration.
- **SC-003**: Kubernetes deployment runs with a single container per Pod (no nginx sidecar) while maintaining HTTPS for client traffic.
- **SC-004**: Server starts in under 2 seconds in both TLS and non-TLS modes.
- **SC-005**: All existing functionality (tool registration, MCP protocol handling, caching, authentication) works identically over HTTPS as it does over HTTP.

## Assumptions

- Certificates are managed externally (e.g., by cert-manager or manual provisioning). The server does not generate, renew, or manage certificates.
- Certificate rotation requires a server restart. Live certificate reload is out of scope for this feature (identified as a potential follow-up).
- The `kubernetes.io/tls` Secret format (with `tls.crt` and `tls.key` keys) is the standard for certificate delivery in Kubernetes.
- The HTTP listener in dual-listener mode is only accessible within the Pod or cluster network, not exposed externally through the Service.

## Constitution Alignment

This feature amends the constitution's Technology Stack section (line 28: "TLS: nginx sidecar for TLS termination") to "TLS: Native Go TLS with optional nginx sidecar". The change aligns with:

- **Principle III (Go Implementation)**: Keeps TLS handling within the Go binary, reducing external dependencies.
- **Principle I (Stateless MCP Server)**: TLS configuration is stateless (file paths from env vars), consistent with the server's stateless design.
- **Principle VI (Transparent Error Handling)**: TLS errors (missing files, mismatched certs) produce clear, actionable error messages.
