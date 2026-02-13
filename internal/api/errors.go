package api

import (
	"encoding/json"
	"fmt"
)

// ErrorResponse represents a structured error returned to MCP clients.
type ErrorResponse struct {
	Type       string `json:"type"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Type, e.Code, e.Message)
}

// JSON serializes the error response.
func (e *ErrorResponse) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// NewValidationError creates a validation error response.
func NewValidationError(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Type:    "validation_error",
		Code:    code,
		Message: message,
	}
}

// NewAuthError creates an authentication error response.
func NewAuthError(message string) *ErrorResponse {
	return &ErrorResponse{
		Type:    "auth_error",
		Code:    "unauthorized",
		Message: message,
	}
}

// NewAPIError creates an upstream API error response.
func NewAPIError(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Type:    "api_error",
		Code:    code,
		Message: message,
	}
}

// NewRateLimitError creates a rate limit error with retry information.
func NewRateLimitError(retryAfter int) *ErrorResponse {
	return &ErrorResponse{
		Type:       "api_error",
		Code:       "rate_limited",
		Message:    fmt.Sprintf("Rate limited by upstream API. Retry after %d seconds.", retryAfter),
		RetryAfter: retryAfter,
	}
}

// NewInternalError creates an internal server error response.
func NewInternalError(message string) *ErrorResponse {
	return &ErrorResponse{
		Type:    "internal_error",
		Code:    "internal",
		Message: message,
	}
}
