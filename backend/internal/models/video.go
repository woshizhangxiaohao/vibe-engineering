package models

import (
	"time"

	"gorm.io/gorm"
)

// VideoAnalysis represents a complete video analysis record.
type VideoAnalysis struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	UserID           uint           `json:"user_id" gorm:"index;not null"`
	VideoID          string         `json:"video_id" gorm:"type:varchar(255);index;not null"`
	Title            string         `json:"title" gorm:"type:varchar(500)"`
	Author           string         `json:"author" gorm:"type:varchar(255)"`
	ThumbnailURL     string         `json:"thumbnail_url" gorm:"type:varchar(1000)"`
	Duration         int            `json:"duration"` // in seconds
	TargetLanguage   string         `json:"target_language" gorm:"type:varchar(10)"`
	Summary          string         `json:"summary" gorm:"type:text"`
	Status           string         `json:"status" gorm:"type:varchar(50);default:'pending'"` // pending, processing, completed, failed
	JobID            string         `json:"job_id" gorm:"type:varchar(100);uniqueIndex"`
	ErrorMessage     string         `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for VideoAnalysis model.
func (VideoAnalysis) TableName() string {
	return "video_analyses"
}

// Chapter represents a video chapter/section.
type Chapter struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AnalysisID  uint      `json:"analysis_id" gorm:"index;not null"`
	Title       string    `json:"title" gorm:"type:varchar(500)"`
	Timestamp   string    `json:"timestamp" gorm:"type:varchar(20)"` // e.g., "05:12"
	Seconds     int       `json:"seconds"`                           // time in seconds
	OrderIndex  int       `json:"order_index"`                       // for maintaining order
	CreatedAt   time.Time `json:"created_at"`
}

// TableName returns the table name for Chapter model.
func (Chapter) TableName() string {
	return "chapters"
}

// Transcription represents a transcribed text segment with timestamp.
type Transcription struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AnalysisID  uint      `json:"analysis_id" gorm:"index;not null"`
	Text        string    `json:"text" gorm:"type:text"`
	Timestamp   string    `json:"timestamp" gorm:"type:varchar(20)"` // e.g., "05:12"
	Seconds     int       `json:"seconds"`                           // time in seconds
	OrderIndex  int       `json:"order_index"`                       // for maintaining order
	CreatedAt   time.Time `json:"created_at"`
}

// TableName returns the table name for Transcription model.
func (Transcription) TableName() string {
	return "transcriptions"
}

// KeyPoint represents a core viewpoint extracted from the video.
type KeyPoint struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AnalysisID  uint      `json:"analysis_id" gorm:"index;not null"`
	Content     string    `json:"content" gorm:"type:text"`
	OrderIndex  int       `json:"order_index"` // for maintaining order
	CreatedAt   time.Time `json:"created_at"`
}

// TableName returns the table name for KeyPoint model.
func (KeyPoint) TableName() string {
	return "key_points"
}

// Request/Response DTOs

// MetadataRequest represents the request to fetch video metadata.
type MetadataRequest struct {
	URL string `json:"url" binding:"required"`
}

// MetadataResponse represents the video metadata response.
type MetadataResponse struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Duration     int    `json:"duration"` // in seconds
}

// MetadataWithAIResponse represents the video metadata response with AI analysis.
type MetadataWithAIResponse struct {
	Title       string       `json:"title"`
	Author      string       `json:"author"`
	Description string       `json:"description"`
	AIAnalysis  string       `json:"aiAnalysis"`
	Metadata    MetadataInfo `json:"metadata"`
}

// MetadataInfo contains additional metadata details.
type MetadataInfo struct {
	Duration  int    `json:"duration"`
	Thumbnail string `json:"thumbnail"`
}

// AnalyzeRequest represents the request to analyze a video.
type AnalyzeRequest struct {
	VideoID        string `json:"videoId" binding:"required"`
	TargetLanguage string `json:"targetLanguage" binding:"required,min=2,max=10"`
}

// AnalyzeResponse represents the response when submitting an analysis job.
type AnalyzeResponse struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

// ChapterResponse represents a chapter in the API response.
type ChapterResponse struct {
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	Seconds   int    `json:"seconds"`
}

// TranscriptionResponse represents a transcription segment in the API response.
type TranscriptionResponse struct {
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
	Seconds   int    `json:"seconds"`
}

// AnalysisResultResponse represents the complete analysis result.
type AnalysisResultResponse struct {
	AnalysisID    uint                     `json:"analysisId,omitempty"`
	Status        string                   `json:"status"`
	Summary       string                   `json:"summary,omitempty"`
	KeyPoints     []string                 `json:"keyPoints,omitempty"`
	Chapters      []ChapterResponse        `json:"chapters,omitempty"`
	Transcription []TranscriptionResponse  `json:"transcription,omitempty"`
}

// HistoryItem represents a single history record.
type HistoryItem struct {
	VideoID      string    `json:"videoId"`
	Title        string    `json:"title"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	CreatedAt    time.Time `json:"createdAt"`
}

// HistoryResponse represents the list of history items.
type HistoryResponse struct {
	Items []HistoryItem `json:"items"`
}

// ExportRequest represents the request to export analysis results.
type ExportRequest struct {
	VideoID string `json:"videoId" binding:"required"`
	Format  string `json:"format" binding:"required,oneof=pdf markdown"`
}

// ExportResponse represents the export result.
type ExportResponse struct {
	DownloadURL string `json:"downloadUrl"`
	FileName    string `json:"fileName"`
}
