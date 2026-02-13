package cache

import (
	"container/list"
	"sync"
	"time"
)

const minEntryLifetime = 30 * time.Second

// LRU implements a thread-safe least-recently-used cache with size tracking.
type LRU struct {
	mu       sync.RWMutex
	maxSize  int64 // maximum total size in bytes
	size     int64 // current total size
	items    map[string]*list.Element
	order    *list.List // front = most recently used
}

// NewLRU creates a new LRU cache with the given maximum size in bytes.
func NewLRU(maxSizeBytes int64) *LRU {
	return &LRU{
		maxSize: maxSizeBytes,
		items:   make(map[string]*list.Element),
		order:   list.New(),
	}
}

// Get retrieves an entry by key. Returns nil if not found or expired.
func (c *LRU) Get(key string) *Entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil
	}

	entry := elem.Value.(*Entry)
	if entry.IsExpired() {
		c.removeLocked(elem)
		return nil
	}

	entry.Touch()
	c.order.MoveToFront(elem)
	return entry
}

// Put adds or updates an entry in the cache. Evicts LRU entries if over size limit.
func (c *LRU) Put(entry *Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if elem, ok := c.items[entry.Key]; ok {
		old := elem.Value.(*Entry)
		c.size -= old.Size
		elem.Value = entry
		c.size += entry.Size
		c.order.MoveToFront(elem)
	} else {
		// Add new entry
		elem := c.order.PushFront(entry)
		c.items[entry.Key] = elem
		c.size += entry.Size
	}

	// Evict LRU entries if over size limit
	c.evictLocked()
}

// Delete removes an entry by key.
func (c *LRU) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.removeLocked(elem)
	}
}

// DeleteByPrefix removes all entries whose keys start with the given prefix.
func (c *LRU) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, elem := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			c.removeLocked(elem)
		}
	}
}

// Size returns the current total size of all cached entries in bytes.
func (c *LRU) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// Len returns the number of entries in the cache.
func (c *LRU) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *LRU) removeLocked(elem *list.Element) {
	entry := elem.Value.(*Entry)
	c.size -= entry.Size
	delete(c.items, entry.Key)
	c.order.Remove(elem)
}

func (c *LRU) evictLocked() {
	for c.size > c.maxSize && c.order.Len() > 0 {
		// Evict from back (least recently used)
		elem := c.order.Back()
		if elem == nil {
			break
		}
		entry := elem.Value.(*Entry)

		// Don't evict entries younger than minEntryLifetime
		if entry.Age() < minEntryLifetime {
			break
		}

		c.removeLocked(elem)
	}
}
