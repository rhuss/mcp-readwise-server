package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// ListBooks returns a paginated list of sources (books) from the Readwise v2 API.
func (c *Client) ListBooks(ctx context.Context, apiKey string, page, pageSize int, category, updatedAfter string) (*types.PageResponse[types.Source], error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	if category != "" {
		params.Set("category", category)
	}
	if updatedAfter != "" {
		params.Set("updated__gt", updatedAfter)
	}

	path := "/books/"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := c.GetV2(ctx, path, apiKey)
	if err != nil {
		return nil, err
	}

	var result types.PageResponse[types.Source]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse books response: %v", err))
	}
	return &result, nil
}

// GetBook returns a single source (book) by ID.
func (c *Client) GetBook(ctx context.Context, apiKey string, id string) (*types.Source, error) {
	body, err := c.GetV2(ctx, fmt.Sprintf("/books/%s/", id), apiKey)
	if err != nil {
		return nil, err
	}

	var result types.Source
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse book response: %v", err))
	}
	return &result, nil
}

// ListHighlights returns a paginated list of highlights from the Readwise v2 API.
func (c *Client) ListHighlights(ctx context.Context, apiKey string, page, pageSize int, bookID, updatedAfter string) (*types.PageResponse[types.Highlight], error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	if bookID != "" {
		params.Set("book_id", bookID)
	}
	if updatedAfter != "" {
		params.Set("updated__gt", updatedAfter)
	}

	path := "/highlights/"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := c.GetV2(ctx, path, apiKey)
	if err != nil {
		return nil, err
	}

	var result types.PageResponse[types.Highlight]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse highlights response: %v", err))
	}
	return &result, nil
}

// GetHighlight returns a single highlight by ID.
func (c *Client) GetHighlight(ctx context.Context, apiKey string, id string) (*types.Highlight, error) {
	body, err := c.GetV2(ctx, fmt.Sprintf("/highlights/%s/", id), apiKey)
	if err != nil {
		return nil, err
	}

	var result types.Highlight
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse highlight response: %v", err))
	}
	return &result, nil
}

// ExportHighlights exports all highlights with cursor-based pagination.
// It loops through all pages and returns the complete result.
func (c *Client) ExportHighlights(ctx context.Context, apiKey string, updatedAfter string) (*types.CursorResponse[types.ExportSource], error) {
	var allResults []types.ExportSource
	totalCount := 0
	cursor := ""

	for {
		params := url.Values{}
		if updatedAfter != "" {
			params.Set("updatedAfter", updatedAfter)
		}
		if cursor != "" {
			params.Set("pageCursor", cursor)
		}

		path := "/export/"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		body, err := c.GetV2(ctx, path, apiKey)
		if err != nil {
			return nil, err
		}

		var page types.CursorResponse[types.ExportSource]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, NewInternalError(fmt.Sprintf("failed to parse export response: %v", err))
		}

		allResults = append(allResults, page.Results...)
		totalCount += page.Count

		if page.NextPageCursor == "" {
			break
		}
		cursor = page.NextPageCursor
	}

	return &types.CursorResponse[types.ExportSource]{
		Count:   totalCount,
		Results: allResults,
	}, nil
}

// GetDailyReview returns today's daily review highlights.
func (c *Client) GetDailyReview(ctx context.Context, apiKey string) (*types.DailyReview, error) {
	body, err := c.GetV2(ctx, "/review/", apiKey)
	if err != nil {
		return nil, err
	}

	var result types.DailyReview
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse daily review response: %v", err))
	}
	return &result, nil
}

// ListBookTags returns tags for a specific book (source).
func (c *Client) ListBookTags(ctx context.Context, apiKey string, bookID string) ([]types.Tag, error) {
	body, err := c.GetV2(ctx, fmt.Sprintf("/books/%s/tags", bookID), apiKey)
	if err != nil {
		return nil, err
	}

	var result []types.Tag
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse book tags response: %v", err))
	}
	return result, nil
}

// ListHighlightTags returns tags for a specific highlight.
func (c *Client) ListHighlightTags(ctx context.Context, apiKey string, highlightID string) ([]types.Tag, error) {
	body, err := c.GetV2(ctx, fmt.Sprintf("/highlights/%s/tags", highlightID), apiKey)
	if err != nil {
		return nil, err
	}

	var result []types.Tag
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to parse highlight tags response: %v", err))
	}
	return result, nil
}
