package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLRUGetPut(t *testing.T) {
	c := NewLRU(1024 * 1024) // 1MB

	c.Put(NewEntry("key1", []byte("value1"), time.Hour))
	c.Put(NewEntry("key2", []byte("value2"), time.Hour))

	e := c.Get("key1")
	if e == nil {
		t.Fatal("expected entry for key1")
	}
	if string(e.Data) != "value1" {
		t.Errorf("data = %q, want %q", string(e.Data), "value1")
	}
}

func TestLRUGetMiss(t *testing.T) {
	c := NewLRU(1024 * 1024)

	e := c.Get("nonexistent")
	if e != nil {
		t.Errorf("expected nil for nonexistent key, got %v", e)
	}
}

func TestLRUExpiredEntry(t *testing.T) {
	c := NewLRU(1024 * 1024)

	c.Put(NewEntry("key1", []byte("value1"), 10*time.Millisecond))
	time.Sleep(15 * time.Millisecond)

	e := c.Get("key1")
	if e != nil {
		t.Error("expected nil for expired entry")
	}
	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0 (expired entry should be removed)", c.Len())
	}
}

func TestLRUEvictOnSizeLimit(t *testing.T) {
	c := NewLRU(100) // 100 bytes limit

	// Add entries that exceed the limit
	// Need to bypass minEntryLifetime for testing
	e1 := NewEntry("key1", make([]byte, 40), time.Hour)
	e1.CreatedAt = time.Now().Add(-time.Minute) // make old enough to evict
	c.Put(e1)

	e2 := NewEntry("key2", make([]byte, 40), time.Hour)
	e2.CreatedAt = time.Now().Add(-time.Minute)
	c.Put(e2)

	e3 := NewEntry("key3", make([]byte, 40), time.Hour)
	c.Put(e3)

	// key1 should be evicted (LRU, oldest)
	if c.Get("key1") != nil {
		t.Error("key1 should have been evicted")
	}
	if c.Size() > 100 {
		t.Errorf("cache size = %d, should be <= 100", c.Size())
	}
}

func TestLRUUpdateExisting(t *testing.T) {
	c := NewLRU(1024 * 1024)

	c.Put(NewEntry("key1", []byte("old"), time.Hour))
	c.Put(NewEntry("key1", []byte("new"), time.Hour))

	e := c.Get("key1")
	if e == nil {
		t.Fatal("expected entry for key1")
	}
	if string(e.Data) != "new" {
		t.Errorf("data = %q, want %q", string(e.Data), "new")
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
}

func TestLRUDelete(t *testing.T) {
	c := NewLRU(1024 * 1024)

	c.Put(NewEntry("key1", []byte("value1"), time.Hour))
	c.Delete("key1")

	if c.Get("key1") != nil {
		t.Error("key1 should be deleted")
	}
	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
}

func TestLRUDeleteByPrefix(t *testing.T) {
	c := NewLRU(1024 * 1024)

	c.Put(NewEntry("user1|endpoint1|a", []byte("data"), time.Hour))
	c.Put(NewEntry("user1|endpoint1|b", []byte("data"), time.Hour))
	c.Put(NewEntry("user1|endpoint2|a", []byte("data"), time.Hour))
	c.Put(NewEntry("user2|endpoint1|a", []byte("data"), time.Hour))

	c.DeleteByPrefix("user1|endpoint1")

	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2 after prefix deletion", c.Len())
	}
	if c.Get("user1|endpoint1|a") != nil {
		t.Error("user1|endpoint1|a should be deleted")
	}
	if c.Get("user1|endpoint2|a") == nil {
		t.Error("user1|endpoint2|a should still exist")
	}
}

func TestLRUSizeTracking(t *testing.T) {
	c := NewLRU(1024 * 1024)

	c.Put(NewEntry("key1", make([]byte, 100), time.Hour))
	if c.Size() != 100 {
		t.Errorf("Size() = %d, want 100", c.Size())
	}

	c.Put(NewEntry("key2", make([]byte, 200), time.Hour))
	if c.Size() != 300 {
		t.Errorf("Size() = %d, want 300", c.Size())
	}

	c.Delete("key1")
	if c.Size() != 200 {
		t.Errorf("Size() = %d, want 200 after delete", c.Size())
	}
}

func TestLRUConcurrentAccess(t *testing.T) {
	c := NewLRU(1024 * 1024)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			c.Put(NewEntry(key, []byte(fmt.Sprintf("value%d", i)), time.Hour))
			c.Get(key)
		}(i)
	}

	wg.Wait()

	if c.Len() != 100 {
		t.Errorf("Len() = %d, want 100 after concurrent puts", c.Len())
	}
}

func TestLRUMinEntryLifetime(t *testing.T) {
	c := NewLRU(50) // very small limit

	// Add a young entry
	c.Put(NewEntry("young", make([]byte, 40), time.Hour))

	// Try to add another that would exceed the limit
	c.Put(NewEntry("new", make([]byte, 30), time.Hour))

	// Young entry should NOT be evicted despite exceeding limit
	if c.Get("young") == nil {
		t.Error("young entry should not be evicted (under minEntryLifetime)")
	}
}
