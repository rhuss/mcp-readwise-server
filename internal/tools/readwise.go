package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// ListSourcesInput defines the parameters for the list_sources tool.
type ListSourcesInput struct {
	PageSize     int    `json:"page_size,omitempty" jsonschema:"Number of results per page (1-1000; default 100)"`
	Page         int    `json:"page,omitempty" jsonschema:"Page number (1-based; default 1)"`
	Category     string `json:"category,omitempty" jsonschema:"Filter by category: books articles tweets supplementals podcasts"`
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"ISO 8601 datetime to filter sources updated after"`
}

// GetSourceInput defines the parameters for the get_source tool.
type GetSourceInput struct {
	ID string `json:"id" jsonschema:"Source ID"`
}

// ListHighlightsInput defines the parameters for the list_highlights tool.
type ListHighlightsInput struct {
	PageSize     int    `json:"page_size,omitempty" jsonschema:"Number of results per page (1-1000; default 100)"`
	Page         int    `json:"page,omitempty" jsonschema:"Page number (1-based; default 1)"`
	SourceID     string `json:"source_id,omitempty" jsonschema:"Filter highlights by source ID"`
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"ISO 8601 datetime to filter highlights updated after"`
}

// GetHighlightInput defines the parameters for the get_highlight tool.
type GetHighlightInput struct {
	ID string `json:"id" jsonschema:"Highlight ID"`
}

// ExportHighlightsInput defines the parameters for the export_highlights tool.
type ExportHighlightsInput struct {
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"ISO 8601 datetime to filter exports updated after"`
}

// ListSourceTagsInput defines the parameters for the list_source_tags tool.
type ListSourceTagsInput struct {
	SourceID string `json:"source_id" jsonschema:"Source ID to list tags for"`
}

// ListHighlightTagsInput defines the parameters for the list_highlight_tags tool.
type ListHighlightTagsInput struct {
	HighlightID string `json:"highlight_id" jsonschema:"Highlight ID to list tags for"`
}

// RegisterReadwiseTools registers the 9 readwise profile tools with the MCP server.
func RegisterReadwiseTools(s *mcp.Server, client *api.Client, cm *cache.Manager) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_sources",
		Description: "List highlight sources (books, articles, etc.) with pagination and optional filtering by category or update time.",
	}, makeListSourcesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_source",
		Description: "Get details of a single source by its ID.",
	}, makeGetSourceHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_highlights",
		Description: "List highlights with pagination and optional filtering by source ID or update time.",
	}, makeListHighlightsHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_highlight",
		Description: "Get a single highlight by its ID.",
	}, makeGetHighlightHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "export_highlights",
		Description: "Bulk export all highlights grouped by source. Paginates through all pages automatically. Primary data source for search.",
	}, makeExportHighlightsHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_daily_review",
		Description: "Get today's daily review highlights from Readwise.",
	}, makeGetDailyReviewHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_source_tags",
		Description: "List all tags applied to a specific source.",
	}, makeListSourceTagsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_highlight_tags",
		Description: "List all tags applied to a specific highlight.",
	}, makeListHighlightTagsHandler(client))
}

func makeListSourcesHandler(client *api.Client) mcp.ToolHandlerFor[ListSourcesInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListSourcesInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}

		if input.PageSize < 0 || input.PageSize > 1000 {
			return nil, nil, fmt.Errorf("page_size must be between 1 and 1000")
		}
		if input.Page < 0 {
			return nil, nil, fmt.Errorf("page must be positive")
		}

		result, err := client.ListBooks(ctx, apiKey, input.Page, input.PageSize, input.Category, input.UpdatedAfter)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeGetSourceHandler(client *api.Client) mcp.ToolHandlerFor[GetSourceInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetSourceInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		result, err := client.GetBook(ctx, apiKey, input.ID)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func isNumericID(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func makeListHighlightsHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[ListHighlightsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListHighlightsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}

		if input.SourceID != "" && !isNumericID(input.SourceID) {
			return listHighlightsForULID(ctx, client, cm, apiKey, input.SourceID)
		}

		if input.PageSize < 0 || input.PageSize > 1000 {
			return nil, nil, fmt.Errorf("page_size must be between 1 and 1000")
		}
		if input.Page < 0 {
			return nil, nil, fmt.Errorf("page must be positive")
		}

		result, err := client.ListHighlights(ctx, apiKey, input.Page, input.PageSize, input.SourceID, input.UpdatedAfter)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func getOrFetchExportSources(ctx context.Context, client *api.Client, cm *cache.Manager, apiKey string) ([]types.ExportSource, error) {
	if cached := cm.Get(apiKey, "/api/v2/export/", nil); cached != nil {
		var resp types.CursorResponse[types.ExportSource]
		if err := json.Unmarshal(cached, &resp); err == nil {
			return resp.Results, nil
		}
	}
	exportData, err := client.ExportHighlights(ctx, apiKey, "")
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(exportData); err == nil {
		cm.Put(apiKey, "/api/v2/export/", nil, data)
	}
	return exportData.Results, nil
}

func listHighlightsForULID(ctx context.Context, client *api.Client, cm *cache.Manager, apiKey, sourceID string) (*mcp.CallToolResult, any, error) {
	doc, err := client.GetDocument(ctx, apiKey, sourceID, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get document %s: %w", sourceID, err)
	}

	sources, err := getOrFetchExportSources(ctx, client, cm, apiKey)
	if err != nil {
		return nil, nil, err
	}

	var highlights []types.Highlight
	for _, source := range sources {
		if matchesDocument(source, doc) {
			highlights = append(highlights, source.Highlights...)
		}
	}

	data, _ := json.Marshal(highlights)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func matchesDocument(source types.ExportSource, doc *types.Document) bool {
	if doc.SourceURL != "" && source.SourceURL != "" &&
		strings.EqualFold(source.SourceURL, doc.SourceURL) {
		return true
	}
	if doc.Title != "" && source.Title != "" &&
		strings.EqualFold(source.Title, doc.Title) &&
		(doc.Author == "" || source.Author == "" || strings.EqualFold(source.Author, doc.Author)) {
		return true
	}
	return false
}

func makeGetHighlightHandler(client *api.Client) mcp.ToolHandlerFor[GetHighlightInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetHighlightInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		result, err := client.GetHighlight(ctx, apiKey, input.ID)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeExportHighlightsHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[ExportHighlightsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ExportHighlightsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}

		result, err := client.ExportHighlights(ctx, apiKey, input.UpdatedAfter)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		if input.UpdatedAfter == "" {
			cm.Put(apiKey, "/api/v2/export/", nil, data)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeGetDailyReviewHandler(client *api.Client) mcp.ToolHandlerFor[struct{}, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}

		result, err := client.GetDailyReview(ctx, apiKey)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeListSourceTagsHandler(client *api.Client) mcp.ToolHandlerFor[ListSourceTagsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListSourceTagsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.SourceID == "" {
			return nil, nil, fmt.Errorf("source_id is required")
		}

		result, err := client.ListBookTags(ctx, apiKey, input.SourceID)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeListHighlightTagsHandler(client *api.Client) mcp.ToolHandlerFor[ListHighlightTagsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListHighlightTagsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.HighlightID == "" {
			return nil, nil, fmt.Errorf("highlight_id is required")
		}

		result, err := client.ListHighlightTags(ctx, apiKey, input.HighlightID)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}
