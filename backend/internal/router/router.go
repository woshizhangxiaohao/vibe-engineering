package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"vibe-backend/internal/cache"
	"vibe-backend/internal/config"
	"vibe-backend/internal/database"
	"vibe-backend/internal/handlers"
	"vibe-backend/internal/middleware"
	"vibe-backend/internal/repository"
	"vibe-backend/internal/services"
)

// New creates and configures a new Gin router.
func New(cfg *config.Config, db *database.PostgresDB, cache *cache.RedisCache, log *zap.Logger) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	r := gin.New()

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Initialize health handler first (always available)
	healthHandler := handlers.NewHealthHandler(db, cache)

	// Health check routes (no auth required) - always available
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// Only register other routes if database is available
	if db == nil {
		log.Warn("Database not available, only health check endpoints are registered")
		return r
	}

	// Initialize other handlers (require database)
	pomodoroRepo := repository.NewPomodoroRepository(db.DB)
	pomodoroHandler := handlers.NewPomodoroHandler(pomodoroRepo)
	parserService := services.NewParserService(log)
	parseHandler := handlers.NewParseHandler(parserService, log)

	// Analysis handlers
	analysisRepo := repository.NewAnalysisRepository(db.DB)
	analysisHandler := handlers.NewAnalysisHandler(analysisRepo, log)

	// YouTube video analysis handlers
	videoRepo := repository.NewVideoRepository(db.DB)
	youtubeService := services.NewYouTubeService(cfg.OpenRouterAPIKey, cfg.GeminiModel, log)
	videoHandler := handlers.NewVideoHandler(videoRepo, youtubeService, log)

	// InsightFlow handlers
	insightRepo := repository.NewInsightRepository(db.DB)
	insightProcessor := services.NewInsightProcessor(insightRepo, youtubeService, log)
	insightHandler := handlers.NewInsightHandler(insightRepo, insightProcessor, log)

	// YouTube Data API v3 handlers (OAuth + API endpoints)
	oauthService := services.NewOAuthService(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL, log)
	youtubeAPIService := services.NewYouTubeAPIService(cfg.YouTubeAPIKey, cache, oauthService, log)
	youtubeAPIHandler := handlers.NewYouTubeAPIHandler(youtubeAPIService, youtubeService, oauthService, log)

	// Transcript service (yt-dlp based subtitle extraction)
	transcriptService := services.NewTranscriptService(log)
	transcriptHandler := handlers.NewTranscriptHandler(transcriptService, log)

	// Chat handlers
	chatRepo := repository.NewChatRepository(db.DB)
	chatService := services.NewChatService(chatRepo, videoRepo, insightRepo, cfg.OpenRouterAPIKey, cfg.GeminiModel, log)
	chatHandler := handlers.NewChatHandler(chatService, log)

	// API routes
	api := r.Group("/api")
	{
		// Parse routes
		api.POST("/parse", parseHandler.Parse)

		// Analysis routes
		analysis := api.Group("/analysis")
		{
			analysis.GET("", analysisHandler.List)
			analysis.POST("", analysisHandler.Create)
			analysis.GET("/:id", analysisHandler.Get)
			analysis.PATCH("/:id", analysisHandler.Update)
			analysis.DELETE("/:id", analysisHandler.Delete)
		}

		// Pomodoro routes
		pomodoros := api.Group("/pomodoros")
		{
			pomodoros.GET("", pomodoroHandler.List)
			pomodoros.POST("", pomodoroHandler.Create)
			pomodoros.GET("/:id", pomodoroHandler.Get)
			pomodoros.PATCH("/:id", pomodoroHandler.Update)
			pomodoros.DELETE("/:id", pomodoroHandler.Delete)
			pomodoros.POST("/:id/complete", pomodoroHandler.Complete)
		}

		// YouTube video analysis routes (API v1)
		v1 := api.Group("/v1")
		{
			// Video routes
			videos := v1.Group("/videos")
			{
				videos.POST("/metadata", videoHandler.GetMetadata)
				videos.POST("/analyze", videoHandler.AnalyzeVideo)
				videos.GET("/result/:jobId", videoHandler.GetResult)
				videos.POST("/export", videoHandler.ExportVideo)
			}

			// History routes
			v1.GET("/history", videoHandler.GetHistory)

			// YouTube Data API v3 routes
			// OAuth 2.0 authentication endpoints
			auth := v1.Group("/auth")
			{
				auth.GET("/google/url", youtubeAPIHandler.GetAuthURL)
				auth.POST("/google/callback", youtubeAPIHandler.HandleCallback)
				auth.POST("/google/refresh", youtubeAPIHandler.RefreshToken)
			}

			youtube := v1.Group("/youtube")
			{
				youtube.GET("/video", youtubeAPIHandler.GetVideoMetadata)
				youtube.GET("/playlist", youtubeAPIHandler.GetPlaylist)
				youtube.GET("/captions", youtubeAPIHandler.GetCaptions)
			}

			system := v1.Group("/system")
			{
				system.GET("/quota", youtubeAPIHandler.GetQuota)
			}

			// Transcript extraction endpoint (yt-dlp based)
			v1.POST("/transcript", transcriptHandler.GetTranscript)

			// InsightFlow routes
			insights := v1.Group("/insights")
			{
				insights.GET("", insightHandler.List)
				insights.POST("", insightHandler.Create)
				insights.GET("/:id", insightHandler.Get)
				insights.PATCH("/:id", insightHandler.Update)
				insights.DELETE("/:id", insightHandler.Delete)
				insights.POST("/:id/process", insightHandler.Process)

				// Highlight routes
				insights.GET("/:id/highlights", insightHandler.ListHighlights)
				insights.POST("/:id/highlights", insightHandler.CreateHighlight)
				insights.PATCH("/:id/highlights/:highlightId", insightHandler.UpdateHighlight)
				insights.DELETE("/:id/highlights/:highlightId", insightHandler.DeleteHighlight)

				// Chat routes (InsightHandler)
				insights.GET("/:id/chat", insightHandler.ListChatMessages)
				insights.POST("/:id/chat", insightHandler.CreateChatMessage)
				insights.DELETE("/:id/chat", insightHandler.ClearChatHistory)

				// Entity analysis route (ChatHandler)
				insights.POST("/:id/analyze-entities", chatHandler.AnalyzeEntities)
			}

			// Shared insight (public access)
			v1.GET("/shared/:token", insightHandler.GetShared)
		}
	}

	return r
}
