package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"vibe-backend/internal/models"
	"vibe-backend/internal/services"
)

// YouTubeAPIHandler handles YouTube API endpoints.
type YouTubeAPIHandler struct {
	youtubeAPI   *services.YouTubeAPIService
	youtubeService *services.YouTubeService
	oauthService *services.OAuthService
	log          *zap.Logger
}

// NewYouTubeAPIHandler creates a new YouTubeAPIHandler.
func NewYouTubeAPIHandler(youtubeAPI *services.YouTubeAPIService, youtubeService *services.YouTubeService, oauthService *services.OAuthService, log *zap.Logger) *YouTubeAPIHandler {
	return &YouTubeAPIHandler{
		youtubeAPI:   youtubeAPI,
		youtubeService: youtubeService,
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

// RefreshToken refreshes an expired OAuth token.
// POST /api/v1/auth/google/refresh
func (h *YouTubeAPIHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    models.ErrorInvalidInput,
			Message: "Refresh token 不能为空",
		})
		return
	}

	// Refresh the token using OAuth service
	token, err := h.oauthService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.log.Error("Failed to refresh token",
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    models.ErrorAuthFailed,
			Message: "Token 刷新失败，请重新登录",
		})
		return
	}

	// Convert token to JSON for storage
	tokenJSON, err := h.oauthService.TokenToJSON(token)
	if err != nil {
		h.log.Error("Failed to serialize refreshed token",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    models.ErrorAuthFailed,
			Message: "Token 刷新失败，请重新登录",
		})
		return
	}

	// Return new token to client
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
	// #region agent log
	logDebug("youtube_api.go:274", "GetCaptions entry", map[string]interface{}{
		"videoId": c.Query("videoId"),
		"hasAuthHeader": c.GetHeader("Authorization") != "",
		"youtubeServiceNil": h.youtubeService == nil,
		"youtubeAPINil": h.youtubeAPI == nil,
		"method": c.Request.Method,
		"path": c.Request.URL.Path,
	}, "A,C,D")
	// #endregion

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

	// #region agent log
	logDebug("youtube_api.go:293", "Before GetCaptions API call", map[string]interface{}{
		"videoId": videoID,
		"hasToken": token != nil,
		"tokenLength": func() int {
			if token != nil {
				return len(token.AccessToken)
			}
			return 0
		}(),
	}, "B,E")
	// #endregion

	// Try YouTube Data API v3 first (only if we have a token)
	var response *models.YouTubeCaptionsResponse
	var err error
	if token != nil && token.AccessToken != "" {
		response, err = h.youtubeAPI.GetCaptions(c.Request.Context(), videoID, token)
		if err != nil {
			// #region agent log
			logDebug("youtube_api.go:296", "GetCaptions API error", map[string]interface{}{
				"videoId": videoID,
				"error": err.Error(),
				"isUnauthorized": isUnauthorizedError(err),
			}, "B,E")
			// #endregion

			h.log.Warn("YouTube Data API v3 failed, trying fallback method",
				zap.Error(err),
				zap.String("video_id", videoID),
			)

			// Only return 401 for actual authorization failures (not missing captions)
			// If it's a NO_CAPTIONS error, we should try fallback
			if isUnauthorizedError(err) && !contains(err.Error(), "NO_CAPTIONS") {
				// #region agent log
				logDebug("youtube_api.go:303", "Unauthorized error detected, returning 401", map[string]interface{}{
					"videoId": videoID,
					"error": err.Error(),
				}, "B")
				// #endregion
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Code:    models.ErrorUnauthorized,
					Message: "需要 OAuth 授权才能访问字幕",
				})
				return
			}
			// If it's NO_CAPTIONS or other errors, fall through to fallback
		}
	} else {
		// #region agent log
		logDebug("youtube_api.go:295", "No OAuth token, skipping API and trying fallback", map[string]interface{}{
			"videoId": videoID,
		}, "B")
		// #endregion
		h.log.Info("No OAuth token provided, using fallback method",
			zap.String("video_id", videoID),
		)
		err = fmt.Errorf("NO_CAPTIONS: no OAuth token")
	}

	// If API returns no captions or no token, try fallback method (web scraping)
	// This works even without OAuth and can access more videos
	if err != nil || response == nil || (response != nil && len(response.Captions) == 0) {
		// #region agent log
		logDebug("youtube_api.go:369", "Checking fallback availability", map[string]interface{}{
			"videoId": videoID,
			"youtubeServiceNil": h.youtubeService == nil,
			"hasError": err != nil,
			"error": func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}, "A")
		// #endregion

		if h.youtubeService != nil {
			// #region agent log
			logDebug("youtube_api.go:315", "Starting fallback method", map[string]interface{}{
				"videoId": videoID,
			}, "C")
			// #endregion

			h.log.Info("Attempting fallback caption extraction via web scraping",
				zap.String("video_id", videoID),
			)
			
			transcript, fallbackErr := h.youtubeService.FetchYouTubeTranscript(c.Request.Context(), videoID)
			
			// #region agent log
			logDebug("youtube_api.go:318", "Fallback method result", map[string]interface{}{
				"videoId": videoID,
				"hasError": fallbackErr != nil,
				"error": func() string {
					if fallbackErr != nil {
						return fallbackErr.Error()
					}
					return ""
				}(),
				"transcriptLength": len(transcript),
				"transcriptEmpty": transcript == "",
			}, "C")
			// #endregion

			if fallbackErr == nil && transcript != "" {
				// Successfully fetched transcript via web scraping
				// Create a response with a virtual caption track
				fallbackResponse := &models.YouTubeCaptionsResponse{
					Captions: []models.YouTubeCaption{
						{
							ID:       "fallback-transcript",
							Language: "auto",
							Name:     "自动生成字幕（网页抓取）",
						},
					},
				}
				h.log.Info("Successfully fetched captions via fallback method",
					zap.String("video_id", videoID),
					zap.Int("transcript_length", len(transcript)),
				)
				// #region agent log
				logDebug("youtube_api.go:335", "Fallback success, returning 200", map[string]interface{}{
					"videoId": videoID,
				}, "C")
				// #endregion
				c.JSON(http.StatusOK, fallbackResponse)
				return
			}

			h.log.Warn("Fallback method also failed",
				zap.Error(fallbackErr),
				zap.String("video_id", videoID),
			)
		} else {
			// #region agent log
			logDebug("youtube_api.go:343", "Fallback skipped - youtubeService is nil", map[string]interface{}{
				"videoId": videoID,
			}, "A")
			// #endregion
		}

		// Both methods failed
		// #region agent log
		logDebug("youtube_api.go:346", "Both methods failed, returning 404", map[string]interface{}{
			"videoId": videoID,
		}, "A,C")
		// #endregion
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:    models.ErrorNoCaptions,
			Message: "该视频未提供 API 可访问的字幕轨道",
		})
		return
	}

	// #region agent log
	logDebug("youtube_api.go:353", "API success, returning 200", map[string]interface{}{
		"videoId": videoID,
		"captionsCount": len(response.Captions),
	}, "D")
	// #endregion
	c.JSON(http.StatusOK, response)
}

// logDebug writes debug logs to file
func logDebug(location, message string, data map[string]interface{}, hypothesisIds string) {
	logEntry := map[string]interface{}{
		"location":      location,
		"message":       message,
		"data":          data,
		"timestamp":     time.Now().UnixMilli(),
		"sessionId":     "debug-session",
		"runId":         "run1",
		"hypothesisId":  hypothesisIds,
	}
	logData, _ := json.Marshal(logEntry)
	
	// Try multiple log paths: env var, mounted volume, or fallback to /tmp
	logPath := os.Getenv("DEBUG_LOG_PATH")
	if logPath == "" {
		// Try workspace path first (for local development)
		workspacePath := "/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor/debug.log"
		if _, err := os.Stat("/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor"); err == nil {
			logPath = workspacePath
		} else {
			// Fallback to /tmp (always exists in Docker)
			logPath = "/tmp/debug.log"
		}
	}
	
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		f.Write(logData)
		f.WriteString("\n")
		f.Close()
	}
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
