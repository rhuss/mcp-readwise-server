package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAuthHeaderInjection(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	_, err := client.GetV2(context.Background(), "/test", "my-secret-key")
	if err != nil {
		t.Fatalf("GetV2 error: %v", err)
	}

	if gotAuth != "Token my-secret-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Token my-secret-key")
	}
}

func TestClientRateLimitHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "42")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	_, err := client.GetV2(context.Background(), "/test", "key")

	if err == nil {
		t.Fatal("expected error for 429 response")
	}

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}

	if apiErr.Code != "rate_limited" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "rate_limited")
	}
	if apiErr.RetryAfter != 42 {
		t.Errorf("RetryAfter = %d, want 42", apiErr.RetryAfter)
	}
}

func TestClientUnauthorizedHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	_, err := client.GetV2(context.Background(), "/test", "bad-key")

	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}

	if apiErr.Type != "auth_error" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "auth_error")
	}
}

func TestClient404Handling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail":"not found"}`))
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	_, err := client.GetV2(context.Background(), "/test", "key")

	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}

	if apiErr.Type != "api_error" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "api_error")
	}
}

func TestClientNoContentHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	body, err := client.DeleteV2(context.Background(), "/test", "key")

	if err != nil {
		t.Fatalf("DeleteV2 error: %v", err)
	}
	if body != nil {
		t.Errorf("body = %v, want nil for 204 response", body)
	}
}

func TestClientV3Methods(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)

	tests := []struct {
		name     string
		call     func() ([]byte, error)
		expected string
	}{
		{"GetV3", func() ([]byte, error) { return client.GetV3(context.Background(), "/list", "key") }, "GET"},
		{"PostV3", func() ([]byte, error) {
			return client.PostV3(context.Background(), "/save", "key", map[string]string{"url": "https://example.com"})
		}, "POST"},
		{"PatchV3", func() ([]byte, error) {
			return client.PatchV3(context.Background(), "/update/1", "key", map[string]string{"title": "new"})
		}, "PATCH"},
		{"DeleteV3", func() ([]byte, error) { return client.DeleteV3(context.Background(), "/delete/1", "key") }, "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.call()
			if err != nil {
				t.Fatalf("%s error: %v", tt.name, err)
			}
			if gotMethod != tt.expected {
				t.Errorf("method = %q, want %q", gotMethod, tt.expected)
			}
		})
	}
}

func TestClientContentTypeOnPost(t *testing.T) {
	var gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	_, err := client.PostV2(context.Background(), "/test", "key", map[string]string{"text": "hello"})
	if err != nil {
		t.Fatalf("PostV2 error: %v", err)
	}

	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/json")
	}
}
