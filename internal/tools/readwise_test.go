package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func TestListHighlightsWithULID(t *testing.T) {
	ulidID := "01krk43zgd57fcsj3tg7xn3skz"
	docURL := "https://example.com/article"
	docTitle := "Test Article"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/list/" && r.URL.Query().Get("id") == ulidID:
			json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{
				Count: 1,
				Results: []types.Document{
					{
						ID:        ulidID,
						Title:     docTitle,
						SourceURL: docURL,
					},
				},
			})
		case r.URL.Path == "/export/":
			json.NewEncoder(w).Encode(types.CursorResponse[types.ExportSource]{
				Count: 2,
				Results: []types.ExportSource{
					{
						UserBookID: 100,
						Title:      docTitle,
						SourceURL:  docURL,
						Highlights: []types.Highlight{
							{ID: 1, Text: "highlight one"},
							{ID: 2, Text: "highlight two"},
						},
					},
					{
						UserBookID: 200,
						Title:      "Other Book",
						SourceURL:  "https://other.com",
						Highlights: []types.Highlight{
							{ID: 3, Text: "other highlight"},
						},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := api.NewClientWithBaseURLs(ts.URL, ts.URL)
	cm := cache.NewManager(16, 300, true)
	handler := makeListHighlightsHandler(client, cm)
	req := newReqWithAPIKey("test-key")

	result, _, err := handler(context.Background(), req, ListHighlightsInput{SourceID: ulidID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result with content")
	}

	textContent := result.Content[0].(*mcp.TextContent)
	var highlights []types.Highlight
	if err := json.Unmarshal([]byte(textContent.Text), &highlights); err != nil {
		t.Fatalf("failed to unmarshal highlights: %v", err)
	}
	if len(highlights) != 2 {
		t.Fatalf("expected 2 highlights, got %d", len(highlights))
	}
}

func TestListHighlightsWithNumericID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/highlights/" && r.URL.Query().Get("book_id") == "42" {
			json.NewEncoder(w).Encode(types.PageResponse[types.Highlight]{
				Count: 1,
				Results: []types.Highlight{
					{ID: 1, Text: "numeric highlight"},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := api.NewClientWithBaseURLs(ts.URL, ts.URL)
	cm := cache.NewManager(16, 300, true)
	handler := makeListHighlightsHandler(client, cm)
	req := newReqWithAPIKey("test-key")

	result, _, err := handler(context.Background(), req, ListHighlightsInput{SourceID: "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result with content")
	}

	textContent := result.Content[0].(*mcp.TextContent)
	var hlResult types.PageResponse[types.Highlight]
	if err := json.Unmarshal([]byte(textContent.Text), &hlResult); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(hlResult.Results) != 1 {
		t.Fatalf("expected 1 highlight, got %d", len(hlResult.Results))
	}
	if hlResult.Results[0].Text != "numeric highlight" {
		t.Errorf("expected 'numeric highlight', got %q", hlResult.Results[0].Text)
	}
}

func TestIsNumericID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"", false},
		{"42", true},
		{"0", true},
		{"12345678", true},
		{"abc", false},
		{"12abc", false},
		{"01krk43zgd57fcsj3tg7xn3skz", false},
		{"-1", false},
	}
	for _, tt := range tests {
		if got := isNumericID(tt.id); got != tt.want {
			t.Errorf("isNumericID(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestMatchesDocument(t *testing.T) {
	source := types.ExportSource{Title: "Test Book", Author: "Author A", SourceURL: "https://example.com/article"}

	if !matchesDocument(source, &types.Document{SourceURL: "https://example.com/article"}) {
		t.Error("should match by URL")
	}
	if !matchesDocument(source, &types.Document{Title: "Test Book"}) {
		t.Error("should match by title when author is empty")
	}
	if !matchesDocument(source, &types.Document{Title: "Test Book", Author: "Author A"}) {
		t.Error("should match by title and author")
	}
	if matchesDocument(source, &types.Document{Title: "Test Book", Author: "Different Author"}) {
		t.Error("should not match by title when authors differ")
	}
	if matchesDocument(source, &types.Document{Title: "Other", SourceURL: "https://other.com"}) {
		t.Error("should not match different title and URL")
	}
}
