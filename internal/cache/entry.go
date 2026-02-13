package cache

import (
	"sync"
	"time"
)

// Entry represents a cached API response with TTL and size tracking.
type Entry struct {
	Key            string
	Data           []byte
	Size           int64
	CreatedAt      time.Time
	TTL            time.Duration
	LastAccessedAt time.Time
	mu             sync.Mutex
}

// NewEntry creates a new cache entry.
func NewEntry(key string, data []byte, ttl time.Duration) *Entry {
	now := time.Now()
	return &Entry{
		Key:            key,
		Data:           data,
		Size:           int64(len(data)),
		CreatedAt:      now,
		TTL:            ttl,
		LastAccessedAt: now,
	}
}

// IsExpired returns true if the entry has exceeded its TTL.
func (e *Entry) IsExpired() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return time.Since(e.CreatedAt) > e.TTL
}

// Touch updates the last access time.
func (e *Entry) Touch() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.LastAccessedAt = time.Now()
}

// Age returns the time since the entry was created.
func (e *Entry) Age() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	return time.Since(e.CreatedAt)
}
