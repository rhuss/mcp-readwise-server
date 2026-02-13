package cache

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Default TTLs for different endpoints.
var defaultTTLs = map[string]time.Duration{
	"/api/v2/export/": 5 * time.Minute,
	"/api/v2/books/":  5 * time.Minute,
	"/api/v3/list/":   5 * time.Minute,
	"/api/v3/tags/":   10 * time.Minute,
}

// invalidationMap maps write/destructive operations to cached endpoints to invalidate.
var invalidationMap = map[string][]string{
	"create_highlight":       {"/api/v2/export/"},
	"update_highlight":       {"/api/v2/export/"},
	"delete_highlight":       {"/api/v2/export/"},
	"bulk_create_highlights": {"/api/v2/export/"},
	"add_source_tag":         {"/api/v2/books/"},
	"delete_source_tag":      {"/api/v2/books/"},
	"add_highlight_tag":      {"/api/v2/export/"},
	"delete_highlight_tag":   {"/api/v2/export/"},
	"save_document":          {"/api/v3/list/"},
	"update_document":        {"/api/v3/list/"},
	"delete_document":        {"/api/v3/list/"},
}

// Manager orchestrates cache operations with per-user isolation
// and endpoint-specific TTLs.
type Manager struct {
	cache      *LRU
	enabled    bool
	defaultTTL time.Duration
}

// NewManager creates a new cache manager.
func NewManager(maxSizeMB int, defaultTTLSeconds int, enabled bool) *Manager {
	maxSizeBytes := int64(maxSizeMB) * 1024 * 1024
	return &Manager{
		cache:      NewLRU(maxSizeBytes),
		enabled:    enabled,
		defaultTTL: time.Duration(defaultTTLSeconds) * time.Second,
	}
}

// buildKey constructs a cache key from the API key hash, endpoint, and query parameters.
// The key preserves the prefix structure (apiKeyHash|endpoint|...) so that
// DeleteByPrefix can invalidate all entries for a user+endpoint combination.
func buildKey(apiKeyHash, endpoint string, params map[string]string) string {
	// Sort params for deterministic keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	paramStr := strings.Join(parts, "&")
	if paramStr != "" {
		return apiKeyHash + "|" + endpoint + "|" + paramStr
	}
	return apiKeyHash + "|" + endpoint
}

// HashAPIKey returns the SHA-256 hash of an API key for cache key construction.
func HashAPIKey(apiKey string) string {
	h := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", h)
}

// userEndpointPrefix returns a prefix for finding all cache entries
// for a specific user and endpoint combination.
func userEndpointPrefix(apiKeyHash, endpoint string) string {
	return apiKeyHash + "|" + endpoint
}

// Get retrieves a cached response for the given parameters.
// Returns nil if caching is disabled, the entry doesn't exist, or it's expired.
func (m *Manager) Get(apiKey, endpoint string, params map[string]string) []byte {
	if !m.enabled {
		return nil
	}
	keyHash := HashAPIKey(apiKey)
	key := buildKey(keyHash, endpoint, params)
	entry := m.cache.Get(key)
	if entry == nil {
		return nil
	}
	return entry.Data
}

// Put stores a response in the cache with endpoint-specific TTL.
func (m *Manager) Put(apiKey, endpoint string, params map[string]string, data []byte) {
	if !m.enabled {
		return
	}
	keyHash := HashAPIKey(apiKey)
	key := buildKey(keyHash, endpoint, params)

	ttl := m.defaultTTL
	if t, ok := defaultTTLs[endpoint]; ok {
		ttl = t
	}

	m.cache.Put(NewEntry(key, data, ttl))
}

// Invalidate removes all cached entries for the given user and operation.
func (m *Manager) Invalidate(apiKey string, operation string) {
	if !m.enabled {
		return
	}
	endpoints, ok := invalidationMap[operation]
	if !ok {
		return
	}

	keyHash := HashAPIKey(apiKey)
	for _, endpoint := range endpoints {
		prefix := userEndpointPrefix(keyHash, endpoint)
		m.cache.DeleteByPrefix(prefix)
	}
}

// InvalidateEndpoint removes all cached entries for a specific user and endpoint.
func (m *Manager) InvalidateEndpoint(apiKey, endpoint string) {
	if !m.enabled {
		return
	}
	keyHash := HashAPIKey(apiKey)
	prefix := userEndpointPrefix(keyHash, endpoint)
	m.cache.DeleteByPrefix(prefix)
}

// TotalSize returns the total cache size in bytes.
func (m *Manager) TotalSize() int64 {
	return m.cache.Size()
}

// Len returns the number of cached entries.
func (m *Manager) Len() int {
	return m.cache.Len()
}

// Enabled returns whether the cache is enabled.
func (m *Manager) Enabled() bool {
	return m.enabled
}
