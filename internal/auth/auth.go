package auth

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

// WithAPIKey stores an API key in the context.
func WithAPIKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, apiKeyContextKey, key)
}

// APIKeyFromRequest extracts the Readwise API key from an MCP CallToolRequest.
// The SDK populates req.Extra.Header with the HTTP request headers.
func APIKeyFromRequest(req *mcp.CallToolRequest) string {
	if req.Extra != nil && req.Extra.Header != nil {
		return ExtractAPIKeyFromHeader(req.Extra.Header)
	}
	return ""
}

// Middleware extracts the Readwise API key from the Authorization header
// and injects it into the request context. Used for non-MCP HTTP endpoints.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := ExtractAPIKeyFromHeader(r.Header)
		if apiKey != "" {
			r = r.WithContext(WithAPIKey(r.Context(), apiKey))
		}
		next.ServeHTTP(w, r)
	})
}

// ExtractAPIKeyFromHeader parses the API key from an HTTP header set.
// Supports formats: "Token <key>" and "Bearer <key>",
// and the custom "X-Readwise-Token" header.
func ExtractAPIKeyFromHeader(h http.Header) string {
	authHeader := h.Get("Authorization")
	if authHeader == "" {
		return h.Get("X-Readwise-Token")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	scheme := strings.ToLower(parts[0])
	if scheme == "token" || scheme == "bearer" {
		return strings.TrimSpace(parts[1])
	}

	return ""
}
