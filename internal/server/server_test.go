package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func testdataPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", file)
}

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

func newTLSTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := types.Config{
		Profiles:        []string{"readwise"},
		Port:            0, // will be assigned dynamically
		LogLevel:        "info",
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
		TLSCertFile:     testdataPath("server-cert.pem"),
		TLSKeyFile:      testdataPath("server-key.pem"),
		TLSPort:         0, // will be assigned dynamically
	}
	logger := slog.Default()
	s, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return s
}

func TestTLSListenerStartup(t *testing.T) {
	s := newTLSTestServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ListenAndServe(ctx)
	}()

	// Wait for both listeners to start
	time.Sleep(500 * time.Millisecond)

	// Test HTTPS listener (MCP endpoint)
	tlsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	httpsURL := fmt.Sprintf("https://localhost:%d/mcp", s.tlsPort())
	resp, err := tlsClient.Get(httpsURL)
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		t.Error("expected /mcp endpoint on HTTPS listener")
	}

	// Test HTTP listener (health endpoint)
	httpURL := fmt.Sprintf("http://localhost:%d/health", s.httpPort())
	resp, err = http.Get(httpURL)
	if err != nil {
		t.Fatalf("HTTP health request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Test HTTP listener (ready endpoint)
	httpURL = fmt.Sprintf("http://localhost:%d/ready", s.httpPort())
	resp, err = http.Get(httpURL)
	if err != nil {
		t.Fatalf("HTTP ready request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ready status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Shutdown
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("ListenAndServe returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down in time")
	}
}

func TestCertExpiryWarning(t *testing.T) {
	// Test with near-expiry cert
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	cfg := types.Config{
		Profiles:        []string{"readwise"},
		Port:            0,
		LogLevel:        "info",
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
		TLSCertFile:     testdataPath("expiring-cert.pem"),
		TLSKeyFile:      testdataPath("expiring-key.pem"),
		TLSPort:         0,
	}

	s, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.ListenAndServe(ctx)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "TLS certificate expires") {
		t.Errorf("expected cert expiry warning in logs, got: %s", logOutput)
	}
}

func TestCertNoExpiryWarning(t *testing.T) {
	// Test with long-lived cert (no warning expected)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	cfg := types.Config{
		Profiles:        []string{"readwise"},
		Port:            0,
		LogLevel:        "info",
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
		TLSCertFile:     testdataPath("server-cert.pem"),
		TLSKeyFile:      testdataPath("server-key.pem"),
		TLSPort:         0,
	}

	s, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.ListenAndServe(ctx)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)

	logOutput := buf.String()
	if strings.Contains(logOutput, "TLS certificate expires") {
		t.Errorf("did not expect cert expiry warning, got: %s", logOutput)
	}
}

func TestNonTLSBackwardCompatibility(t *testing.T) {
	cfg := types.Config{
		Profiles:        []string{"readwise"},
		Port:            0,
		LogLevel:        "info",
		CacheMaxSizeMB:  16,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
		// No TLS config
	}
	logger := slog.Default()
	s, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ListenAndServe(ctx)
	}()

	time.Sleep(500 * time.Millisecond)

	port := s.httpPort()
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// All endpoints should work over HTTP
	for _, path := range []string{"/health", "/ready", "/mcp"} {
		resp, err := http.Get(baseURL + path)
		if err != nil {
			t.Fatalf("GET %s failed: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("expected %s endpoint to be registered on HTTP", path)
		}
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("ListenAndServe returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down in time")
	}
}
