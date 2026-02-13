package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
)

// ListSourcesInput defines the parameters for the list_sources tool.
type ListSourcesInput struct {
	PageSize     int    `json:"page_size,omitempty" jsonschema:"description=Number of results per page (1-1000),default=100"`
	Page         int    `json:"page,omitempty" jsonschema:"description=Page number (1-based),default=1"`
	Category     string `json:"category,omitempty" jsonschema:"description=Filter by category: books articles tweets supplementals podcasts"`
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"description=ISO 8601 datetime to filter sources updated after"`
}

// GetSourceInput defines the parameters for the get_source tool.
type GetSourceInput struct {
	ID string `json:"id" jsonschema:"required,description=Source ID"`
}

// ListHighlightsInput defines the parameters for the list_highlights tool.
type ListHighlightsInput struct {
	PageSize     int    `json:"page_size,omitempty" jsonschema:"description=Number of results per page (1-1000),default=100"`
	Page         int    `json:"page,omitempty" jsonschema:"description=Page number (1-based),default=1"`
	SourceID     string `json:"source_id,omitempty" jsonschema:"description=Filter highlights by source ID"`
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"description=ISO 8601 datetime to filter highlights updated after"`
}

// GetHighlightInput defines the parameters for the get_highlight tool.
type GetHighlightInput struct {
	ID string `json:"id" jsonschema:"required,description=Highlight ID"`
}

// ExportHighlightsInput defines the parameters for the export_highlights tool.
type ExportHighlightsInput struct {
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"description=ISO 8601 datetime to filter exports updated after"`
}

// ListSourceTagsInput defines the parameters for the list_source_tags tool.
type ListSourceTagsInput struct {
	SourceID string `json:"source_id" jsonschema:"required,description=Source ID to list tags for"`
}

// ListHighlightTagsInput defines the parameters for the list_highlight_tags tool.
type ListHighlightTagsInput struct {
	HighlightID string `json:"highlight_id" jsonschema:"required,description=Highlight ID to list tags for"`
}

// RegisterReadwiseTools registers the 9 readwise profile tools with the MCP server.
func RegisterReadwiseTools(s *mcp.Server, client *api.Client) {
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
	}, makeListHighlightsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_highlight",
		Description: "Get a single highlight by its ID.",
	}, makeGetHighlightHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "export_highlights",
		Description: "Bulk export all highlights grouped by source. Paginates through all pages automatically. Primary data source for search.",
	}, makeExportHighlightsHandler(client))

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

func makeListHighlightsHandler(client *api.Client) mcp.ToolHandlerFor[ListHighlightsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListHighlightsInput) (*mcp.CallToolResult, any, error) {
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

func makeExportHighlightsHandler(client *api.Client) mcp.ToolHandlerFor[ExportHighlightsInput, any] {
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
