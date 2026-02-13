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

// ListVideosInput defines the parameters for the list_videos tool.
type ListVideosInput struct {
	Location string `json:"location,omitempty" jsonschema:"Filter by location: new later shortlist archive feed"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Max results (1-100 default 50)"`
}

// GetVideoInput defines the parameters for the get_video tool.
type GetVideoInput struct {
	ID string `json:"id" jsonschema:"Video document ID"`
}

// GetVideoPositionInput defines the parameters for the get_video_position tool.
type GetVideoPositionInput struct {
	ID string `json:"id" jsonschema:"Video document ID"`
}

// UpdateVideoPositionInput defines the parameters for the update_video_position tool.
type UpdateVideoPositionInput struct {
	ID       string   `json:"id" jsonschema:"Video document ID"`
	Position *float64 `json:"position" jsonschema:"Reading progress (0.0 to 1.0)"`
}

// CreateVideoHighlightInput defines the parameters for the create_video_highlight tool.
type CreateVideoHighlightInput struct {
	ID           string   `json:"id" jsonschema:"Video document ID"`
	Text         string   `json:"text" jsonschema:"Highlight text (max 8191 chars)"`
	Timestamp    *float64 `json:"timestamp" jsonschema:"Timestamp in seconds"`
	EndTimestamp *float64 `json:"end_timestamp,omitempty" jsonschema:"End timestamp in seconds"`
	Note         string   `json:"note,omitempty" jsonschema:"Note attached to the highlight"`
}

// RegisterVideoTools registers the 5 video profile tools with the MCP server.
func RegisterVideoTools(s *mcp.Server, client *api.Client, cm *cache.Manager) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_videos",
		Description: "List video documents from Reader, filtered to video category.",
	}, makeListVideosHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_video",
		Description: "Get a video document with transcript content.",
	}, makeGetVideoHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_video_position",
		Description: "Get the current playback position of a video.",
	}, makeGetVideoPositionHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_video_position",
		Description: "Update the playback position of a video.",
	}, makeUpdateVideoPositionHandler(client, cm))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_video_highlight",
		Description: "Create a timestamped highlight on a video document.",
	}, makeCreateVideoHighlightHandler(client, cm))
}

func makeListVideosHandler(client *api.Client) mcp.ToolHandlerFor[ListVideosInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListVideosInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 100 {
			limit = 100
		}

		result, err := client.ListDocuments(ctx, apiKey, input.Location, "video", "", limit)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeGetVideoHandler(client *api.Client) mcp.ToolHandlerFor[GetVideoInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetVideoInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		result, err := client.GetDocument(ctx, apiKey, input.ID, true)
		if err != nil {
			return nil, nil, err
		}

		data, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeGetVideoPositionHandler(client *api.Client) mcp.ToolHandlerFor[GetVideoPositionInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetVideoPositionInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}

		doc, err := client.GetDocument(ctx, apiKey, input.ID, false)
		if err != nil {
			return nil, nil, err
		}

		positionInfo := struct {
			ReadingProgress float64 `json:"reading_progress"`
			LastOpenedAt    string  `json:"last_opened_at"`
		}{
			ReadingProgress: doc.ReadingProgress,
			LastOpenedAt:    doc.LastOpenedAt.Format("2006-01-02T15:04:05Z"),
		}

		data, _ := json.Marshal(positionInfo)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeUpdateVideoPositionHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[UpdateVideoPositionInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateVideoPositionInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}
		if input.Position == nil {
			return nil, nil, fmt.Errorf("position is required")
		}
		if *input.Position < 0 {
			return nil, nil, fmt.Errorf("position must be >= 0.0")
		}

		result, err := client.UpdateDocument(ctx, apiKey, input.ID, types.UpdateDocumentRequest{
			ReadingProgress: input.Position,
		})
		if err != nil {
			return nil, nil, err
		}

		cm.Invalidate(apiKey, "update_document")

		positionResult := struct {
			ReadingProgress float64 `json:"reading_progress"`
		}{
			ReadingProgress: result.ReadingProgress,
		}

		data, _ := json.Marshal(positionResult)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeCreateVideoHighlightHandler(client *api.Client, cm *cache.Manager) mcp.ToolHandlerFor[CreateVideoHighlightInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateVideoHighlightInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key")
		}
		if input.ID == "" {
			return nil, nil, fmt.Errorf("id is required")
		}
		if input.Text == "" {
			return nil, nil, fmt.Errorf("text is required")
		}
		if len(input.Text) > 8191 {
			return nil, nil, fmt.Errorf("text exceeds maximum length of 8191 characters")
		}
		if input.Timestamp == nil {
			return nil, nil, fmt.Errorf("timestamp is required")
		}
		if *input.Timestamp < 0 {
			return nil, nil, fmt.Errorf("timestamp must be >= 0.0")
		}

		hlReq := types.CreateHighlightRequest{
			Text:         input.Text,
			Note:         input.Note,
			Location:     int(*input.Timestamp),
			LocationType: "time_offset",
		}

		// Parse document ID as book ID
		var bookID int64
		if _, err := fmt.Sscanf(input.ID, "%d", &bookID); err == nil {
			hlReq.BookID = bookID
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
