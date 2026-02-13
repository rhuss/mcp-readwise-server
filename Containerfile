# Build stage
FROM docker.io/library/golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -trimpath -ldflags="-s -w" \
    -o readwise-mcp-server \
    ./cmd/readwise-mcp/

# Runtime stage
FROM docker.io/library/alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/readwise-mcp-server /usr/local/bin/readwise-mcp-server

EXPOSE 8080

USER 65534:65534

ENTRYPOINT ["readwise-mcp-server"]
