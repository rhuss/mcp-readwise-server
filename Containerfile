# Build stage
FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -trimpath -ldflags="-s -w" \
    -o readwise-mcp-server \
    ./cmd/readwise-mcp/

# Runtime stage - minimal distroless image with CA certs, runs as non-root by default
FROM cgr.dev/chainguard/static:latest

COPY --from=builder /build/readwise-mcp-server /usr/local/bin/readwise-mcp-server

EXPOSE 8080
EXPOSE 8443

ENTRYPOINT ["readwise-mcp-server"]
