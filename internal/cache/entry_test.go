package cache

import (
	"testing"
	"time"
)

func TestNewEntry(t *testing.T) {
	data := []byte("test data")
	e := NewEntry("key1", data, 5*time.Minute)

	if e.Key != "key1" {
		t.Errorf("Key = %q, want %q", e.Key, "key1")
	}
	if e.Size != int64(len(data)) {
		t.Errorf("Size = %d, want %d", e.Size, len(data))
	}
	if e.TTL != 5*time.Minute {
		t.Errorf("TTL = %v, want %v", e.TTL, 5*time.Minute)
	}
}

func TestEntryIsExpired(t *testing.T) {
	e := NewEntry("key", []byte("data"), 10*time.Millisecond)

	if e.IsExpired() {
		t.Error("entry should not be expired immediately")
	}

	time.Sleep(15 * time.Millisecond)

	if !e.IsExpired() {
		t.Error("entry should be expired after TTL")
	}
}

func TestEntryTouch(t *testing.T) {
	e := NewEntry("key", []byte("data"), time.Hour)
	initial := e.LastAccessedAt

	time.Sleep(time.Millisecond)
	e.Touch()

	if !e.LastAccessedAt.After(initial) {
		t.Error("Touch() should update LastAccessedAt")
	}
}

func TestEntryAge(t *testing.T) {
	e := NewEntry("key", []byte("data"), time.Hour)
	time.Sleep(5 * time.Millisecond)
	age := e.Age()

	if age < 5*time.Millisecond {
		t.Errorf("Age() = %v, expected >= 5ms", age)
	}
}
