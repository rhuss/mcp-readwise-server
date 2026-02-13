package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func newTestV2Server(handler http.HandlerFunc) (*Client, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	return client, ts
}

func TestListBooks(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/books/" {
			t.Errorf("path = %q, want /books/", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("page = %q, want 1", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("page_size = %q, want 10", r.URL.Query().Get("page_size"))
		}

		resp := types.PageResponse[types.Source]{
			Count: 2,
			Results: []types.Source{
				{ID: 1, Title: "Book One", Category: "books"},
				{ID: 2, Title: "Book Two", Category: "articles"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	result, err := client.ListBooks(context.Background(), "key", 1, 10, "", "")
	if err != nil {
		t.Fatalf("ListBooks error: %v", err)
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(result.Results))
	}
	if result.Results[0].Title != "Book One" {
		t.Errorf("Results[0].Title = %q, want %q", result.Results[0].Title, "Book One")
	}
}

func TestListBooksWithCategory(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("category") != "articles" {
			t.Errorf("category = %q, want articles", r.URL.Query().Get("category"))
		}
		json.NewEncoder(w).Encode(types.PageResponse[types.Source]{Count: 0, Results: []types.Source{}})
	})
	defer ts.Close()

	_, err := client.ListBooks(context.Background(), "key", 0, 0, "articles", "")
	if err != nil {
		t.Fatalf("ListBooks error: %v", err)
	}
}

func TestGetBook(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/books/123/" {
			t.Errorf("path = %q, want /books/123/", r.URL.Path)
		}
		json.NewEncoder(w).Encode(types.Source{ID: 123, Title: "Test Book"})
	})
	defer ts.Close()

	result, err := client.GetBook(context.Background(), "key", "123")
	if err != nil {
		t.Fatalf("GetBook error: %v", err)
	}
	if result.ID != 123 {
		t.Errorf("ID = %d, want 123", result.ID)
	}
}

func TestListHighlights(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/highlights/" {
			t.Errorf("path = %q, want /highlights/", r.URL.Path)
		}
		resp := types.PageResponse[types.Highlight]{
			Count:   1,
			Results: []types.Highlight{{ID: 42, Text: "test highlight", BookID: 1}},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	result, err := client.ListHighlights(context.Background(), "key", 1, 100, "", "")
	if err != nil {
		t.Fatalf("ListHighlights error: %v", err)
	}
	if result.Results[0].Text != "test highlight" {
		t.Errorf("Text = %q, want %q", result.Results[0].Text, "test highlight")
	}
}

func TestListHighlightsWithBookID(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("book_id") != "55" {
			t.Errorf("book_id = %q, want 55", r.URL.Query().Get("book_id"))
		}
		json.NewEncoder(w).Encode(types.PageResponse[types.Highlight]{Count: 0, Results: []types.Highlight{}})
	})
	defer ts.Close()

	_, err := client.ListHighlights(context.Background(), "key", 0, 0, "55", "")
	if err != nil {
		t.Fatalf("ListHighlights error: %v", err)
	}
}

func TestGetHighlight(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/highlights/42/" {
			t.Errorf("path = %q, want /highlights/42/", r.URL.Path)
		}
		json.NewEncoder(w).Encode(types.Highlight{ID: 42, Text: "my highlight"})
	})
	defer ts.Close()

	result, err := client.GetHighlight(context.Background(), "key", "42")
	if err != nil {
		t.Fatalf("GetHighlight error: %v", err)
	}
	if result.Text != "my highlight" {
		t.Errorf("Text = %q, want %q", result.Text, "my highlight")
	}
}

func TestExportHighlights(t *testing.T) {
	callCount := 0
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path != "/export/" {
			t.Errorf("path = %q, want /export/", r.URL.Path)
		}

		if callCount == 1 {
			// First page with a cursor
			resp := types.CursorResponse[types.ExportSource]{
				Count:          1,
				NextPageCursor: "cursor2",
				Results: []types.ExportSource{
					{UserBookID: 1, Title: "Source 1", Highlights: []types.Highlight{{ID: 10, Text: "h1"}}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			// Second page without cursor (final)
			resp := types.CursorResponse[types.ExportSource]{
				Count: 1,
				Results: []types.ExportSource{
					{UserBookID: 2, Title: "Source 2", Highlights: []types.Highlight{{ID: 20, Text: "h2"}}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer ts.Close()

	result, err := client.ExportHighlights(context.Background(), "key", "")
	if err != nil {
		t.Fatalf("ExportHighlights error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(result.Results))
	}
	if callCount != 2 {
		t.Errorf("API call count = %d, want 2", callCount)
	}
}

func TestExportHighlightsWithUpdatedAfter(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("updatedAfter") != "2024-01-01" {
			t.Errorf("updatedAfter = %q, want 2024-01-01", r.URL.Query().Get("updatedAfter"))
		}
		json.NewEncoder(w).Encode(types.CursorResponse[types.ExportSource]{Count: 0, Results: []types.ExportSource{}})
	})
	defer ts.Close()

	_, err := client.ExportHighlights(context.Background(), "key", "2024-01-01")
	if err != nil {
		t.Fatalf("ExportHighlights error: %v", err)
	}
}

func TestGetDailyReview(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/review/" {
			t.Errorf("path = %q, want /review/", r.URL.Path)
		}
		json.NewEncoder(w).Encode(types.DailyReview{
			ReviewID:  789,
			ReviewURL: "https://readwise.io/review/789",
			Highlights: []types.ReviewHighlight{
				{ID: 1, Text: "review highlight"},
			},
		})
	})
	defer ts.Close()

	result, err := client.GetDailyReview(context.Background(), "key")
	if err != nil {
		t.Fatalf("GetDailyReview error: %v", err)
	}
	if result.ReviewID != 789 {
		t.Errorf("ReviewID = %d, want 789", result.ReviewID)
	}
	if len(result.Highlights) != 1 {
		t.Fatalf("len(Highlights) = %d, want 1", len(result.Highlights))
	}
}

func TestListBookTags(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/books/5/tags" {
			t.Errorf("path = %q, want /books/5/tags", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]types.Tag{{ID: 1, Name: "science"}, {ID: 2, Name: "physics"}})
	})
	defer ts.Close()

	result, err := client.ListBookTags(context.Background(), "key", "5")
	if err != nil {
		t.Fatalf("ListBookTags error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].Name != "science" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "science")
	}
}

func TestListHighlightTags(t *testing.T) {
	client, ts := newTestV2Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/highlights/42/tags" {
			t.Errorf("path = %q, want /highlights/42/tags", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]types.Tag{{ID: 10, Name: "important"}})
	})
	defer ts.Close()

	result, err := client.ListHighlightTags(context.Background(), "key", "42")
	if err != nil {
		t.Fatalf("ListHighlightTags error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
	if result[0].Name != "important" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "important")
	}
}
