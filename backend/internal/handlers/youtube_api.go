package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"vibe-backend/internal/models"
	"vibe-backend/internal/services"
)

// YouTubeAPIHandler handles YouTube API endpoints.
type YouTubeAPIHandler struct {
	youtubeAPI   *services.YouTubeAPIService
	oauthService *services.OAuthService
	log          *zap.Logger
}

// NewYouTubeAPIHandler creates a new YouTubeAPIHandler.
func NewYouTubeAPIHandler(youtubeAPI *services.YouTubeAPIService, oauthService *services.OAuthService, log *zap.Logger) *YouTubeAPIHandler {
	return &YouTubeAPIHandler{
		youtubeAPI:   youtubeAPI,
		oauthService: oauthService,
		log:          log,
	}
}

// GetAuthURL generates Google OAuth authorization URL.
// GET /api/v1/auth/google/url
func (h *YouTubeAPIHandler) GetAuthURL(c *gin.Context) {
	// Check if OAuth is configured
	if h.oauthService == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    models.ErrorAuthConfig,
			Message: "OAuth 配置缺失",
		})
		return
	}

	// Generate state token (in production, this should be stored and verified)
	state := "random-state-token" // TODO: Generate secure random state

	authURL := h.oauthService.GetAuthURL(state)

	c.JSON(http.StatusOK, models.AuthURLResponse{
		URL: authURL,
	})
}

// HandleCallback handles Google OAuth callback.
// POST /api/v1/auth/google/callback
func (h *YouTubeAPIHandler) HandleCallback(c *gin.Context) {
	var req models.OAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "无效的请求参数",
		})
		return
	}

	// Validate required fields
	if req.Code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "授权码不能为空",
		})
		return
	}

	// Exchange authorization code for access token
	token, err := h.oauthService.ExchangeCode(c.Request.Context(), req.Code)
	if err != nil {
		h.log.Error("Failed to exchange authorization code",
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    models.ErrorAuthFailed,
			Message: "授权失败，请重试",
		})
		return
	}

	// Convert token to JSON for storage
	tokenJSON, err := h.oauthService.TokenToJSON(token)
	if err != nil {
		h.log.Error("Failed to serialize token",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    models.ErrorAuthFailed,
			Message: "授权失败，请重试",
		})
		return
	}

	// Return token to client for storage
	c.JSON(http.StatusOK, models.OAuthCallbackResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		TokenJSON:    tokenJSON,
	})
}

// GetVideoMetadata fetches video metadata.
// GET /api/v1/youtube/video?input=<url-or-id>
func (h *YouTubeAPIHandler) GetVideoMetadata(c *gin.Context) {
	input := c.Query("input")
	if input == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "无效的 URL 或视频 ID",
		})
		return
	}

	response, err := h.youtubeAPI.GetVideoMetadata(c.Request.Context(), input)
	if err != nil {
		h.log.Error("Failed to get video metadata",
			zap.Error(err),
			zap.String("input", input),
		)

		// Map error to appropriate response
		if isQuotaError(err) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Code:    models.ErrorQuotaExceeded,
				Message: "服务暂时不可用",
			})
			return
		}

		if isNotFoundError(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    models.ErrorVideoNotFound,
				Message: "视频无法访问或不存在",
			})
			return
		}

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPlaylist fetches playlist items.
// GET /api/v1/youtube/playlist?playlistId=<id>
func (h *YouTubeAPIHandler) GetPlaylist(c *gin.Context) {
	playlistID := c.Query("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "播放列表 ID 不能为空",
		})
		return
	}

	// Get OAuth token from Authorization header for private playlists
	var token *oauth2.Token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken := authHeader[7:]
		token = &oauth2.Token{
			AccessToken: accessToken,
		}
	}

	response, err := h.youtubeAPI.GetPlaylist(c.Request.Context(), playlistID, token)
	if err != nil {
		h.log.Error("Failed to get playlist",
			zap.Error(err),
			zap.String("playlist_id", playlistID),
		)

		// Map error to appropriate response
		if isUnauthorizedError(err) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    models.ErrorUnauthorized,
				Message: "需要 OAuth 授权",
			})
			return
		}

		if isQuotaError(err) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Code:    models.ErrorQuotaExceeded,
				Message: "服务暂时不可用",
			})
			return
		}

		if isNotFoundError(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    models.ErrorPlaylistNotFound,
				Message: "播放列表不存在",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    models.ErrorPlaylistNotFound,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCaptions fetches video caption tracks.
// GET /api/v1/youtube/captions?videoId=<id>
func (h *YouTubeAPIHandler) GetCaptions(c *gin.Context) {
	videoID := c.Query("videoId")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "视频 ID 不能为空",
		})
		return
	}

	// Get OAuth token from Authorization header
	var token *oauth2.Token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken := authHeader[7:]
		token = &oauth2.Token{
			AccessToken: accessToken,
		}
	}

	// Captions API requires OAuth authorization
	response, err := h.youtubeAPI.GetCaptions(c.Request.Context(), videoID, token)
	if err != nil {
		h.log.Error("Failed to get captions",
			zap.Error(err),
			zap.String("video_id", videoID),
		)

		// Map error to appropriate response
		if isUnauthorizedError(err) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    models.ErrorUnauthorized,
				Message: "需要 OAuth 授权才能访问字幕",
			})
			return
		}

		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:    models.ErrorNoCaptions,
			Message: "该视频未提供 API 可访问的字幕轨道",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetQuota returns the current API quota status.
// GET /api/v1/system/quota
func (h *YouTubeAPIHandler) GetQuota(c *gin.Context) {
	response := h.youtubeAPI.GetQuotaStatus(c.Request.Context())
	c.JSON(http.StatusOK, response)
}

// Helper functions to check error types

func isQuotaError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "quota") || contains(errStr, "quotaExceeded") || contains(errStr, "QUOTA_EXCEEDED")
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "VIDEO_NOT_FOUND") || contains(errStr, "PLAYLIST_NOT_FOUND") || contains(errStr, "not found")
}

func isUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "UNAUTHORIZED") || contains(errStr, "unauthorized") || contains(errStr, "forbidden")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
