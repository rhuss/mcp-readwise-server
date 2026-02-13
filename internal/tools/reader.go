package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
)

// ListDocumentsInput defines the parameters for the list_documents tool.
type ListDocumentsInput struct {
	Location     string `json:"location,omitempty" jsonschema:"Filter by location: new later shortlist archive feed"`
	Category     string `json:"category,omitempty" jsonschema:"Filter by category: article email rss highlight note pdf epub tweet video"`
	UpdatedAfter string `json:"updated_after,omitempty" jsonschema:"ISO 8601 datetime to filter documents updated after"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum number of results (1-100; default 100)"`
}

// GetDocumentInput defines the parameters for the get_document tool.
type GetDocumentInput struct {
	ID             string `json:"id" jsonschema:"Document ID"`
	IncludeContent bool   `json:"include_content,omitempty" jsonschema:"Include full document content (default false)"`
}

// ListReaderTagsInput is empty since no parameters are needed.
type ListReaderTagsInput struct{}

// RegisterReaderTools registers the 4 reader profile tools with the MCP server.
func RegisterReaderTools(s *mcp.Server, client *api.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_documents",
		Description: "List Reader documents with optional filtering by location (new, later, archive) or category (article, pdf, email, video, etc.).",
	}, makeListDocumentsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_document",
		Description: "Get a single Reader document by ID, optionally including its full content.",
	}, makeGetDocumentHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_reader_tags",
		Description: "List all tags in Reader.",
	}, makeListReaderTagsHandler(client))
}

func makeListDocumentsHandler(client *api.Client) mcp.ToolHandlerFor[ListDocumentsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListDocumentsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}

		if input.Limit < 0 || input.Limit > 100 {
			return nil, nil, fmt.Errorf("limit must be between 1 and 100")
		}

		result, err := client.ListDocuments(ctx, apiKey, input.Location, input.Category, input.UpdatedAfter, input.Limit)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeGetDocumentHandler(client *api.Client) mcp.ToolHandlerFor[GetDocumentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetDocumentInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		result, err := client.GetDocument(ctx, apiKey, input.ID, input.IncludeContent)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeListReaderTagsHandler(client *api.Client) mcp.ToolHandlerFor[ListReaderTagsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ ListReaderTagsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}

		result, err := client.ListReaderTags(ctx, apiKey)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}
