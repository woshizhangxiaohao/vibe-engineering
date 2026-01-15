package models

import (
	"time"

	"gorm.io/gorm"
)

// ChatMessage represents a single chat message in a conversation.
type ChatMessage struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	AnalysisID  uint           `json:"analysis_id" gorm:"index;not null"`
	Role        string         `json:"role" gorm:"type:varchar(20);not null"` // "user" or "assistant"
	Content     string         `json:"content" gorm:"type:text;not null"`
	HighlightID *uint          `json:"highlight_id,omitempty" gorm:"index"` // optional: linked to a highlight
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for ChatMessage model.
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// Entity represents a detected entity in the content.
type Entity struct {
	Type   string `json:"type"`   // "stock", "crypto", etc.
	Name   string `json:"name"`   // "NVIDIA", "Bitcoin"
	Ticker string `json:"ticker"` // "NVDA", "BTC"
}

// Suggestion represents an AI-generated suggestion.
type Suggestion struct {
	Type   string `json:"type"`   // "position", "prediction"
	Entity string `json:"entity"` // ticker
	Prompt string `json:"prompt"` // suggested prompt
}

// Request/Response DTOs

// ChatRequest represents a chat message request.
type ChatRequest struct {
	Message     string `json:"message" binding:"required"`
	HighlightID *uint  `json:"highlight_id,omitempty"`
}

// ChatMessageResponse represents a single message in the response.
type ChatMessageResponse struct {
	ID        uint      `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatHistoryResponse represents the chat history response.
type ChatHistoryResponse struct {
	Messages []ChatMessageResponse `json:"messages"`
}

// ChatStreamEvent represents a streaming chat event.
type ChatStreamEvent struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Done      bool   `json:"done"`
	MessageID *uint  `json:"message_id,omitempty"`
}

// AnalyzeEntitiesResponse represents the entity analysis response.
type AnalyzeEntitiesResponse struct {
	Entities    []Entity     `json:"entities"`
	Suggestions []Suggestion `json:"suggestions"`
}
