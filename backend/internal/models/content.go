package models

import "time"

// ContentSource represents the source platform of the content.
type ContentSource string

const (
	SourceYouTube ContentSource = "youtube"
	SourceTwitter ContentSource = "twitter"
)

// ParseRequest represents the request body for parsing a URL.
type ParseRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// ContentMetadata represents additional metadata for the parsed content.
type ContentMetadata struct {
	PublishedAt string `json:"publishedAt,omitempty"`
}

// ParseResponse represents the response for a parsed URL.
type ParseResponse struct {
	ID           string          `json:"id"`
	Source       ContentSource   `json:"source"`
	Title        string          `json:"title"`
	Author       string          `json:"author"`
	Summary      string          `json:"summary"`
	ThumbnailURL string          `json:"thumbnailUrl"`
	OriginalURL  string          `json:"originalUrl"`
	Metadata     ContentMetadata `json:"metadata"`
}

// ParseError represents a parsing error response.
type ParseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Common error codes
const (
	ErrorCodeInvalidURL    = "INVALID_URL"
	ErrorCodeParsingFailed = "PARSING_FAILED"
)

// ParsedContent represents the internal parsed content structure.
type ParsedContent struct {
	ID           string
	Source       ContentSource
	Title        string
	Author       string
	Summary      string
	ThumbnailURL string
	OriginalURL  string
	PublishedAt  *time.Time
}

// ToResponse converts ParsedContent to ParseResponse.
func (p *ParsedContent) ToResponse() *ParseResponse {
	resp := &ParseResponse{
		ID:           p.ID,
		Source:       p.Source,
		Title:        p.Title,
		Author:       p.Author,
		Summary:      p.Summary,
		ThumbnailURL: p.ThumbnailURL,
		OriginalURL:  p.OriginalURL,
		Metadata:     ContentMetadata{},
	}

	if p.PublishedAt != nil {
		resp.Metadata.PublishedAt = p.PublishedAt.Format(time.RFC3339)
	}

	return resp
}
