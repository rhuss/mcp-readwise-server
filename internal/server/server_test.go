package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := types.Config{
		Profiles:        []string{"readwise"},
		Port:            8080,
		LogLevel:        "info",
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
	}
	logger := slog.Default()
	s, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return s
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
}

func TestReadyEndpoint(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != "ready" {
		t.Errorf("status = %q, want %q", resp["status"], "ready")
	}
}

func TestMCPEndpointRegistered(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()

	s.Handler().ServeHTTP(rec, req)

	// The MCP handler should respond (not 404)
	if rec.Code == http.StatusNotFound {
		t.Error("expected /mcp endpoint to be registered")
	}
}

func TestServerCreationWithInvalidProfile(t *testing.T) {
	cfg := types.Config{
		Profiles:        []string{"nonexistent"},
		Port:            8080,
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
	}
	logger := slog.Default()
	_, err := New(cfg, logger)
	if err == nil {
		t.Fatal("expected error for invalid profile")
	}
}

func TestServerCreationWithAllProfiles(t *testing.T) {
	cfg := types.Config{
		Profiles:        []string{"all"},
		Port:            8080,
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
	}
	logger := slog.Default()
	s, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() with all profiles error: %v", err)
	}
	if s == nil {
		t.Fatal("expected server to be created")
	}
}
