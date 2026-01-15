package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"vibe-backend/internal/models"
	"vibe-backend/internal/services"
)

// ChatHandler handles chat-related HTTP requests.
type ChatHandler struct {
	chatService *services.ChatService
	log         *zap.Logger
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(chatService *services.ChatService, log *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		log:         log,
	}
}

// Chat handles POST /api/v1/insights/:id/chat - streaming chat
func (h *ChatHandler) Chat(c *gin.Context) {
	// Parse analysis ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid analysis ID",
			"request_id": c.GetString("request_id"),
		})
		return
	}

	// Parse request body
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		})
		return
	}

	// Start streaming
	stream, err := h.chatService.ChatStream(c.Request.Context(), uint(id), req.Message, req.HighlightID)
	if err != nil {
		h.log.Error("Failed to start chat stream", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Stream response
	c.Stream(func(w io.Writer) bool {
		event, ok := <-stream
		if !ok {
			return false
		}

		c.SSEvent("message", event)
		return !event.Done
	})
}

// GetHistory handles GET /api/v1/insights/:id/chat - get chat history
func (h *ChatHandler) GetHistory(c *gin.Context) {
	// Parse analysis ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid analysis ID",
			"request_id": c.GetString("request_id"),
		})
		return
	}

	history, err := h.chatService.GetChatHistory(c.Request.Context(), uint(id))
	if err != nil {
		h.log.Error("Failed to get chat history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		})
		return
	}

	c.JSON(http.StatusOK, history)
}

// AnalyzeEntities handles POST /api/v1/insights/:id/analyze-entities
func (h *ChatHandler) AnalyzeEntities(c *gin.Context) {
	// Parse analysis ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid analysis ID",
			"request_id": c.GetString("request_id"),
		})
		return
	}

	result, err := h.chatService.AnalyzeEntities(c.Request.Context(), uint(id))
	if err != nil {
		h.log.Error("Failed to analyze entities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
