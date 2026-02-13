package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
)

// DeleteHighlightInput defines the parameters for the delete_highlight tool.
type DeleteHighlightInput struct {
	ID string `json:"id" jsonschema:"required,description=Highlight ID to delete"`
}

// DeleteHighlightTagInput defines the parameters for the delete_highlight_tag tool.
type DeleteHighlightTagInput struct {
	HighlightID string `json:"highlight_id" jsonschema:"required,description=Highlight ID"`
	TagID       string `json:"tag_id" jsonschema:"required,description=Tag ID to remove"`
}

// DeleteSourceTagInput defines the parameters for the delete_source_tag tool.
type DeleteSourceTagInput struct {
	SourceID string `json:"source_id" jsonschema:"required,description=Source ID"`
	TagID    string `json:"tag_id" jsonschema:"required,description=Tag ID to remove"`
}

// DeleteDocumentInput defines the parameters for the delete_document tool.
type DeleteDocumentInput struct {
	ID string `json:"id" jsonschema:"required,description=Document ID to delete"`
}

const deletedResponse = `{"deleted":true}`

// RegisterDestructiveTools registers the 4 destructive profile tools with the MCP server.
func RegisterDestructiveTools(s *mcp.Server, client *api.Client, cm *cache.Manager) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_highlight",
		Description: "Delete a highlight permanently.",
	}, makeDeleteHighlightHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_highlight_tag",
		Description: "Remove a tag from a highlight.",
	}, makeDeleteHighlightTagHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_source_tag",
		Description: "Remove a tag from a source.",
	}, makeDeleteSourceTagHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_document",
		Description: "Delete a Reader document permanently.",
	}, makeDeleteDocumentHandler(client, cm))
}

func makeDeleteHighlightHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[DeleteHighlightInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteHighlightInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		if err := client.DeleteHighlight(ctx, apiKey, input.ID); err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "delete_highlight")

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: deletedResponse}},
		}, nil, nil
	}
}

func makeDeleteHighlightTagHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[DeleteHighlightTagInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteHighlightTagInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.HighlightID == "" || input.TagID == "" {
			return nil, nil, fmt.Errorf("highlight_id and tag_id are required")
		}

		if err := client.DeleteHighlightTag(ctx, apiKey, input.HighlightID, input.TagID); err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "delete_highlight_tag")

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: deletedResponse}},
		}, nil, nil
	}
}

func makeDeleteSourceTagHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[DeleteSourceTagInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteSourceTagInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.SourceID == "" || input.TagID == "" {
			return nil, nil, fmt.Errorf("source_id and tag_id are required")
		}

		if err := client.DeleteBookTag(ctx, apiKey, input.SourceID, input.TagID); err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "delete_source_tag")

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: deletedResponse}},
		}, nil, nil
	}
}

func makeDeleteDocumentHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[DeleteDocumentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteDocumentInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		if err := client.DeleteDocument(ctx, apiKey, input.ID); err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "delete_document")

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: deletedResponse}},
		}, nil, nil
	}
}
