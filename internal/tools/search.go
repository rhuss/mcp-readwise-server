package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/auth"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// SearchHighlightsInput defines the parameters for the search_highlights tool.
type SearchHighlightsInput struct {
	Query    string `json:"query" jsonschema:"required,description=Search query to match against highlight text and notes and source titles"`
	SourceID string `json:"source_id,omitempty" jsonschema:"description=Filter results to a specific source ID"`
	Limit    int    `json:"limit,omitempty" jsonschema:"description=Maximum number of results (1-200),default=50"`
}

// SearchHighlightResult represents a single search result.
type SearchHighlightResult struct {
	Highlight      types.Highlight `json:"highlight"`
	SourceTitle    string          `json:"source_title"`
	RelevanceScore float64         `json:"relevance_score"`
}

// SearchDocumentsInput defines the parameters for the search_documents tool.
type SearchDocumentsInput struct {
	Query    string `json:"query" jsonschema:"required,description=Search query to match against document title and author and summary and notes"`
	Location string `json:"location,omitempty" jsonschema:"description=Filter by location: new later shortlist archive feed"`
	Category string `json:"category,omitempty" jsonschema:"description=Filter by category: article email rss highlight note pdf epub tweet video"`
	Limit    int    `json:"limit,omitempty" jsonschema:"description=Maximum number of results (1-200),default=50"`
}

// SearchDocumentResult represents a single document search result.
type SearchDocumentResult struct {
	Document       types.Document `json:"document"`
	RelevanceScore float64        `json:"relevance_score"`
}

// RegisterSearchHighlightsTool registers the search_highlights tool.
func RegisterSearchHighlightsTool(s *mcp.Server, client *api.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_highlights",
		Description: "Search highlights by query. Searches across highlight text, notes, and source titles. Returns results ranked by relevance.",
	}, makeSearchHighlightsHandler(client))
}

// RegisterSearchDocumentsTool registers the search_documents tool.
func RegisterSearchDocumentsTool(s *mcp.Server, client *api.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_documents",
		Description: "Search Reader documents by query. Searches across title, author, summary, and notes. Supports location and category filtering.",
	}, makeSearchDocumentsHandler(client))
}

func makeSearchHighlightsHandler(client *api.Client) mcp.ToolHandlerFor[SearchHighlightsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchHighlightsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.Query == "" {
			return nil, nil, fmt.Errorf("query is required")
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}

		// Fetch export data (in future phases, this will use the cache)
		exportData, err := client.ExportHighlights(ctx, apiKey, "")
		if err != nil {
			return nil, nil, err
		}

		results := searchHighlights(exportData.Results, input.Query, input.SourceID, limit)

		data, _ := json.Marshal(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

func makeSearchDocumentsHandler(client *api.Client) mcp.ToolHandlerFor[SearchDocumentsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchDocumentsInput) (*mcp.CallToolResult, any, error) {
		apiKey := auth.APIKeyFromRequest(req)
		if apiKey == "" {
			return nil, nil, fmt.Errorf("missing API key: provide your Readwise API key via the Authorization header")
		}
		if input.Query == "" {
			return nil, nil, fmt.Errorf("query is required")
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}

		// Fetch document list (in future phases, this will use the cache)
		docData, err := client.ListDocuments(ctx, apiKey, "", "", "", 0)
		if err != nil {
			return nil, nil, err
		}

		results := searchDocuments(docData.Results, input.Query, input.Location, input.Category, limit)

		data, _ := json.Marshal(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

// searchHighlights performs case-insensitive search across export data.
func searchHighlights(sources []types.ExportSource, query, sourceID string, limit int) []SearchHighlightResult {
	queryLower := strings.ToLower(query)
	var results []SearchHighlightResult

	for _, source := range sources {
		if sourceID != "" && fmt.Sprintf("%d", source.UserBookID) != sourceID {
			continue
		}

		titleLower := strings.ToLower(source.Title)

		for _, h := range source.Highlights {
			score := scoreHighlight(h, titleLower, queryLower)
			if score > 0 {
				results = append(results, SearchHighlightResult{
					Highlight:      h,
					SourceTitle:    source.Title,
					RelevanceScore: score,
				})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// scoreHighlight calculates a relevance score for a highlight against a query.
func scoreHighlight(h types.Highlight, titleLower, queryLower string) float64 {
	var score float64

	textLower := strings.ToLower(h.Text)
	noteLower := strings.ToLower(h.Note)

	// Exact matches score higher
	if strings.EqualFold(h.Text, queryLower) {
		score += 10.0
	} else if strings.Contains(textLower, queryLower) {
		score += 5.0
	}

	if strings.EqualFold(h.Note, queryLower) {
		score += 8.0
	} else if noteLower != "" && strings.Contains(noteLower, queryLower) {
		score += 4.0
	}

	if strings.Contains(titleLower, queryLower) {
		score += 3.0
	}

	return score
}

// searchDocuments performs case-insensitive search across document data.
func searchDocuments(docs []types.Document, query, location, category string, limit int) []SearchDocumentResult {
	queryLower := strings.ToLower(query)
	var results []SearchDocumentResult

	for _, doc := range docs {
		if location != "" && doc.Location != location {
			continue
		}
		if category != "" && doc.Category != category {
			continue
		}

		score := scoreDocument(doc, queryLower)
		if score > 0 {
			results = append(results, SearchDocumentResult{
				Document:       doc,
				RelevanceScore: score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// scoreDocument calculates a relevance score for a document against a query.
func scoreDocument(doc types.Document, queryLower string) float64 {
	var score float64

	titleLower := strings.ToLower(doc.Title)
	authorLower := strings.ToLower(doc.Author)
	summaryLower := strings.ToLower(doc.Summary)
	notesLower := strings.ToLower(doc.Notes)

	if strings.EqualFold(doc.Title, queryLower) {
		score += 10.0
	} else if strings.Contains(titleLower, queryLower) {
		score += 5.0
	}

	if strings.EqualFold(doc.Author, queryLower) {
		score += 8.0
	} else if authorLower != "" && strings.Contains(authorLower, queryLower) {
		score += 4.0
	}

	if summaryLower != "" && strings.Contains(summaryLower, queryLower) {
		score += 3.0
	}

	if notesLower != "" && strings.Contains(notesLower, queryLower) {
		score += 3.0
	}

	return score
}
