package api

import (
	"encoding/json"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	e := NewValidationError("invalid_param", "page must be positive")
	if e.Type != "validation_error" {
		t.Errorf("Type = %q, want %q", e.Type, "validation_error")
	}
	if e.Code != "invalid_param" {
		t.Errorf("Code = %q, want %q", e.Code, "invalid_param")
	}
	if e.Message != "page must be positive" {
		t.Errorf("Message = %q, want %q", e.Message, "page must be positive")
	}
	if e.RetryAfter != 0 {
		t.Errorf("RetryAfter = %d, want 0", e.RetryAfter)
	}
}

func TestNewAuthError(t *testing.T) {
	e := NewAuthError("Invalid API key")
	if e.Type != "auth_error" {
		t.Errorf("Type = %q, want %q", e.Type, "auth_error")
	}
	if e.Code != "unauthorized" {
		t.Errorf("Code = %q, want %q", e.Code, "unauthorized")
	}
}

func TestNewAPIError(t *testing.T) {
	e := NewAPIError("not_found", "Resource not found")
	if e.Type != "api_error" {
		t.Errorf("Type = %q, want %q", e.Type, "api_error")
	}
}

func TestNewRateLimitError(t *testing.T) {
	e := NewRateLimitError(30)
	if e.Type != "api_error" {
		t.Errorf("Type = %q, want %q", e.Type, "api_error")
	}
	if e.Code != "rate_limited" {
		t.Errorf("Code = %q, want %q", e.Code, "rate_limited")
	}
	if e.RetryAfter != 30 {
		t.Errorf("RetryAfter = %d, want 30", e.RetryAfter)
	}
}

func TestNewInternalError(t *testing.T) {
	e := NewInternalError("something went wrong")
	if e.Type != "internal_error" {
		t.Errorf("Type = %q, want %q", e.Type, "internal_error")
	}
}

func TestErrorResponseError(t *testing.T) {
	e := NewValidationError("bad_input", "invalid")
	msg := e.Error()
	if msg != "validation_error: bad_input: invalid" {
		t.Errorf("Error() = %q, want %q", msg, "validation_error: bad_input: invalid")
	}
}

func TestErrorResponseJSON(t *testing.T) {
	e := NewRateLimitError(45)
	data, err := e.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Type != "api_error" {
		t.Errorf("decoded Type = %q, want %q", decoded.Type, "api_error")
	}
	if decoded.RetryAfter != 45 {
		t.Errorf("decoded RetryAfter = %d, want 45", decoded.RetryAfter)
	}
}

func TestErrorResponseJSONOmitsRetryAfterWhenZero(t *testing.T) {
	e := NewValidationError("test", "test")
	data, err := e.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := raw["retry_after"]; ok {
		t.Error("retry_after should be omitted when zero")
	}
}
