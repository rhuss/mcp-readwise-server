package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func TestCreateHighlight(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/highlights/" {
			t.Errorf("path = %q, want /highlights/", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req types.CreateHighlightsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if len(req.Highlights) != 1 {
			t.Fatalf("expected 1 highlight, got %d", len(req.Highlights))
		}
		if req.Highlights[0].Text != "test highlight" {
			t.Errorf("text = %q, want %q", req.Highlights[0].Text, "test highlight")
		}

		json.NewEncoder(w).Encode([]types.Highlight{
			{ID: 100, Text: "test highlight", BookID: 1},
		})
	})
	defer ts.Close()

	results, err := client.CreateHighlight(context.Background(), "key", types.CreateHighlightsRequest{
		Highlights: []types.CreateHighlightRequest{
			{Text: "test highlight", Title: "My Book"},
		},
	})
	if err != nil {
		t.Fatalf("CreateHighlight error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].ID != 100 {
		t.Errorf("ID = %d, want 100", results[0].ID)
	}
}

func TestCreateHighlightBatch(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req types.CreateHighlightsRequest
		json.Unmarshal(body, &req)

		if len(req.Highlights) != 3 {
			t.Errorf("expected 3 highlights, got %d", len(req.Highlights))
		}

		results := make([]types.Highlight, len(req.Highlights))
		for i, h := range req.Highlights {
			results[i] = types.Highlight{ID: int64(i + 1), Text: h.Text}
		}
		json.NewEncoder(w).Encode(results)
	})
	defer ts.Close()

	results, err := client.CreateHighlight(context.Background(), "key", types.CreateHighlightsRequest{
		Highlights: []types.CreateHighlightRequest{
			{Text: "highlight 1", Title: "Book"},
			{Text: "highlight 2", Title: "Book"},
			{Text: "highlight 3", Title: "Book"},
		},
	})
	if err != nil {
		t.Fatalf("CreateHighlight error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
}

func TestUpdateHighlight(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/highlights/42/" {
			t.Errorf("path = %q, want /highlights/42/", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req types.UpdateHighlightRequest
		json.Unmarshal(body, &req)

		if req.Text != "updated text" {
			t.Errorf("text = %q, want %q", req.Text, "updated text")
		}

		json.NewEncoder(w).Encode(types.Highlight{ID: 42, Text: "updated text", Note: "a note"})
	})
	defer ts.Close()

	result, err := client.UpdateHighlight(context.Background(), "key", "42", types.UpdateHighlightRequest{
		Text: "updated text",
	})
	if err != nil {
		t.Fatalf("UpdateHighlight error: %v", err)
	}
	if result.Text != "updated text" {
		t.Errorf("Text = %q, want %q", result.Text, "updated text")
	}
}

func TestAddBookTag(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/books/5/tags/" {
			t.Errorf("path = %q, want /books/5/tags/", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req types.CreateTagRequest
		json.Unmarshal(body, &req)

		if req.Name != "science" {
			t.Errorf("name = %q, want %q", req.Name, "science")
		}

		json.NewEncoder(w).Encode(types.Tag{ID: 10, Name: "science"})
	})
	defer ts.Close()

	result, err := client.AddBookTag(context.Background(), "key", "5", types.CreateTagRequest{Name: "science"})
	if err != nil {
		t.Fatalf("AddBookTag error: %v", err)
	}
	if result.Name != "science" {
		t.Errorf("Name = %q, want %q", result.Name, "science")
	}
}

func TestAddHighlightTag(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/highlights/42/tags/" {
			t.Errorf("path = %q, want /highlights/42/tags/", r.URL.Path)
		}

		json.NewEncoder(w).Encode(types.Tag{ID: 20, Name: "key-insight"})
	})
	defer ts.Close()

	result, err := client.AddHighlightTag(context.Background(), "key", "42", types.CreateTagRequest{Name: "key-insight"})
	if err != nil {
		t.Fatalf("AddHighlightTag error: %v", err)
	}
	if result.Name != "key-insight" {
		t.Errorf("Name = %q, want %q", result.Name, "key-insight")
	}
}

func TestDeleteHighlight(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/highlights/42/" {
			t.Errorf("path = %q, want /highlights/42/", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := client.DeleteHighlight(context.Background(), "key", "42")
	if err != nil {
		t.Fatalf("DeleteHighlight error: %v", err)
	}
}

func TestDeleteBookTag(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/books/5/tags/10" {
			t.Errorf("path = %q, want /books/5/tags/10", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := client.DeleteBookTag(context.Background(), "key", "5", "10")
	if err != nil {
		t.Fatalf("DeleteBookTag error: %v", err)
	}
}

func TestDeleteHighlightTag(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/highlights/42/tags/20" {
			t.Errorf("path = %q, want /highlights/42/tags/20", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := client.DeleteHighlightTag(context.Background(), "key", "42", "20")
	if err != nil {
		t.Fatalf("DeleteHighlightTag error: %v", err)
	}
}
