# Quickstart: Native TLS Support

**Feature**: 002-native-tls | **Date**: 2026-02-14

## Local Development (without TLS)

No changes. The server runs on HTTP as before:

```bash
PORT=8080 go run ./cmd/readwise-mcp/
```

## Local Development (with TLS)

Generate self-signed test certificates:

```bash
openssl req -x509 -newkey rsa:2048 -keyout test-key.pem -out test-cert.pem \
  -days 365 -nodes -subj '/CN=localhost'
```

Start the server with TLS:

```bash
TLS_CERT_FILE=test-cert.pem TLS_KEY_FILE=test-key.pem \
  PORT=8080 TLS_PORT=8443 \
  go run ./cmd/readwise-mcp/
```

Verify:

```bash
# HTTPS (MCP endpoint)
curl -k https://localhost:8443/mcp

# HTTP (health probe)
curl http://localhost:8080/health
```

## Kubernetes Deployment

The deployment expects a `kubernetes.io/tls` Secret named `readwise-mcp-tls`:

```bash
# Create TLS secret (if not using cert-manager)
kubectl -n readwise-mcp create secret tls readwise-mcp-tls \
  --cert=tls.crt --key=tls.key

# Apply deployment
kustomize build deploy/ | kubectl apply -f -
```

The Go container reads certificates from `/etc/tls/tls.crt` and `/etc/tls/tls.key`.

## Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `TLS_CERT_FILE` | (empty) | No | Path to PEM certificate. TLS enabled when set with `TLS_KEY_FILE` |
| `TLS_KEY_FILE` | (empty) | No | Path to PEM private key. TLS enabled when set with `TLS_CERT_FILE` |
| `TLS_PORT` | 8443 | No | HTTPS listen port (only used when TLS enabled) |
| `PORT` | 8080 | No | HTTP listen port (probes in TLS mode, all traffic otherwise) |

## Verifying the Deployment

```bash
# Check single container (no nginx sidecar)
kubectl -n readwise-mcp get pods -o jsonpath='{.items[0].spec.containers[*].name}'
# Expected: readwise-mcp

# Check HTTPS
curl -k https://<service-ip>:8443/mcp

# Check health probe (from within cluster)
curl http://<pod-ip>:8080/health
```
