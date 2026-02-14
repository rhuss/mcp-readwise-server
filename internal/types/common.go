package types

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TLS validation errors.
var (
	ErrTLSIncomplete = fmt.Errorf("TLS configuration incomplete: both TLS_CERT_FILE and TLS_KEY_FILE must be set")
)

// Config holds server configuration loaded from environment variables.
type Config struct {
	Profiles        []string
	Port            int
	LogLevel        string
	CacheMaxSizeMB  int
	CacheTTLSeconds int
	CacheEnabled    bool
	TLSCertFile     string
	TLSKeyFile      string
	TLSPort         int
}

// LoadConfig reads configuration from environment variables with defaults.
func LoadConfig() Config {
	c := Config{
		Profiles:        []string{"readwise"},
		Port:            8080,
		LogLevel:        "info",
		CacheMaxSizeMB:  128,
		CacheTTLSeconds: 300,
		CacheEnabled:    true,
		TLSPort:         8443,
	}

	if v := os.Getenv("READWISE_PROFILES"); v != "" {
		profiles := strings.Split(v, ",")
		c.Profiles = make([]string, 0, len(profiles))
		for _, p := range profiles {
			p = strings.TrimSpace(p)
			if p != "" {
				c.Profiles = append(c.Profiles, p)
			}
		}
	}

	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Port = n
		}
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = strings.ToLower(v)
	}

	if v := os.Getenv("CACHE_MAX_SIZE_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			c.CacheMaxSizeMB = n
		}
	}

	if v := os.Getenv("CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			c.CacheTTLSeconds = n
		}
	}

	if v := os.Getenv("CACHE_ENABLED"); v != "" {
		c.CacheEnabled = strings.ToLower(v) == "true" || v == "1"
	}

	if v := os.Getenv("TLS_CERT_FILE"); v != "" {
		c.TLSCertFile = v
	}

	if v := os.Getenv("TLS_KEY_FILE"); v != "" {
		c.TLSKeyFile = v
	}

	if v := os.Getenv("TLS_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.TLSPort = n
		}
	}

	return c
}

// TLSEnabled returns true when both TLS certificate and key files are configured.
func (c Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

// TLSAddr returns the listen address string for the TLS port.
func (c Config) TLSAddr() string {
	return fmt.Sprintf(":%d", c.TLSPort)
}

// ValidateTLS checks TLS configuration for consistency and file accessibility.
// Returns nil if TLS is not configured (both fields empty).
func (c Config) ValidateTLS() error {
	hasCert := c.TLSCertFile != ""
	hasKey := c.TLSKeyFile != ""

	if hasCert != hasKey {
		return ErrTLSIncomplete
	}

	if !c.TLSEnabled() {
		return nil
	}

	if _, err := os.Stat(c.TLSCertFile); err != nil {
		return fmt.Errorf("TLS certificate file not found: %s", c.TLSCertFile)
	}

	if _, err := os.Stat(c.TLSKeyFile); err != nil {
		return fmt.Errorf("TLS key file not found: %s", c.TLSKeyFile)
	}

	if c.TLSPort == c.Port {
		return fmt.Errorf("TLS port %d conflicts with HTTP port", c.TLSPort)
	}

	return nil
}

// PageResponse represents a page-number paginated API response.
type PageResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

// CursorResponse represents a cursor-based paginated API response.
type CursorResponse[T any] struct {
	Count          int    `json:"count"`
	NextPageCursor string `json:"nextPageCursor"`
	Results        []T    `json:"results"`
}

// Addr returns the listen address string for the configured port.
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}
