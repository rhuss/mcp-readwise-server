package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// SaveDocumentInput defines the parameters for the save_document tool.
type SaveDocumentInput struct {
	URL      string   `json:"url" jsonschema:"required,description=URL to save to Reader"`
	Title    string   `json:"title,omitempty" jsonschema:"description=Optional title override"`
	Author   string   `json:"author,omitempty" jsonschema:"description=Optional author"`
	Summary  string   `json:"summary,omitempty" jsonschema:"description=Optional summary"`
	Tags     []string `json:"tags,omitempty" jsonschema:"description=Tags to apply"`
	Location string   `json:"location,omitempty" jsonschema:"description=Location: new later shortlist archive"`
	Category string   `json:"category,omitempty" jsonschema:"description=Category override"`
}

// UpdateDocumentInput defines the parameters for the update_document tool.
type UpdateDocumentInput struct {
	ID       string   `json:"id" jsonschema:"required,description=Document ID to update"`
	Title    string   `json:"title,omitempty" jsonschema:"description=New title"`
	Author   string   `json:"author,omitempty" jsonschema:"description=New author"`
	Summary  string   `json:"summary,omitempty" jsonschema:"description=New summary"`
	Location string   `json:"location,omitempty" jsonschema:"description=New location: new later shortlist archive"`
	Tags     []string `json:"tags,omitempty" jsonschema:"description=New tags"`
	Seen     *bool    `json:"seen,omitempty" jsonschema:"description=Mark as seen"`
}

// CreateHighlightInput defines the parameters for the create_highlight tool.
type CreateHighlightInput struct {
	Text          string `json:"text" jsonschema:"required,description=Highlight text (max 8191 chars)"`
	SourceID      string `json:"source_id,omitempty" jsonschema:"description=Source ID to attach highlight to"`
	SourceTitle   string `json:"source_title,omitempty" jsonschema:"description=Source title (required if source_id not set)"`
	SourceAuthor  string `json:"source_author,omitempty" jsonschema:"description=Source author"`
	SourceURL     string `json:"source_url,omitempty" jsonschema:"description=Source URL"`
	Note          string `json:"note,omitempty" jsonschema:"description=Note attached to the highlight"`
	Location      int    `json:"location,omitempty" jsonschema:"description=Position in source"`
	LocationType  string `json:"location_type,omitempty" jsonschema:"description=Location type: page order time_offset"`
	HighlightedAt string `json:"highlighted_at,omitempty" jsonschema:"description=ISO 8601 datetime when highlighted"`
}

// UpdateHighlightInput defines the parameters for the update_highlight tool.
type UpdateHighlightInput struct {
	ID       string `json:"id" jsonschema:"required,description=Highlight ID to update"`
	Text     string `json:"text,omitempty" jsonschema:"description=New highlight text (max 8191 chars)"`
	Note     string `json:"note,omitempty" jsonschema:"description=New note"`
	Location int    `json:"location,omitempty" jsonschema:"description=New position in source"`
	Color    string `json:"color,omitempty" jsonschema:"description=New highlight color"`
}

// AddSourceTagInput defines the parameters for the add_source_tag tool.
type AddSourceTagInput struct {
	SourceID string `json:"source_id" jsonschema:"required,description=Source ID to tag"`
	Name     string `json:"name" jsonschema:"required,description=Tag name"`
}

// AddHighlightTagInput defines the parameters for the add_highlight_tag tool.
type AddHighlightTagInput struct {
	HighlightID string `json:"highlight_id" jsonschema:"required,description=Highlight ID to tag"`
	Name        string `json:"name" jsonschema:"required,description=Tag name"`
}

// BulkHighlightItem represents a single highlight in a bulk create request.
type BulkHighlightItem struct {
	Text          string `json:"text" jsonschema:"required,description=Highlight text"`
	SourceTitle   string `json:"source_title" jsonschema:"required,description=Source title"`
	SourceAuthor  string `json:"source_author,omitempty" jsonschema:"description=Source author"`
	SourceURL     string `json:"source_url,omitempty" jsonschema:"description=Source URL"`
	Note          string `json:"note,omitempty" jsonschema:"description=Note"`
	Location      int    `json:"location,omitempty" jsonschema:"description=Position in source"`
	HighlightedAt string `json:"highlighted_at,omitempty" jsonschema:"description=ISO 8601 datetime"`
}

// BulkCreateHighlightsInput defines the parameters for the bulk_create_highlights tool.
type BulkCreateHighlightsInput struct {
	Highlights []BulkHighlightItem `json:"highlights" jsonschema:"required,description=Array of highlights to create"`
}

// RegisterWriteTools registers the 7 write profile tools with the MCP server.
func RegisterWriteTools(s *mcp.Server, client *api.Client, cm *cache.Manager) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "save_document",
		Description: "Save a URL to Reader.",
	}, makeSaveDocumentHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_document",
		Description: "Update Reader document metadata (title, author, summary, location, tags).",
	}, makeUpdateDocumentHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_highlight",
		Description: "Create a new highlight. Requires either source_id or source_title.",
	}, makeCreateHighlightHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_highlight",
		Description: "Update an existing highlight's text, note, location, or color.",
	}, makeUpdateHighlightHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_source_tag",
		Description: "Add a tag to a source.",
	}, makeAddSourceTagHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_highlight_tag",
		Description: "Add a tag to a highlight.",
	}, makeAddHighlightTagHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "bulk_create_highlights",
		Description: "Create multiple highlights in a single request. Each highlight requires text and source_title.",
	}, makeBulkCreateHighlightsHandler(client, cm))
}

func makeSaveDocumentHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[SaveDocumentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SaveDocumentInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.URL == "" {
			return nil, nil, fmt.Errorf("url is required")
		}

		result, err := client.SaveDocument(ctx, apiKey, types.SaveDocumentRequest{
			URL:      input.URL,
			Title:    input.Title,
			Author:   input.Author,
			Summary:  input.Summary,
			Tags:     input.Tags,
			Location: input.Location,
			Category: input.Category,
		})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "save_document")

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeUpdateDocumentHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[UpdateDocumentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateDocumentInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		result, err := client.UpdateDocument(ctx, apiKey, input.ID, types.UpdateDocumentRequest{
			Title:    input.Title,
			Author:   input.Author,
			Summary:  input.Summary,
			Location: input.Location,
			Tags:     input.Tags,
			Seen:     input.Seen,
		})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "update_document")

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeCreateHighlightHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[CreateHighlightInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateHighlightInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.Text == "" {
			return nil, nil, fmt.Errorf("text is required")
		}
		if len(input.Text) > 8191 {
			return nil, nil, fmt.Errorf("text exceeds maximum length of 8191 characters")
		}

		hlReq := types.CreateHighlightRequest{
			Text:          input.Text,
			Title:         input.SourceTitle,
			Author:        input.SourceAuthor,
			SourceURL:     input.SourceURL,
			Note:          input.Note,
			Location:      input.Location,
			LocationType:  input.LocationType,
			HighlightedAt: input.HighlightedAt,
		}

		if input.SourceID != "" {
			// Parse source ID if provided
			var bookID int64
			if _, err := fmt.Sscanf(input.SourceID, "%d", &bookID); err == nil {
				hlReq.BookID = bookID
			}
		}

		results, err := client.CreateHighlight(ctx, apiKey, types.CreateHighlightsRequest{
			Highlights: []types.CreateHighlightRequest{hlReq},
		})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "create_highlight")

		data, _ := json.Marshal(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeUpdateHighlightHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[UpdateHighlightInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateHighlightInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}
		if len(input.Text) > 8191 {
			return nil, nil, fmt.Errorf("text exceeds maximum length of 8191 characters")
		}

		result, err := client.UpdateHighlight(ctx, apiKey, input.ID, types.UpdateHighlightRequest{
			Text:     input.Text,
			Note:     input.Note,
			Location: input.Location,
			Color:    input.Color,
		})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "update_highlight")

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeAddSourceTagHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[AddSourceTagInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AddSourceTagInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.SourceID == "" || input.Name == "" {
			return nil, nil, fmt.Errorf("source_id and name are required")
		}

		result, err := client.AddBookTag(ctx, apiKey, input.SourceID, types.CreateTagRequest{Name: input.Name})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "add_source_tag")

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeAddHighlightTagHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[AddHighlightTagInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AddHighlightTagInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.HighlightID == "" || input.Name == "" {
			return nil, nil, fmt.Errorf("highlight_id and name are required")
		}

		result, err := client.AddHighlightTag(ctx, apiKey, input.HighlightID, types.CreateTagRequest{Name: input.Name})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "add_highlight_tag")

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeBulkCreateHighlightsHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[BulkCreateHighlightsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BulkCreateHighlightsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if len(input.Highlights) == 0 {
			return nil, nil, fmt.Errorf("highlights array must not be empty")
		}

		hlReqs := make([]types.CreateHighlightRequest, len(input.Highlights))
		for i, h := range input.Highlights {
			if h.Text == "" {
				return nil, nil, fmt.Errorf("highlights[%d].text is required", i)
			}
			if h.SourceTitle == "" {
				return nil, nil, fmt.Errorf("highlights[%d].source_title is required", i)
			}
			hlReqs[i] = types.CreateHighlightRequest{
				Text:          h.Text,
				Title:         h.SourceTitle,
				Author:        h.SourceAuthor,
				SourceURL:     h.SourceURL,
				Note:          h.Note,
				Location:      h.Location,
				HighlightedAt: h.HighlightedAt,
			}
		}

		results, err := client.CreateHighlight(ctx, apiKey, types.CreateHighlightsRequest{
			Highlights: hlReqs,
		})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "bulk_create_highlights")

		data, _ := json.Marshal(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}
