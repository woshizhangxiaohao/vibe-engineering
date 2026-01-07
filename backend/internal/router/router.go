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

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db, cache)
	pomodoroRepo := repository.NewPomodoroRepository(db.DB)
	pomodoroHandler := handlers.NewPomodoroHandler(pomodoroRepo)
	parserService := services.NewParserService(log)
	parseHandler := handlers.NewParseHandler(parserService, log)

	// Health check routes (no auth required)
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// API routes
	api := r.Group("/api")
	{
		// Parse routes
		api.POST("/parse", parseHandler.Parse)

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
	}

	return r
}
