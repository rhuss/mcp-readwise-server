package types

// Tag represents a label applied to sources, highlights, or documents.
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CreateTagRequest represents a request to create/add a tag.
type CreateTagRequest struct {
	Name string `json:"name"`
}
