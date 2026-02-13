package types

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds server configuration loaded from environment variables.
type Config struct {
	Profiles       []string
	Port           int
	LogLevel       string
	CacheMaxSizeMB int
	CacheTTLSeconds int
	CacheEnabled   bool
}

// LoadConfig reads configuration from environment variables with defaults.
func LoadConfig() Config {
	c := Config{
		Profiles:       []string{"readwise"},
		Port:           8080,
		LogLevel:       "info",
		CacheMaxSizeMB: 128,
		CacheTTLSeconds: 300,
		CacheEnabled:   true,
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

	return c
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
