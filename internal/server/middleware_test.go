package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
)

func TestExtractAPIKeyFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name:     "Token format",
			headers:  map[string]string{"Authorization": "Token my-api-key"},
			expected: "my-api-key",
		},
		{
			name:     "Bearer format",
			headers:  map[string]string{"Authorization": "Bearer my-api-key"},
			expected: "my-api-key",
		},
		{
			name:     "custom header",
			headers:  map[string]string{"X-Readwise-Token": "custom-key"},
			expected: "custom-key",
		},
		{
			name:     "no header",
			headers:  map[string]string{},
			expected: "",
		},
		{
			name:     "invalid format",
			headers:  map[string]string{"Authorization": "Basic abc123"},
			expected: "",
		},
		{
			name:     "missing value",
			headers:  map[string]string{"Authorization": "Token"},
			expected: "",
		},
		{
			name:     "auth takes precedence over custom header",
			headers:  map[string]string{"Authorization": "Token auth-key", "X-Readwise-Token": "custom-key"},
			expected: "auth-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			got := auth.ExtractAPIKeyFromHeader(h)
			if got != tt.expected {
				t.Errorf("ExtractAPIKeyFromHeader() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAPIKeyMiddleware(t *testing.T) {
	var gotKey string
	handler := APIKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = APIKeyFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token middleware-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if gotKey != "middleware-key" {
		t.Errorf("APIKeyFromContext() = %q, want %q", gotKey, "middleware-key")
	}
}

func TestAPIKeyMiddlewareMissing(t *testing.T) {
	var gotKey string
	handler := APIKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = APIKeyFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if gotKey != "" {
		t.Errorf("APIKeyFromContext() = %q, want empty", gotKey)
	}
}

func TestAPIKeyFromRequest(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Token request-key")

	req := &mcp.CallToolRequest{
		Extra: &mcp.RequestExtra{
			Header: h,
		},
	}

	got := APIKeyFromRequest(req)
	if got != "request-key" {
		t.Errorf("APIKeyFromRequest() = %q, want %q", got, "request-key")
	}
}

func TestAPIKeyFromRequestNilExtra(t *testing.T) {
	req := &mcp.CallToolRequest{}
	got := APIKeyFromRequest(req)
	if got != "" {
		t.Errorf("APIKeyFromRequest() = %q, want empty", got)
	}
}

func TestAPIKeyFromContextNil(t *testing.T) {
	got := APIKeyFromContext(nil) //nolint:staticcheck
	if got != "" {
		t.Errorf("APIKeyFromContext(nil) = %q, want empty", got)
	}
}

func TestAPIKeyFromContextEmpty(t *testing.T) {
	got := APIKeyFromContext(context.Background())
	if got != "" {
		t.Errorf("APIKeyFromContext(empty) = %q, want empty", got)
	}
}
