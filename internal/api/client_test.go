package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestClientAuthHeaderInjection(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
	_, err := client.GetV2(context.Background(), "/test", "my-secret-key")
	if err != nil {
		t.Fatalf("GetV2 error: %v", err)
	}

	if gotAuth != "Token my-secret-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Token my-secret-key")
	}
}

func TestClientRateLimitHandling(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
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
	if apiErr.RetryAfter != 0 {
		t.Errorf("RetryAfter = %d, want 0", apiErr.RetryAfter)
	}
	// Should have been called maxRetries (3) times due to retry logic
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestClientUnauthorizedHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
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

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
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

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
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

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))

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

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
	_, err := client.PostV2(context.Background(), "/test", "key", map[string]string{"text": "hello"})
	if err != nil {
		t.Fatalf("PostV2 error: %v", err)
	}

	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/json")
	}
}

func TestDoRequestRetryOn429(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
	body, err := client.GetV2(context.Background(), "/test", "key")

	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if string(body) != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", string(body), `{"status":"ok"}`)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

func TestDoRequestRetryExhausted(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
	_, err := client.GetV2(context.Background(), "/test", "key")

	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}
	if apiErr.Code != "rate_limited" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "rate_limited")
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3 (maxRetries)", callCount)
	}
}

func TestDoRequestRateLimiter(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	// Rate limiter: 10 requests per second (100ms between requests)
	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Every(100*time.Millisecond), 1))

	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := client.GetV2(context.Background(), "/test", "key")
		if err != nil {
			t.Fatalf("request %d error: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// 3 requests with 100ms spacing should take at least 200ms (first is immediate, then 2 waits)
	if elapsed < 150*time.Millisecond {
		t.Errorf("elapsed = %v, want >= 150ms (rate limiter should space requests)", elapsed)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestDoRequestRetryWithJitter(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
	_, err := client.GetV2(context.Background(), "/test", "key")

	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestDoRequestRetryAfterHeaderMissing(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return 429 without Retry-After header
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	// Use a context with a short timeout to avoid waiting the full 60s default backoff
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	client := NewClientWithRateLimiter(ts.URL, ts.URL, rate.NewLimiter(rate.Inf, 1))
	_, err := client.GetV2(ctx, "/test", "key")

	if err == nil {
		t.Fatal("expected error")
	}

	// Should have been called once (first attempt), then context timeout during backoff
	// since the default 60s backoff exceeds the 500ms context timeout
	if callCount < 1 {
		t.Errorf("callCount = %d, want >= 1", callCount)
	}

	// The error should indicate rate limiting or context cancellation
	apiErr, ok := err.(*ErrorResponse)
	if ok {
		// If all retries finished (unlikely with 60s default), check rate limit error
		if apiErr.Code == "rate_limited" && apiErr.RetryAfter != 60 {
			t.Errorf("RetryAfter = %d, want 60 (default)", apiErr.RetryAfter)
		}
	}
}
