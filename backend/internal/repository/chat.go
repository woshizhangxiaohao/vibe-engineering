package repository

import (
	"context"

	"gorm.io/gorm"
	"vibe-backend/internal/models"
)

// ChatRepository handles database operations for chat messages.
type ChatRepository struct {
	db *gorm.DB
}

// NewChatRepository creates a new ChatRepository.
func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateMessage creates a new chat message.
func (r *ChatRepository) CreateMessage(ctx context.Context, message *models.ChatMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetMessagesByAnalysisID returns all chat messages for an analysis.
func (r *ChatRepository) GetMessagesByAnalysisID(ctx context.Context, analysisID uint) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.WithContext(ctx).
		Where("analysis_id = ?", analysisID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// GetMessageByID returns a chat message by ID.
func (r *ChatRepository) GetMessageByID(ctx context.Context, id uint) (*models.ChatMessage, error) {
	var message models.ChatMessage
	err := r.db.WithContext(ctx).First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// DeleteMessagesByAnalysisID deletes all chat messages for an analysis.
func (r *ChatRepository) DeleteMessagesByAnalysisID(ctx context.Context, analysisID uint) error {
	return r.db.WithContext(ctx).
		Where("analysis_id = ?", analysisID).
		Delete(&models.ChatMessage{}).Error
}
