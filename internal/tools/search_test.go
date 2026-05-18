package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func TestSearchHighlightsExactMatch(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Test Book",
			Highlights: []types.Highlight{
				{ID: 1, Text: "The quick brown fox"},
				{ID: 2, Text: "Hello world"},
				{ID: 3, Text: "Another highlight"},
			},
		},
	}

	results := searchHighlights(sources, "Hello world", "", 50)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	// Exact match should score higher than partial
	if results[0].Highlight.ID != 2 {
		t.Errorf("expected exact match first, got ID %d", results[0].Highlight.ID)
	}
}

func TestSearchHighlightsCaseInsensitive(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Book",
			Highlights: []types.Highlight{
				{ID: 1, Text: "IMPORTANT NOTE"},
			},
		},
	}

	results := searchHighlights(sources, "important note", "", 50)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchHighlightsInNotes(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Book",
			Highlights: []types.Highlight{
				{ID: 1, Text: "some text", Note: "this is a key insight"},
			},
		},
	}

	results := searchHighlights(sources, "key insight", "", 50)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchHighlightsInSourceTitle(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Atomic Habits",
			Highlights: []types.Highlight{
				{ID: 1, Text: "some text"},
			},
		},
	}

	results := searchHighlights(sources, "Atomic Habits", "", 50)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchHighlightsSourceIDFilter(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Book 1",
			Highlights: []types.Highlight{
				{ID: 1, Text: "matching text"},
			},
		},
		{
			UserBookID: 2,
			Title:      "Book 2",
			Highlights: []types.Highlight{
				{ID: 2, Text: "also matching text"},
			},
		},
	}

	results := searchHighlights(sources, "matching", "1", 50)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Highlight.ID != 1 {
		t.Errorf("expected ID 1, got %d", results[0].Highlight.ID)
	}
}

func TestSearchHighlightsLimit(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Book",
			Highlights: []types.Highlight{
				{ID: 1, Text: "matching one"},
				{ID: 2, Text: "matching two"},
				{ID: 3, Text: "matching three"},
			},
		},
	}

	results := searchHighlights(sources, "matching", "", 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchHighlightsNoMatch(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Book",
			Highlights: []types.Highlight{
				{ID: 1, Text: "some text"},
			},
		},
	}

	results := searchHighlights(sources, "nonexistent query", "", 50)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchHighlightsRelevanceScoring(t *testing.T) {
	sources := []types.ExportSource{
		{
			UserBookID: 1,
			Title:      "Book",
			Highlights: []types.Highlight{
				{ID: 1, Text: "partial match of query terms"},
				{ID: 2, Text: "query"},  // exact match
			},
		},
	}

	results := searchHighlights(sources, "query", "", 50)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Highlight.ID != 2 {
		t.Errorf("exact match should rank first, got ID %d", results[0].Highlight.ID)
	}
	if results[0].RelevanceScore <= results[1].RelevanceScore {
		t.Errorf("exact match score (%f) should be > partial (%f)",
			results[0].RelevanceScore, results[1].RelevanceScore)
	}
}

func TestSearchDocumentsBasic(t *testing.T) {
	docs := []types.Document{
		{ID: "1", Title: "Go Programming", Author: "Rob Pike"},
		{ID: "2", Title: "Rust Programming", Author: "Mozilla"},
		{ID: "3", Title: "Python Basics", Author: "Guido"},
	}

	results := searchDocuments(docs, "Programming", "", "", 50)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchDocumentsLocationFilter(t *testing.T) {
	docs := []types.Document{
		{ID: "1", Title: "Go Programming", Location: "later"},
		{ID: "2", Title: "Rust Programming", Location: "archive"},
	}

	results := searchDocuments(docs, "Programming", "later", "", 50)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Document.ID != "1" {
		t.Errorf("expected document 1, got %s", results[0].Document.ID)
	}
}

func TestSearchDocumentsCategoryFilter(t *testing.T) {
	docs := []types.Document{
		{ID: "1", Title: "Go Article", Category: "article"},
		{ID: "2", Title: "Go Video", Category: "video"},
	}

	results := searchDocuments(docs, "Go", "", "article", 50)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchDocumentsInSummaryAndNotes(t *testing.T) {
	docs := []types.Document{
		{ID: "1", Title: "unrelated", Summary: "contains the search term"},
		{ID: "2", Title: "also unrelated", Notes: "has the search term here"},
	}

	results := searchDocuments(docs, "search term", "", "", 50)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func newSearchTestDeps(handler http.HandlerFunc) (*api.Client, *cache.Manager, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := api.NewClientWithBaseURLs(ts.URL, ts.URL)
	cm := cache.NewManager(16, 300, true)
	return client, cm, ts
}

func TestSearchDocumentsCacheHit(t *testing.T) {
	var requestCount atomic.Int32
	docs := types.CursorResponse[types.Document]{
		Count:   2,
		Results: []types.Document{
			{ID: "1", Title: "Go Programming", Author: "Rob Pike"},
			{ID: "2", Title: "Rust Programming", Author: "Mozilla"},
		},
	}

	client, cm, ts := newSearchTestDeps(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		json.NewEncoder(w).Encode(docs)
	})
	defer ts.Close()

	handler := makeSearchDocumentsHandler(client, cm)
	req := newReqWithAPIKey("test-key")

	result1, _, err := handler(context.Background(), req, SearchDocumentsInput{Query: "Programming"})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if result1 == nil {
		t.Fatal("first call: expected result")
	}
	firstRequestCount := requestCount.Load()
	if firstRequestCount == 0 {
		t.Fatal("expected at least one upstream request on first call")
	}

	result2, _, err := handler(context.Background(), req, SearchDocumentsInput{Query: "Programming"})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if result2 == nil {
		t.Fatal("second call: expected result")
	}

	if requestCount.Load() != firstRequestCount {
		t.Errorf("expected no additional upstream requests on cache hit, got %d total", requestCount.Load())
	}

	var results []SearchDocumentResult
	textContent := result2.Content[0].(*mcp.TextContent)
	if err := json.Unmarshal([]byte(textContent.Text), &results); err != nil {
		t.Fatalf("failed to unmarshal cached result: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results from cache, got %d", len(results))
	}
}

func TestSearchDocumentsCacheMiss(t *testing.T) {
	var requestCount atomic.Int32
	docs := types.CursorResponse[types.Document]{
		Count:   1,
		Results: []types.Document{
			{ID: "1", Title: "Go Programming", Author: "Rob Pike"},
		},
	}

	client, cm, ts := newSearchTestDeps(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		json.NewEncoder(w).Encode(docs)
	})
	defer ts.Close()

	handler := makeSearchDocumentsHandler(client, cm)
	req := newReqWithAPIKey("test-key")

	result, _, err := handler(context.Background(), req, SearchDocumentsInput{Query: "Go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if requestCount.Load() == 0 {
		t.Fatal("expected upstream requests on cache miss")
	}

	var results []SearchDocumentResult
	textContent := result.Content[0].(*mcp.TextContent)
	if err := json.Unmarshal([]byte(textContent.Text), &results); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Document.Title != "Go Programming" {
		t.Errorf("expected 'Go Programming', got %q", results[0].Document.Title)
	}
}

func TestSearchHighlightsCacheHit(t *testing.T) {
	var requestCount atomic.Int32
	exports := types.CursorResponse[types.ExportSource]{
		Count: 1,
		Results: []types.ExportSource{
			{
				UserBookID: 1,
				Title:      "Test Book",
				Highlights: []types.Highlight{
					{ID: 1, Text: "important insight about Go"},
					{ID: 2, Text: "another highlight"},
				},
			},
		},
	}

	client, cm, ts := newSearchTestDeps(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		json.NewEncoder(w).Encode(exports)
	})
	defer ts.Close()

	handler := makeSearchHighlightsHandler(client, cm)
	req := newReqWithAPIKey("test-key")

	result1, _, err := handler(context.Background(), req, SearchHighlightsInput{Query: "Go"})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if result1 == nil {
		t.Fatal("first call: expected result")
	}
	firstRequestCount := requestCount.Load()
	if firstRequestCount == 0 {
		t.Fatal("expected at least one upstream request on first call")
	}

	result2, _, err := handler(context.Background(), req, SearchHighlightsInput{Query: "Go"})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if result2 == nil {
		t.Fatal("second call: expected result")
	}

	if requestCount.Load() != firstRequestCount {
		t.Errorf("expected no additional upstream requests on cache hit, got %d total", requestCount.Load())
	}

	var results []SearchHighlightResult
	textContent := result2.Content[0].(*mcp.TextContent)
	if err := json.Unmarshal([]byte(textContent.Text), &results); err != nil {
		t.Fatalf("failed to unmarshal cached result: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result from cache, got %d", len(results))
	}
	if results[0].Highlight.Text != "important insight about Go" {
		t.Errorf("expected cached highlight text, got %q", results[0].Highlight.Text)
	}
}

func TestSearchHighlightsCacheMiss(t *testing.T) {
	var requestCount atomic.Int32
	exports := types.CursorResponse[types.ExportSource]{
		Count: 1,
		Results: []types.ExportSource{
			{
				UserBookID: 1,
				Title:      "Test Book",
				Highlights: []types.Highlight{
					{ID: 1, Text: "insight about Go"},
				},
			},
		},
	}

	client, cm, ts := newSearchTestDeps(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		json.NewEncoder(w).Encode(exports)
	})
	defer ts.Close()

	handler := makeSearchHighlightsHandler(client, cm)
	req := newReqWithAPIKey("test-key")

	result, _, err := handler(context.Background(), req, SearchHighlightsInput{Query: "Go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount.Load() == 0 {
		t.Fatal("expected upstream request on cache miss")
	}

	var results []SearchHighlightResult
	textContent := result.Content[0].(*mcp.TextContent)
	if err := json.Unmarshal([]byte(textContent.Text), &results); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Highlight.Text != "insight about Go" {
		t.Errorf("expected 'insight about Go', got %q", results[0].Highlight.Text)
	}
}
