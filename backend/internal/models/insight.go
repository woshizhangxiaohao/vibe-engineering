package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SourceType represents the type of content source.
type SourceType string

const (
	SourceTypeYouTube SourceType = "youtube"
	SourceTypeTwitter SourceType = "twitter"
	SourceTypePodcast SourceType = "podcast"
)

// InsightStatus represents the processing status of an insight.
type InsightStatus string

const (
	InsightStatusPending    InsightStatus = "pending"
	InsightStatusProcessing InsightStatus = "processing"
	InsightStatusCompleted  InsightStatus = "completed"
	InsightStatusFailed     InsightStatus = "failed"
)

// Insight represents a media content analysis record (video, tweet, podcast, etc.).
type Insight struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"index;not null"`

	// Source information
	SourceType SourceType `json:"source_type" gorm:"type:varchar(20);not null"`
	SourceURL  string     `json:"source_url" gorm:"type:varchar(2000);not null"`
	SourceID   string     `json:"source_id" gorm:"type:varchar(100);index"` // video_id, tweet_id, etc.

	// Content metadata
	Title        string     `json:"title" gorm:"type:varchar(500)"`
	Author       string     `json:"author" gorm:"type:varchar(255)"`
	ThumbnailURL string     `json:"thumbnail_url" gorm:"type:varchar(1000)"`
	Duration     int        `json:"duration"`                      // Duration in seconds (for video/audio)
	PublishedAt  *time.Time `json:"published_at" gorm:"index"`     // Original publish time

	// AI generated content
	Summary    string         `json:"summary" gorm:"type:text"`                            // AI generated summary
	KeyPoints  datatypes.JSON `json:"key_points" gorm:"type:jsonb"`                        // Key points as JSON array
	TargetLang string         `json:"target_lang" gorm:"type:varchar(10);default:'zh'"`   // Target language for translation

	// Raw content
	RawContent   string `json:"raw_content" gorm:"type:text"`   // Original transcription/content
	TransContent string `json:"trans_content" gorm:"type:text"` // Translated content

	// Transcripts with timestamps (for video/audio)
	Transcripts datatypes.JSON `json:"transcripts" gorm:"type:jsonb"` // Array of {timestamp, seconds, text}

	// Processing status
	Status       InsightStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	ErrorMessage string        `json:"error_message,omitempty" gorm:"type:text"`

	// Sharing
	ShareToken    *string    `json:"share_token,omitempty" gorm:"type:varchar(64);uniqueIndex"`
	SharePassword string     `json:"-" gorm:"type:varchar(255)"`             // bcrypt hash, never exposed in JSON
	ShareConfig   datatypes.JSON `json:"share_config,omitempty" gorm:"type:jsonb"` // What to include in share
	IsPublic      bool       `json:"is_public" gorm:"default:false"`
	SharedAt      *time.Time `json:"shared_at,omitempty"`

	// Associations
	Highlights   []Highlight   `json:"highlights,omitempty" gorm:"foreignKey:InsightID"`
	ChatMessages []ChatMessage `json:"chat_messages,omitempty" gorm:"foreignKey:InsightID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for Insight model.
func (Insight) TableName() string {
	return "insights"
}

// ShareConfigData represents the configuration for sharing an insight.
type ShareConfigData struct {
	IncludeSummary    bool `json:"include_summary"`
	IncludeKeyPoints  bool `json:"include_key_points"`
	IncludeHighlights bool `json:"include_highlights"`
	IncludeChat       bool `json:"include_chat"`
}

// TranscriptItem represents a single transcript segment with timestamp.
type TranscriptItem struct {
	Timestamp      string `json:"timestamp"`       // e.g., "05:12"
	Seconds        int    `json:"seconds"`         // time in seconds
	Text           string `json:"text"`            // original transcript text
	TranslatedText string `json:"translated_text,omitempty"` // translated text (if available)
}

// Highlight represents a user-created highlight/annotation on content.
type Highlight struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	InsightID uint `json:"insight_id" gorm:"index;not null"`
	UserID    uint `json:"user_id" gorm:"index;not null"`

	Text        string `json:"text" gorm:"type:text;not null"`                    // Highlighted text
	StartOffset int    `json:"start_offset" gorm:"not null"`                      // Start position in content
	EndOffset   int    `json:"end_offset" gorm:"not null"`                        // End position in content
	Color       string `json:"color" gorm:"type:varchar(20);default:'yellow'"`    // Highlight color
	Note        string `json:"note,omitempty" gorm:"type:text"`                   // User's note on the highlight

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for Highlight model.
func (Highlight) TableName() string {
	return "highlights"
}

// ChatMessage represents a message in AI conversation about an insight.
type ChatMessage struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	InsightID uint `json:"insight_id" gorm:"index;not null"`
	UserID    uint `json:"user_id" gorm:"index;not null"`

	Role    string `json:"role" gorm:"type:varchar(20);not null"` // "user" or "assistant"
	Content string `json:"content" gorm:"type:text;not null"`

	// Optional: link to a specific highlight
	HighlightID *uint `json:"highlight_id,omitempty" gorm:"index"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the table name for ChatMessage model.
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// Request/Response DTOs

// CreateInsightRequest represents the request to create a new insight.
type CreateInsightRequest struct {
	SourceURL  string `json:"source_url" binding:"required,url"`
	TargetLang string `json:"target_lang" binding:"omitempty,min=2,max=10"`
}

// CreateInsightResponse represents the response after creating an insight.
type CreateInsightResponse struct {
	ID      uint          `json:"id"`
	Status  InsightStatus `json:"status"`
	Message string        `json:"message"`
}

// InsightListItem represents a single insight in list view.
type InsightListItem struct {
	ID           uint       `json:"id"`
	SourceType   SourceType `json:"source_type"`
	Title        string     `json:"title"`
	Author       string     `json:"author"`
	ThumbnailURL string     `json:"thumbnail_url"`
	Status       InsightStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

// InsightListResponse represents the grouped insight list response.
type InsightListResponse struct {
	Today     []InsightListItem `json:"today"`
	Yesterday []InsightListItem `json:"yesterday"`
	Previous  []InsightListItem `json:"previous"`
	Total     int               `json:"total"`
}

// InsightDetailResponse represents the full insight detail response.
type InsightDetailResponse struct {
	ID           uint             `json:"id"`
	SourceType   SourceType       `json:"source_type"`
	SourceURL    string           `json:"source_url"`
	SourceID     string           `json:"source_id"`
	Title        string           `json:"title"`
	Author       string           `json:"author"`
	ThumbnailURL string           `json:"thumbnail_url"`
	Duration     int              `json:"duration"`
	PublishedAt  *time.Time       `json:"published_at,omitempty"`
	Summary      string           `json:"summary"`
	KeyPoints    []string         `json:"key_points"`
	RawContent   string           `json:"raw_content,omitempty"`
	TransContent string           `json:"trans_content,omitempty"`
	Transcripts  []TranscriptItem `json:"transcripts,omitempty"`
	Status       InsightStatus    `json:"status"`
	Highlights   []Highlight      `json:"highlights,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

// CreateHighlightRequest represents the request to create a highlight.
type CreateHighlightRequest struct {
	Text        string `json:"text" binding:"required"`
	StartOffset int    `json:"start_offset" binding:"required,min=0"`
	EndOffset   int    `json:"end_offset" binding:"required,gtfield=StartOffset"`
	Color       string `json:"color" binding:"omitempty,oneof=yellow green blue purple red"`
	Note        string `json:"note" binding:"omitempty"`
}

// ChatRequest represents a chat message request.
type ChatRequest struct {
	Message     string `json:"message" binding:"required"`
	HighlightID *uint  `json:"highlight_id" binding:"omitempty"`
}

// ChatResponse represents a chat message response (for non-streaming).
type ChatResponse struct {
	ID      uint   `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ShareInsightRequest represents the request to share an insight.
type ShareInsightRequest struct {
	IncludeSummary    bool   `json:"include_summary"`
	IncludeKeyPoints  bool   `json:"include_key_points"`
	IncludeHighlights bool   `json:"include_highlights"`
	IncludeChat       bool   `json:"include_chat"`
	IsPublic          bool   `json:"is_public"`
	Password          string `json:"password,omitempty"`
}

// ShareInsightResponse represents the response after sharing an insight.
type ShareInsightResponse struct {
	ShareToken string     `json:"share_token"`
	ShareURL   string     `json:"share_url"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// SharedInsightResponse represents the public view of a shared insight.
type SharedInsightResponse struct {
	Title        string           `json:"title"`
	Author       string           `json:"author"`
	ThumbnailURL string           `json:"thumbnail_url"`
	SharedBy     string           `json:"shared_by"`
	SharedAt     *time.Time       `json:"shared_at"`
	SourceType   SourceType       `json:"source_type"`
	SourceURL    string           `json:"source_url"`
	Content      SharedContent    `json:"content"`
}

// SharedContent represents the content included in a shared insight.
type SharedContent struct {
	Summary    string      `json:"summary,omitempty"`
	KeyPoints  []string    `json:"key_points,omitempty"`
	Highlights []Highlight `json:"highlights,omitempty"`
	Chat       []ChatMessage `json:"chat,omitempty"`
}
