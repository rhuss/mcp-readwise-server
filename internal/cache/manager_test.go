package cache

import (
	"testing"
)

func TestManagerGetPut(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("api-key-1", "/api/v2/books/", nil, []byte("books data"))
	data := m.Get("api-key-1", "/api/v2/books/", nil)

	if data == nil {
		t.Fatal("expected cached data")
	}
	if string(data) != "books data" {
		t.Errorf("data = %q, want %q", string(data), "books data")
	}
}

func TestManagerPerUserIsolation(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("user1-key", "/api/v2/books/", nil, []byte("user1 books"))
	m.Put("user2-key", "/api/v2/books/", nil, []byte("user2 books"))

	data1 := m.Get("user1-key", "/api/v2/books/", nil)
	data2 := m.Get("user2-key", "/api/v2/books/", nil)

	if string(data1) != "user1 books" {
		t.Errorf("user1 data = %q, want %q", string(data1), "user1 books")
	}
	if string(data2) != "user2 books" {
		t.Errorf("user2 data = %q, want %q", string(data2), "user2 books")
	}
}

func TestManagerInvalidation(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("api-key", "/api/v2/export/", nil, []byte("export data"))

	// Creating a highlight should invalidate export cache
	m.Invalidate("api-key", "create_highlight")

	data := m.Get("api-key", "/api/v2/export/", nil)
	if data != nil {
		t.Error("export cache should be invalidated after create_highlight")
	}
}

func TestManagerInvalidationBooks(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("api-key", "/api/v2/books/", nil, []byte("books data"))

	m.Invalidate("api-key", "add_source_tag")

	data := m.Get("api-key", "/api/v2/books/", nil)
	if data != nil {
		t.Error("books cache should be invalidated after add_source_tag")
	}
}

func TestManagerInvalidationDocuments(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("api-key", "/api/v3/list/", nil, []byte("documents data"))

	m.Invalidate("api-key", "save_document")

	data := m.Get("api-key", "/api/v3/list/", nil)
	if data != nil {
		t.Error("documents cache should be invalidated after save_document")
	}
}

func TestManagerInvalidationDoesNotAffectOtherUsers(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("user1-key", "/api/v2/export/", nil, []byte("user1 export"))
	m.Put("user2-key", "/api/v2/export/", nil, []byte("user2 export"))

	m.Invalidate("user1-key", "create_highlight")

	if m.Get("user1-key", "/api/v2/export/", nil) != nil {
		t.Error("user1 export should be invalidated")
	}
	if m.Get("user2-key", "/api/v2/export/", nil) == nil {
		t.Error("user2 export should NOT be invalidated")
	}
}

func TestManagerDisabled(t *testing.T) {
	m := NewManager(1, 300, false)

	m.Put("api-key", "/api/v2/books/", nil, []byte("data"))
	data := m.Get("api-key", "/api/v2/books/", nil)

	if data != nil {
		t.Error("disabled cache should return nil")
	}
}

func TestManagerWithParams(t *testing.T) {
	m := NewManager(1, 300, true)

	params1 := map[string]string{"page": "1", "category": "books"}
	params2 := map[string]string{"page": "2", "category": "books"}

	m.Put("api-key", "/api/v2/books/", params1, []byte("page1"))
	m.Put("api-key", "/api/v2/books/", params2, []byte("page2"))

	data1 := m.Get("api-key", "/api/v2/books/", params1)
	data2 := m.Get("api-key", "/api/v2/books/", params2)

	if string(data1) != "page1" {
		t.Errorf("page1 data = %q, want %q", string(data1), "page1")
	}
	if string(data2) != "page2" {
		t.Errorf("page2 data = %q, want %q", string(data2), "page2")
	}
}

func TestHashAPIKey(t *testing.T) {
	hash1 := HashAPIKey("key1")
	hash2 := HashAPIKey("key2")
	hash1again := HashAPIKey("key1")

	if hash1 == hash2 {
		t.Error("different keys should produce different hashes")
	}
	if hash1 != hash1again {
		t.Error("same key should produce same hash")
	}
	if len(hash1) != 64 {
		t.Errorf("SHA-256 hash should be 64 hex chars, got %d", len(hash1))
	}
}

func TestManagerTotalSize(t *testing.T) {
	m := NewManager(1, 300, true)

	m.Put("key", "/api/v2/books/", nil, make([]byte, 100))
	if m.TotalSize() != 100 {
		t.Errorf("TotalSize() = %d, want 100", m.TotalSize())
	}
}
