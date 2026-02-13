package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func TestListVideosHandler(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("category") != "video" {
			t.Errorf("category = %q, want video", r.URL.Query().Get("category"))
		}
		json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{
			Count: 2,
			Results: []types.Document{
				{ID: "v1", Title: "Video One", Category: "video"},
				{ID: "v2", Title: "Video Two", Category: "video"},
			},
		})
	})
	defer ts.Close()

	handler := makeListVideosHandler(client)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), ListVideosInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	// Verify the response contains the videos
	textContent := result.Content[0].(*mcp.TextContent)
	if textContent == nil {
		t.Fatal("expected text content")
	}
}

func TestListVideosHandlerWithLocation(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("location") != "later" {
			t.Errorf("location = %q, want later", r.URL.Query().Get("location"))
		}
		if r.URL.Query().Get("category") != "video" {
			t.Errorf("category = %q, want video", r.URL.Query().Get("category"))
		}
		json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{
			Count: 0, Results: []types.Document{},
		})
	})
	defer ts.Close()

	handler := makeListVideosHandler(client)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), ListVideosInput{
		Location: "later",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListVideosHandlerMissingAPIKey(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeListVideosHandler(client)
	_, _, err := handler(context.Background(), &mcp.CallToolRequest{}, ListVideosInput{})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestGetVideoHandler(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "v1" {
			t.Errorf("id = %q, want v1", r.URL.Query().Get("id"))
		}
		if r.URL.Query().Get("withHtmlContent") != "true" {
			t.Errorf("expected withHtmlContent=true")
		}
		json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{
			Count: 1,
			Results: []types.Document{
				{ID: "v1", Title: "My Video", Category: "video", Content: "<p>Transcript</p>"},
			},
		})
	})
	defer ts.Close()

	handler := makeGetVideoHandler(client)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), GetVideoInput{ID: "v1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestGetVideoHandlerMissingID(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeGetVideoHandler(client)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), GetVideoInput{})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestGetVideoPositionHandler(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.CursorResponse[types.Document]{
			Count: 1,
			Results: []types.Document{
				{ID: "v1", ReadingProgress: 0.75},
			},
		})
	})
	defer ts.Close()

	handler := makeGetVideoPositionHandler(client)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), GetVideoPositionInput{ID: "v1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestGetVideoPositionHandlerMissingID(t *testing.T) {
	client, _, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeGetVideoPositionHandler(client)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), GetVideoPositionInput{})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateVideoPositionHandler(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.Document{ID: "v1", ReadingProgress: 0.5})
	})
	defer ts.Close()

	handler := makeUpdateVideoPositionHandler(client, cm)
	pos := 0.5
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateVideoPositionInput{
		ID:       "v1",
		Position: &pos,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestUpdateVideoPositionHandlerMissingID(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeUpdateVideoPositionHandler(client, cm)
	pos := 0.5
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateVideoPositionInput{
		Position: &pos,
	})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateVideoPositionHandlerMissingPosition(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeUpdateVideoPositionHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateVideoPositionInput{
		ID: "v1",
	})
	if err == nil {
		t.Fatal("expected error for missing position")
	}
}

func TestUpdateVideoPositionHandlerNegativePosition(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeUpdateVideoPositionHandler(client, cm)
	pos := -0.5
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), UpdateVideoPositionInput{
		ID:       "v1",
		Position: &pos,
	})
	if err == nil {
		t.Fatal("expected error for negative position")
	}
}

func TestCreateVideoHighlightHandler(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]types.Highlight{{ID: 1, Text: "spoken text", Location: 120}})
	})
	defer ts.Close()

	handler := makeCreateVideoHighlightHandler(client, cm)
	ts2 := 120.0
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateVideoHighlightInput{
		ID:        "v1",
		Text:      "spoken text",
		Timestamp: &ts2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestCreateVideoHighlightHandlerMissingFields(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeCreateVideoHighlightHandler(client, cm)

	// Missing ID
	ts2 := 120.0
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateVideoHighlightInput{
		Text:      "text",
		Timestamp: &ts2,
	})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}

	// Missing text
	_, _, err = handler(context.Background(), newReqWithAPIKey("test-key"), CreateVideoHighlightInput{
		ID:        "v1",
		Timestamp: &ts2,
	})
	if err == nil {
		t.Fatal("expected error for missing text")
	}

	// Missing timestamp
	_, _, err = handler(context.Background(), newReqWithAPIKey("test-key"), CreateVideoHighlightInput{
		ID:   "v1",
		Text: "text",
	})
	if err == nil {
		t.Fatal("expected error for missing timestamp")
	}
}

func TestCreateVideoHighlightHandlerNegativeTimestamp(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeCreateVideoHighlightHandler(client, cm)
	neg := -1.0
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateVideoHighlightInput{
		ID:        "v1",
		Text:      "text",
		Timestamp: &neg,
	})
	if err == nil {
		t.Fatal("expected error for negative timestamp")
	}
}

func TestCreateVideoHighlightHandlerTextTooLong(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeCreateVideoHighlightHandler(client, cm)
	longText := make([]byte, 8192)
	for i := range longText {
		longText[i] = 'a'
	}
	ts2 := 60.0
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), CreateVideoHighlightInput{
		ID:        "v1",
		Text:      string(longText),
		Timestamp: &ts2,
	})
	if err == nil {
		t.Fatal("expected error for text too long")
	}
}

func TestUpdateVideoPositionCacheInvalidation(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.Document{ID: "v1", ReadingProgress: 0.5})
	})
	defer ts.Close()

	apiKey := "test-key"
	cm.Put(apiKey, "/api/v3/list/", nil, []byte(`cached`))
	if cm.Len() != 1 {
		t.Fatalf("expected 1 cache entry, got %d", cm.Len())
	}

	handler := makeUpdateVideoPositionHandler(client, cm)
	pos := 0.5
	_, _, err := handler(context.Background(), newReqWithAPIKey(apiKey), UpdateVideoPositionInput{
		ID:       "v1",
		Position: &pos,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.Len() != 0 {
		t.Errorf("expected cache to be invalidated, got %d entries", cm.Len())
	}
}
