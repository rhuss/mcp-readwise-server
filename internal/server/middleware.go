package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type contextKey string

const apiKeyContextKey contextKey = "api_key"

// APIKeyFromContext retrieves the API key stored in the context.
func APIKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(apiKeyContextKey).(string); ok {
		return v
	}
	return ""
}

// withAPIKey stores an API key in the context.
func withAPIKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, apiKeyContextKey, key)
}

// APIKeyFromRequest extracts the Readwise API key from an MCP CallToolRequest.
// The SDK populates req.Extra.Header with the HTTP request headers.
func APIKeyFromRequest(req *mcp.CallToolRequest) string {
	if req.Extra != nil && req.Extra.Header != nil {
		return extractAPIKeyFromHeader(req.Extra.Header)
	}
	return ""
}

// APIKeyMiddleware extracts the Readwise API key from the Authorization header
// and injects it into the request context. Used for non-MCP HTTP endpoints.
func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := extractAPIKeyFromHeader(r.Header)
		if apiKey != "" {
			r = r.WithContext(withAPIKey(r.Context(), apiKey))
		}
		next.ServeHTTP(w, r)
	})
}

// extractAPIKeyFromHeader parses the API key from an HTTP header set.
// Supports formats: "Token <key>" and "Bearer <key>",
// and the custom "X-Readwise-Token" header.
func extractAPIKeyFromHeader(h http.Header) string {
	auth := h.Get("Authorization")
	if auth == "" {
		return h.Get("X-Readwise-Token")
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	scheme := strings.ToLower(parts[0])
	if scheme == "token" || scheme == "bearer" {
		return strings.TrimSpace(parts[1])
	}

	return ""
}
