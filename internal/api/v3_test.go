package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func newTestV3Server(handler http.HandlerFunc) (*Client, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := NewClientWithBaseURLs(ts.URL, ts.URL)
	return client, ts
}

func TestListDocuments(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/" {
			t.Errorf("path = %q, want /list/", r.URL.Path)
		}

		resp := types.CursorResponse[types.Document]{
			Count: 2,
			Results: []types.Document{
				{ID: "doc1", Title: "Article One", Category: "article"},
				{ID: "doc2", Title: "Article Two", Category: "article"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	result, err := client.ListDocuments(context.Background(), "key", "", "", "", 0)
	if err != nil {
		t.Fatalf("ListDocuments error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(result.Results))
	}
	if result.Results[0].Title != "Article One" {
		t.Errorf("Results[0].Title = %q, want %q", result.Results[0].Title, "Article One")
	}
}

func TestListDocumentsWithFilters(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("location") != "later" {
			t.Errorf("location = %q, want later", r.URL.Query().Get("location"))
		}
		if r.URL.Query().Get("category") != "article" {
			t.Errorf("category = %q, want article", r.URL.Query().Get("category"))
		}
		json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{Count: 0, Results: []types.Document{}})
	})
	defer ts.Close()

	_, err := client.ListDocuments(context.Background(), "key", "later", "article", "", 0)
	if err != nil {
		t.Fatalf("ListDocuments error: %v", err)
	}
}

func TestListDocumentsPagination(t *testing.T) {
	callCount := 0
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			resp := types.CursorResponse[types.Document]{
				Count:          1,
				NextPageCursor: "cursor2",
				Results:        []types.Document{{ID: "doc1", Title: "First"}},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := types.CursorResponse[types.Document]{
				Count:   1,
				Results: []types.Document{{ID: "doc2", Title: "Second"}},
			}
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer ts.Close()

	result, err := client.ListDocuments(context.Background(), "key", "", "", "", 0)
	if err != nil {
		t.Fatalf("ListDocuments error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(result.Results))
	}
	if callCount != 2 {
		t.Errorf("API call count = %d, want 2", callCount)
	}
}

func TestListDocumentsLimit(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CursorResponse[types.Document]{
			Count:          3,
			NextPageCursor: "more",
			Results: []types.Document{
				{ID: "doc1"}, {ID: "doc2"}, {ID: "doc3"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	result, err := client.ListDocuments(context.Background(), "key", "", "", "", 2)
	if err != nil {
		t.Fatalf("ListDocuments error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(result.Results))
	}
}

func TestGetDocument(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/" {
			t.Errorf("path = %q, want /list/", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "doc123" {
			t.Errorf("id = %q, want doc123", r.URL.Query().Get("id"))
		}

		resp := types.CursorResponse[types.Document]{
			Count:   1,
			Results: []types.Document{{ID: "doc123", Title: "My Document"}},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	result, err := client.GetDocument(context.Background(), "key", "doc123", false)
	if err != nil {
		t.Fatalf("GetDocument error: %v", err)
	}
	if result.ID != "doc123" {
		t.Errorf("ID = %q, want %q", result.ID, "doc123")
	}
}

func TestGetDocumentWithContent(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("withHtmlContent") != "true" {
			t.Errorf("withHtmlContent = %q, want true", r.URL.Query().Get("withHtmlContent"))
		}

		resp := types.CursorResponse[types.Document]{
			Count:   1,
			Results: []types.Document{{ID: "doc1", Content: "<p>Hello</p>"}},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	result, err := client.GetDocument(context.Background(), "key", "doc1", true)
	if err != nil {
		t.Fatalf("GetDocument error: %v", err)
	}
	if result.Content != "<p>Hello</p>" {
		t.Errorf("Content = %q, want %q", result.Content, "<p>Hello</p>")
	}
}

func TestGetDocumentNotFound(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CursorResponse[types.Document]{Count: 0, Results: []types.Document{}}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	_, err := client.GetDocument(context.Background(), "key", "nonexistent", false)
	if err == nil {
		t.Fatal("expected error for nonexistent document")
	}
}

func TestListReaderTags(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags/" {
			t.Errorf("path = %q, want /tags/", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]types.Tag{{ID: 1, Name: "tech"}, {ID: 2, Name: "science"}})
	})
	defer ts.Close()

	result, err := client.ListReaderTags(context.Background(), "key")
	if err != nil {
		t.Fatalf("ListReaderTags error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
}

func TestSaveDocument(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/save/" {
			t.Errorf("path = %q, want /save/", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req types.SaveDocumentRequest
		json.Unmarshal(body, &req)

		if req.URL != "https://example.com/article" {
			t.Errorf("URL = %q, want %q", req.URL, "https://example.com/article")
		}
		if req.Title != "Custom Title" {
			t.Errorf("Title = %q, want %q", req.Title, "Custom Title")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(types.SaveDocumentResponse{
			ID:  "saved-doc-1",
			URL: "https://example.com/article",
		})
	})
	defer ts.Close()

	result, err := client.SaveDocument(context.Background(), "key", types.SaveDocumentRequest{
		URL:   "https://example.com/article",
		Title: "Custom Title",
	})
	if err != nil {
		t.Fatalf("SaveDocument error: %v", err)
	}
	if result.ID != "saved-doc-1" {
		t.Errorf("ID = %q, want %q", result.ID, "saved-doc-1")
	}
}

func TestSaveDocumentWithTags(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req types.SaveDocumentRequest
		json.Unmarshal(body, &req)

		if len(req.Tags) != 2 || req.Tags[0] != "tech" || req.Tags[1] != "go" {
			t.Errorf("Tags = %v, want [tech go]", req.Tags)
		}

		json.NewEncoder(w).Encode(types.SaveDocumentResponse{ID: "doc-2", URL: "https://example.com"})
	})
	defer ts.Close()

	_, err := client.SaveDocument(context.Background(), "key", types.SaveDocumentRequest{
		URL:  "https://example.com",
		Tags: []string{"tech", "go"},
	})
	if err != nil {
		t.Fatalf("SaveDocument error: %v", err)
	}
}

func TestUpdateDocument(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/update/doc-1/" {
			t.Errorf("path = %q, want /update/doc-1/", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req types.UpdateDocumentRequest
		json.Unmarshal(body, &req)

		if req.Title != "Updated Title" {
			t.Errorf("Title = %q, want %q", req.Title, "Updated Title")
		}
		if req.Location != "archive" {
			t.Errorf("Location = %q, want %q", req.Location, "archive")
		}

		json.NewEncoder(w).Encode(types.Document{ID: "doc-1", Title: "Updated Title", Location: "archive"})
	})
	defer ts.Close()

	result, err := client.UpdateDocument(context.Background(), "key", "doc-1", types.UpdateDocumentRequest{
		Title:    "Updated Title",
		Location: "archive",
	})
	if err != nil {
		t.Fatalf("UpdateDocument error: %v", err)
	}
	if result.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", result.Title, "Updated Title")
	}
}

func TestDeleteDocument(t *testing.T) {
	client, ts := newTestV3Server(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/delete/doc-1/" {
			t.Errorf("path = %q, want /delete/doc-1/", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := client.DeleteDocument(context.Background(), "key", "doc-1")
	if err != nil {
		t.Fatalf("DeleteDocument error: %v", err)
	}
}
