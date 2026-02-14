# Data Model: Native TLS Support

**Feature**: 002-native-tls | **Date**: 2026-02-14

## Entities

### Config (modified)

Extends the existing `types.Config` struct with TLS fields.

| Field | Type | Default | Source | Description |
|-------|------|---------|--------|-------------|
| TLSCertFile | string | "" | `TLS_CERT_FILE` | Path to PEM-encoded TLS certificate |
| TLSKeyFile | string | "" | `TLS_KEY_FILE` | Path to PEM-encoded TLS private key |
| TLSPort | int | 8443 | `TLS_PORT` | HTTPS listen port |

### Derived State

| Method | Returns | Description |
|--------|---------|-------------|
| `TLSEnabled()` | bool | True when both `TLSCertFile` and `TLSKeyFile` are non-empty |
| `TLSAddr()` | string | Returns `":TLSPort"` formatted address string |

### Validation Rules

| Rule | Error Condition | Error Message |
|------|-----------------|---------------|
| Both or neither | Only one of `TLS_CERT_FILE` / `TLS_KEY_FILE` set | "TLS configuration incomplete: both TLS_CERT_FILE and TLS_KEY_FILE must be set" |
| File existence | `TLS_CERT_FILE` path does not exist | "TLS certificate file not found: {path}" |
| File existence | `TLS_KEY_FILE` path does not exist | "TLS key file not found: {path}" |
| Port conflict | `TLS_PORT` equals `PORT` | "TLS port {port} conflicts with HTTP port" |
| Cert validity | Certificate expires within 30 days | Warning (non-fatal): "TLS certificate expires in {N} days" |

## No New Entities

This feature does not introduce new persistent entities. It modifies the existing `Config` struct and adds behavioral changes to the `Server` struct. No database, file storage, or external state is involved.
