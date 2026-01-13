package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"vibe-backend/internal/models"
	"vibe-backend/internal/services"
)

// TranscriptHandler handles transcript extraction endpoints.
type TranscriptHandler struct {
	transcriptService *services.TranscriptService
	log               *zap.Logger
}

// NewTranscriptHandler creates a new TranscriptHandler.
func NewTranscriptHandler(transcriptService *services.TranscriptService, log *zap.Logger) *TranscriptHandler {
	return &TranscriptHandler{
		transcriptService: transcriptService,
		log:               log,
	}
}

// GetTranscript fetches transcript for a YouTube video.
// POST /api/v1/transcript
func (h *TranscriptHandler) GetTranscript(c *gin.Context) {
	var req struct {
		Input string `json:"input" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "URL 或视频 ID 不能为空",
		})
		return
	}

	// Get transcript
	response, err := h.transcriptService.GetTranscript(c.Request.Context(), req.Input)
	if err != nil {
		h.log.Error("Failed to get transcript",
			zap.Error(err),
			zap.String("input", req.Input),
		)

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
