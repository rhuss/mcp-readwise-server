package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	ReadwiseV2BaseURL = "https://readwise.io/api/v2"
	ReaderV3BaseURL   = "https://readwise.io/api/v3"

	defaultTimeout = 30 * time.Second
)

// Client wraps an HTTP client for Readwise/Reader API calls.
type Client struct {
	httpClient *http.Client
	v2BaseURL  string
	v3BaseURL  string
}

// NewClient creates a new API client with default configuration.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		v2BaseURL:  ReadwiseV2BaseURL,
		v3BaseURL:  ReaderV3BaseURL,
	}
}

// NewClientWithBaseURLs creates a client with custom base URLs (for testing).
func NewClientWithBaseURLs(v2, v3 string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		v2BaseURL:  v2,
		v3BaseURL:  v3,
	}
}

// doRequest executes an HTTP request with the given API key and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, url, apiKey string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, NewInternalError(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to create request: %v", err))
	}

	req.Header.Set("Authorization", "Token "+apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAPIError("connection_error", fmt.Sprintf("failed to connect to API: %v", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to read response body: %v", err))
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := 60
		if v := resp.Header.Get("Retry-After"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				retryAfter = n
			}
		}
		return nil, NewRateLimitError(retryAfter)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, NewAuthError("Invalid or expired API key")
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, NewAPIError(
			fmt.Sprintf("http_%d", resp.StatusCode),
			fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)),
		)
	}

	return respBody, nil
}

// GetV2 performs a GET request against the Readwise v2 API.
func (c *Client) GetV2(ctx context.Context, path, apiKey string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, c.v2BaseURL+path, apiKey, nil)
}

// PostV2 performs a POST request against the Readwise v2 API.
func (c *Client) PostV2(ctx context.Context, path, apiKey string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, c.v2BaseURL+path, apiKey, body)
}

// PatchV2 performs a PATCH request against the Readwise v2 API.
func (c *Client) PatchV2(ctx context.Context, path, apiKey string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, c.v2BaseURL+path, apiKey, body)
}

// DeleteV2 performs a DELETE request against the Readwise v2 API.
func (c *Client) DeleteV2(ctx context.Context, path, apiKey string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, c.v2BaseURL+path, apiKey, nil)
}

// GetV3 performs a GET request against the Reader v3 API.
func (c *Client) GetV3(ctx context.Context, path, apiKey string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, c.v3BaseURL+path, apiKey, nil)
}

// PostV3 performs a POST request against the Reader v3 API.
func (c *Client) PostV3(ctx context.Context, path, apiKey string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, c.v3BaseURL+path, apiKey, body)
}

// PatchV3 performs a PATCH request against the Reader v3 API.
func (c *Client) PatchV3(ctx context.Context, path, apiKey string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, c.v3BaseURL+path, apiKey, body)
}

// DeleteV3 performs a DELETE request against the Reader v3 API.
func (c *Client) DeleteV3(ctx context.Context, path, apiKey string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, c.v3BaseURL+path, apiKey, nil)
}
