package tools

import (
	"testing"

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
