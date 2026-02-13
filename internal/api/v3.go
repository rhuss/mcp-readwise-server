package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// ListDocuments returns documents from the Reader v3 API with cursor-based pagination.
// It paginates through all pages up to the specified limit.
func (c *Client) ListDocuments(ctx context.Context, apiKey string, location, category, updatedAfter string, limit int) (*types.CursorResponse[types.Document], error) {
	var allResults []types.Document
	cursor := ""

	for {
		params := url.Values{}
		if location != "" {
			params.Set("location", location)
		}
		if category != "" {
			params.Set("category", category)
		}
		if updatedAfter != "" {
			params.Set("updatedAfter", updatedAfter)
		}
		if cursor != "" {
			params.Set("pageCursor", cursor)
		}
		// Use limit parameter to control page size
		pageLimit := 100
		if limit > 0 && limit < 100 {
			pageLimit = limit
		}
		params.Set("limit", fmt.Sprintf("%d", pageLimit))

		path := "/list/"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		body, err := c.GetV3(ctx, path, apiKey)
		if err != nil {
			return nil, err
		}

		var page types.CursorResponse[types.Document]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, NewInternalError(fmt.Sprintf("failed to parse documents response: %v", err))
		}

		allResults = append(allResults, page.Results...)

		// Stop if we've reached the limit or no more pages
		if limit > 0 && len(allResults) >= limit {
			allResults = allResults[:limit]
			break
		}
		if page.NextPageCursor == "" {
			break
		}
		cursor = page.NextPageCursor
	}

	return &types.CursorResponse[types.Document]{
		Count:   len(allResults),
		Results: allResults,
	}, nil
}

// GetDocument returns a single Reader document by ID, optionally including content.
func (c *Client) GetDocument(ctx context.Context, apiKey string, id string, withContent bool) (*types.Document, error) {
	params := url.Values{}
	params.Set("id", id)
	if withContent {
		params.Set("withHtmlContent", "true")
	}

	path := "/list/?" + params.Encode()
	body, err := c.GetV3(ctx, path, apiKey)
	if err != nil {
		return nil, err
	}

	var result types.CursorResponse[types.Document]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse document response: %v", err))
	}

	if len(result.Results) == 0 {
		return nil, NewAPIError("not_found", fmt.Sprintf("document %s not found", id))
	}

	return &result.Results[0], nil
}

// ListReaderTags returns all tags from Reader v3 API.
func (c *Client) ListReaderTags(ctx context.Context, apiKey string) ([]types.Tag, error) {
	body, err := c.GetV3(ctx, "/tags/", apiKey)
	if err != nil {
		return nil, err
	}

	var result []types.Tag
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse reader tags response: %v", err))
	}
	return result, nil
}
