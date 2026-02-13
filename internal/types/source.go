package types

import "time"

// Source represents a book, article, or other item in the user's Readwise library.
// Maps to the Readwise v2 /books/ API response.
type Source struct {
	ID              int64     `json:"user_book_id"`
	Title           string    `json:"title"`
	ReadableTitle   string    `json:"readable_title"`
	Author          string    `json:"author"`
	Category        string    `json:"category"`
	Source          string    `json:"source"`
	CoverImageURL   string    `json:"cover_image_url"`
	SourceURL       string    `json:"source_url"`
	ReadwiseURL     string    `json:"readwise_url"`
	UniqueURL       string    `json:"unique_url"`
	HighlightCount  int       `json:"num_highlights"`
	Tags            []Tag     `json:"book_tags"`
	DocumentNote    string    `json:"document_note"`
	Summary         string    `json:"summary"`
	LastHighlightAt time.Time `json:"last_highlight_at"`
	UpdatedAt       time.Time `json:"updated"`
}

// ExportSource represents a source in the export API response, which includes
// embedded highlights.
type ExportSource struct {
	UserBookID     int64            `json:"user_book_id"`
	Title          string           `json:"title"`
	ReadableTitle  string           `json:"readable_title"`
	Author         string           `json:"author"`
	Category       string           `json:"category"`
	Source         string           `json:"source"`
	CoverImageURL  string           `json:"cover_image_url"`
	SourceURL      string           `json:"source_url"`
	ReadwiseURL    string           `json:"readwise_url"`
	UniqueURL      string           `json:"unique_url"`
	HighlightCount int              `json:"num_highlights"`
	Highlights     []Highlight      `json:"highlights"`
}
