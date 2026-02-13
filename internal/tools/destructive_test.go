package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDeleteHighlightHandlerMissingID(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeDeleteHighlightHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightInput{})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestDeleteHighlightHandlerMissingAPIKey(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeDeleteHighlightHandler(client, cm)
	_, _, err := handler(context.Background(), &mcp.CallToolRequest{}, DeleteHighlightInput{ID: "42"})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestDeleteHighlightHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	handler := makeDeleteHighlightHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightInput{ID: "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestDeleteHighlightTagHandlerMissingFields(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeDeleteHighlightTagHandler(client, cm)

	// Missing both
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightTagInput{})
	if err == nil {
		t.Fatal("expected error for missing fields")
	}

	// Missing tag_id
	_, _, err = handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightTagInput{HighlightID: "42"})
	if err == nil {
		t.Fatal("expected error for missing tag_id")
	}

	// Missing highlight_id
	_, _, err = handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightTagInput{TagID: "10"})
	if err == nil {
		t.Fatal("expected error for missing highlight_id")
	}
}

func TestDeleteHighlightTagHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	handler := makeDeleteHighlightTagHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightTagInput{
		HighlightID: "42",
		TagID:       "10",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestDeleteSourceTagHandlerMissingFields(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeDeleteSourceTagHandler(client, cm)

	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteSourceTagInput{})
	if err == nil {
		t.Fatal("expected error for missing fields")
	}
}

func TestDeleteSourceTagHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	handler := makeDeleteSourceTagHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteSourceTagInput{
		SourceID: "5",
		TagID:    "10",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestDeleteDocumentHandlerMissingID(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {})
	defer ts.Close()

	handler := makeDeleteDocumentHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteDocumentInput{})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestDeleteDocumentHandlerSuccess(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	handler := makeDeleteDocumentHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteDocumentInput{ID: "doc-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestDeleteHighlightCacheInvalidation(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	apiKey := "test-key"
	cm.Put(apiKey, "/api/v2/export/", nil, []byte(`cached`))
	if cm.Len() != 1 {
		t.Fatalf("expected 1 cache entry, got %d", cm.Len())
	}

	handler := makeDeleteHighlightHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey(apiKey), DeleteHighlightInput{ID: "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.Len() != 0 {
		t.Errorf("expected cache to be invalidated, got %d entries", cm.Len())
	}
}

func TestDeleteDocumentCacheInvalidation(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	apiKey := "test-key"
	cm.Put(apiKey, "/api/v3/list/", nil, []byte(`cached`))
	if cm.Len() != 1 {
		t.Fatalf("expected 1 cache entry, got %d", cm.Len())
	}

	handler := makeDeleteDocumentHandler(client, cm)
	_, _, err := handler(context.Background(), newReqWithAPIKey(apiKey), DeleteDocumentInput{ID: "doc-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.Len() != 0 {
		t.Errorf("expected cache to be invalidated, got %d entries", cm.Len())
	}
}

func TestDeleteResponseFormat(t *testing.T) {
	client, cm, ts := newWriteTestDeps(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	handler := makeDeleteHighlightHandler(client, cm)
	result, _, err := handler(context.Background(), newReqWithAPIKey("test-key"), DeleteHighlightInput{ID: "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the response format is {"deleted": true}
	text := result.Content[0].(*mcp.TextContent)
	var resp map[string]bool
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !resp["deleted"] {
		t.Error("expected deleted=true in response")
	}
}
