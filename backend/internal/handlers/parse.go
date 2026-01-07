package handlers

import (
	"errors"
	"net/http"

	"vibe-backend/internal/models"
	"vibe-backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ParseHandler handles URL parsing requests.
type ParseHandler struct {
	parser *services.ParserService
	log    *zap.Logger
}

// NewParseHandler creates a new ParseHandler.
func NewParseHandler(parser *services.ParserService, log *zap.Logger) *ParseHandler {
	return &ParseHandler{
		parser: parser,
		log:    log,
	}
}

// Parse handles POST /api/parse requests.
func (h *ParseHandler) Parse(c *gin.Context) {
	var req models.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ParseError{
			Code:    models.ErrorCodeInvalidURL,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	h.log.Info("Parse request received",
		zap.String("url", req.URL),
		zap.String("request_id", c.GetString("request_id")),
	)

	// Parse the URL
	content, err := h.parser.Parse(c.Request.Context(), req.URL)
	if err != nil {
		h.handleParseError(c, err)
		return
	}

	// Return response
	c.JSON(http.StatusOK, content.ToResponse())
}

// handleParseError handles parsing errors and returns appropriate HTTP responses.
func (h *ParseHandler) handleParseError(c *gin.Context, err error) {
	requestID := c.GetString("request_id")

	h.log.Error("Parse error",
		zap.Error(err),
		zap.String("request_id", requestID),
	)

	if errors.Is(err, services.ErrInvalidURL) {
		c.JSON(http.StatusBadRequest, models.ParseError{
			Code:    models.ErrorCodeInvalidURL,
			Message: "The provided URL is not a supported YouTube or Twitter link.",
		})
		return
	}

	if errors.Is(err, services.ErrParsingFailed) {
		c.JSON(http.StatusInternalServerError, models.ParseError{
			Code:    models.ErrorCodeParsingFailed,
			Message: "Unable to parse this link. Please try again later.",
		})
		return
	}

	// Generic error
	c.JSON(http.StatusInternalServerError, models.ParseError{
		Code:    models.ErrorCodeParsingFailed,
		Message: "An unexpected error occurred while parsing the URL.",
	})
}
