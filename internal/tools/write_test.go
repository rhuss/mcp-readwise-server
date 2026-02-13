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

func newWriteTestDeps(handler http.HandlerFunc) (*api.Client, *cache.Manager, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := api.NewClientWithBaseURLs(ts.URL, ts.URL)
	cm := cache.NewManager(16, 300, true)
	return client, cm, ts
}

func newReqWithAPIKey(apiKey string) *mcp.CallToolRequest {
	return &mcp.CallToolRequest{
		Extra: &mcp.RequestExtra{
			Header: http.Header{
				"Authorization": []string{"Token " + apiKey},
			},
		},
	}
}

func TestSaveDocumentHandlerMissingURL(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeSaveDocumentHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), SaveDocumentInput{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if err.Error() != "url is required" {
		t.Errorf("error = %q, want %q", err.Error(), "url is required")
	}
}

func TestSaveDocumentHandlerMissingAPIKey(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeSaveDocumentHandler(client, cm)
	req := &mcp.CallToolRequest{}
	_, _, err := handler(context.Background(), req, SaveDocumentInput{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestSaveDocumentHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SaveDocumentResponse{
			ID:  "saved-1",
			URL: "https://example.com",
		})
	})
	defer ts.Close()

	handler := makeSaveDocumentHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), SaveDocumentInput{
		URL:   "https://example.com",
		Title: "Test Article",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestUpdateDocumentHandlerMissingID(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeUpdateDocumentHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateDocumentInput{})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateDocumentHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.Document{ID: "doc-1", Title: "Updated"})
	})
	defer ts.Close()

	handler := makeUpdateDocumentHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateDocumentInput{
		ID:    "doc-1",
		Title: "Updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestCreateHighlightHandlerMissingText(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeCreateHighlightHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateHighlightInput{})
	if err == nil {
		t.Fatal("expected error for missing text")
	}
}

func TestCreateHighlightHandlerTextTooLong(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeCreateHighlightHandler(client, cm)
	longText := make([]byte, 8192)
	for i := range longText {
		longText[i] = 'a'
	}
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateHighlightInput{
		Text: string(longText),
	})
	if err == nil {
		t.Fatal("expected error for text too long")
	}
}

func TestCreateHighlightHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]types.Highlight{{ID: 1, Text: "test"}})
	})
	defer ts.Close()

	handler := makeCreateHighlightHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateHighlightInput{
		Text:        "test",
		SourceTitle: "My Book",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestCreateHighlightHandlerWithSourceID(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]types.Highlight{{ID: 1, Text: "test", BookID: 42}})
	})
	defer ts.Close()

	handler := makeCreateHighlightHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateHighlightInput{
		Text:     "test",
		SourceID: "42",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestUpdateHighlightHandlerMissingID(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeUpdateHighlightHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateHighlightInput{})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateHighlightHandlerTextTooLong(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeUpdateHighlightHandler(client, cm)
	longText := make([]byte, 8192)
	for i := range longText {
		longText[i] = 'a'
	}
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateHighlightInput{
		ID:   "1",
		Text: string(longText),
	})
	if err == nil {
		t.Fatal("expected error for text too long")
	}
}

func TestAddSourceTagHandlerMissingFields(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeAddSourceTagHandler(client, cm)

	// Missing both
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), AddSourceTagInput{})
	if err == nil {
		t.Fatal("expected error for missing fields")
	}

	// Missing name
	_, _, err = handler(context.Background(), newReqWithAPIKey("test-key"), AddSourceTagInput{SourceID: "1"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestAddHighlightTagHandlerMissingFields(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeAddHighlightTagHandler(client, cm)

	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), AddHighlightTagInput{})
	if err == nil {
		t.Fatal("expected error for missing fields")
	}
}

func TestBulkCreateHighlightsHandlerEmpty(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeBulkCreateHighlightsHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), BulkCreateHighlightsInput{
		Highlights: []BulkHighlightItem{},
	})
	if err == nil {
		t.Fatal("expected error for empty highlights")
	}
}

func TestBulkCreateHighlightsHandlerMissingText(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeBulkCreateHighlightsHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), BulkCreateHighlightsInput{
		Highlights: []BulkHighlightItem{
			{Text: "", SourceTitle: "Book"},
		},
	})
	if err == nil {
		t.Fatal("expected error for missing text")
	}
}

func TestBulkCreateHighlightsHandlerMissingSourceTitle(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeBulkCreateHighlightsHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), BulkCreateHighlightsInput{
		Highlights: []BulkHighlightItem{
			{Text: "some text", SourceTitle: ""},
		},
	})
	if err == nil {
		t.Fatal("expected error for missing source_title")
	}
}

func TestBulkCreateHighlightsHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]types.Highlight{
			{ID: 1, Text: "h1"},
			{ID: 2, Text: "h2"},
		})
	})
	defer ts.Close()

	handler := makeBulkCreateHighlightsHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), BulkCreateHighlightsInput{
		Highlights: []BulkHighlightItem{
			{Text: "h1", SourceTitle: "Book 1"},
			{Text: "h2", SourceTitle: "Book 2"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestSaveDocumentCacheInvalidation(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{
				Count:   1,
				Results: []types.Document{{ID: "doc-1", Title: "Cached"}},
			})
		} else {
			json.NewEncoder(w).Encode(types.SaveDocumentResponse{ID: "new-doc", URL: "https://example.com"})
		}
	})
	defer ts.Close()

	apiKey := "test-key"

	// Populate cache
	cm.Put(apiKey, "/api/v3/list/", nil, []byte(`cached data`))
	if cm.Len() != 1 {
		t.Fatalf("expected 1 cache entry, got %d", cm.Len())
	}

	// Save a document, which should invalidate the cache
	handler := makeSaveDocumentHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey(apiKey), SaveDocumentInput{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cache should be invalidated
	if cm.Len() != 0 {
		t.Errorf("expected cache to be invalidated, got %d entries", cm.Len())
	}
}
