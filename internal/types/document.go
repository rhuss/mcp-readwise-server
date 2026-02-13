package types

import "time"

// Document represents a Reader item (article, email, PDF, video, etc.).
// Maps to the Reader v3 /list/ API response.
type Document struct {
	ID              string          `json:"id"`
	URL             string          `json:"url"`
	SourceURL       string          `json:"source_url"`
	Title           string          `json:"title"`
	Author          string          `json:"author"`
	Source          string          `json:"source"`
	Category        string          `json:"category"`
	Location        string          `json:"location"`
	Tags            map[string]Tag  `json:"tags"`
	SiteName        string          `json:"site_name"`
	WordCount       int             `json:"word_count"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	PublishedDate   string          `json:"published_date"`
	Summary         string          `json:"summary"`
	ImageURL        string          `json:"image_url"`
	Notes           string          `json:"notes"`
	ParentID        string          `json:"parent_id"`
	ReadingProgress float64         `json:"reading_progress"`
	FirstOpenedAt   time.Time       `json:"first_opened_at"`
	LastOpenedAt    time.Time       `json:"last_opened_at"`
	LastMovedAt     time.Time       `json:"last_moved_at"`
	SavedAt         time.Time       `json:"saved_at"`
	Content         string          `json:"html,omitempty"`
}

// SaveDocumentRequest represents a request to save a document to Reader.
type SaveDocumentRequest struct {
	URL            string   `json:"url"`
	HTML           string   `json:"html,omitempty"`
	ShouldCleanHTML bool    `json:"should_clean_html,omitempty"`
	Title          string   `json:"title,omitempty"`
	Author         string   `json:"author,omitempty"`
	Summary        string   `json:"summary,omitempty"`
	PublishedDate  string   `json:"published_date,omitempty"`
	ImageURL       string   `json:"image_url,omitempty"`
	Location       string   `json:"location,omitempty"`
	Category       string   `json:"category,omitempty"`
	SavedUsing     string   `json:"saved_using,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

// SaveDocumentResponse represents the response from saving a document.
type SaveDocumentResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// UpdateDocumentRequest represents a request to update document metadata.
type UpdateDocumentRequest struct {
	Title    string   `json:"title,omitempty"`
	Author   string   `json:"author,omitempty"`
	Summary  string   `json:"summary,omitempty"`
	Location string   `json:"location,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Seen     *bool    `json:"seen,omitempty"`
	ReadingProgress *float64 `json:"reading_progress,omitempty"`
}
