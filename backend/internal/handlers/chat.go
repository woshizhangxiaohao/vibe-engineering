package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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
	requestID := c.GetString("request_id")
	
	// Parse insight ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.log.Warn("Invalid insight ID format",
			zap.String("insight_id", idStr),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      "INVALID_ID",
			Message:   "Invalid insight ID format.",
			RequestID: requestID,
		})
		return
	}

	// Parse request body
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid request body",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      "INVALID_REQUEST",
			Message:   fmt.Sprintf("Invalid request format: %v", err),
			RequestID: requestID,
		})
		return
	}

	// Start streaming
	stream, err := h.chatService.ChatStream(c.Request.Context(), uint(id), req.Message, req.HighlightID)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			h.log.Error("Insight not found",
				zap.String("error_code", "INSIGHT_NOT_FOUND"),
				zap.Uint64("insight_id", id),
				zap.String("request_id", requestID),
			)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:      "INSIGHT_NOT_FOUND",
				Message:   "Insight not found.",
				RequestID: requestID,
			})
			return
		}

		h.log.Error("Failed to start chat stream",
			zap.String("error_code", "INTERNAL_SERVER_ERROR"),
			zap.Uint64("insight_id", id),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      "INTERNAL_SERVER_ERROR",
			Message:   "Failed to start chat stream.",
			RequestID: requestID,
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
	requestID := c.GetString("request_id")
	
	// Parse insight ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.log.Warn("Invalid insight ID format",
			zap.String("insight_id", idStr),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      "INVALID_ID",
			Message:   "Invalid insight ID format.",
			RequestID: requestID,
		})
		return
	}

	history, err := h.chatService.GetChatHistory(c.Request.Context(), uint(id))
	if err != nil {
		h.log.Error("Failed to get chat history",
			zap.String("error_code", "INTERNAL_SERVER_ERROR"),
			zap.Uint64("insight_id", id),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      "INTERNAL_SERVER_ERROR",
			Message:   "Failed to get chat history.",
			RequestID: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, history)
}

// AnalyzeEntities handles POST /api/v1/insights/:id/analyze-entities
func (h *ChatHandler) AnalyzeEntities(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	// Parse insight ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.log.Warn("Invalid insight ID format",
			zap.String("insight_id", idStr),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      "INVALID_ID",
			Message:   "Invalid insight ID format.",
			RequestID: requestID,
		})
		return
	}

	result, err := h.chatService.AnalyzeEntities(c.Request.Context(), uint(id))
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			h.log.Error("Insight not found",
				zap.String("error_code", "INSIGHT_NOT_FOUND"),
				zap.Uint64("insight_id", id),
				zap.String("request_id", requestID),
			)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:      "INSIGHT_NOT_FOUND",
				Message:   "Insight not found.",
				RequestID: requestID,
			})
			return
		}

		// Other errors (AI service, parsing, etc.)
		h.log.Error("Failed to analyze entities",
			zap.String("error_code", "INTERNAL_SERVER_ERROR"),
			zap.Uint64("insight_id", id),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      "INTERNAL_SERVER_ERROR",
			Message:   "Failed to analyze entities.",
			RequestID: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
