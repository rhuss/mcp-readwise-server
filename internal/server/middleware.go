package server

import (
	"github.com/rhuss/readwise-mcp-server/internal/auth"
)

// Re-export auth functions for backward compatibility with server package tests.
var (
	APIKeyFromContext = auth.APIKeyFromContext
	APIKeyFromRequest = auth.APIKeyFromRequest
	APIKeyMiddleware  = auth.Middleware
)
