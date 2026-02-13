package types

import "time"

// Highlight represents a text excerpt from a source.
// Maps to the Readwise v2 /highlights/ API response.
type Highlight struct {
	ID            int64     `json:"id"`
	Text          string    `json:"text"`
	Note          string    `json:"note"`
	Location      int       `json:"location"`
	LocationType  string    `json:"location_type"`
	Color         string    `json:"color"`
	HighlightedAt time.Time `json:"highlighted_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	BookID        int64     `json:"book_id"`
	URL           string    `json:"url"`
	ReadwiseURL   string    `json:"readwise_url"`
	Tags          []Tag     `json:"tags"`
	IsFavorite    bool      `json:"is_favorite"`
	IsDiscard     bool      `json:"is_discard"`
	ExternalID    string    `json:"external_id"`
}

// ReviewHighlight represents a highlight in the daily review response.
type ReviewHighlight struct {
	ID            int64     `json:"id"`
	Text          string    `json:"text"`
	Note          string    `json:"note"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	URL           string    `json:"url"`
	SourceURL     string    `json:"source_url"`
	SourceType    string    `json:"source_type"`
	HighlightedAt time.Time `json:"highlighted_at"`
}

// DailyReview represents the daily review response.
type DailyReview struct {
	ReviewID        int64             `json:"review_id"`
	ReviewURL       string            `json:"review_url"`
	ReviewCompleted bool              `json:"review_completed"`
	Highlights      []ReviewHighlight `json:"highlights"`
}

// CreateHighlightRequest represents a request to create a new highlight.
type CreateHighlightRequest struct {
	Text          string `json:"text"`
	Title         string `json:"title,omitempty"`
	Author        string `json:"author,omitempty"`
	SourceURL     string `json:"source_url,omitempty"`
	SourceType    string `json:"source_type,omitempty"`
	Note          string `json:"note,omitempty"`
	Location      int    `json:"location,omitempty"`
	LocationType  string `json:"location_type,omitempty"`
	HighlightedAt string `json:"highlighted_at,omitempty"`
	BookID        int64  `json:"book_id,omitempty"`
}

// CreateHighlightsRequest wraps a batch of highlight creation requests.
type CreateHighlightsRequest struct {
	Highlights []CreateHighlightRequest `json:"highlights"`
}

// UpdateHighlightRequest represents a request to update an existing highlight.
type UpdateHighlightRequest struct {
	Text     string `json:"text,omitempty"`
	Note     string `json:"note,omitempty"`
	Location int    `json:"location,omitempty"`
	Color    string `json:"color,omitempty"`
}
