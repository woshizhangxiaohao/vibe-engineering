package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vibe-backend/internal/cache"
	"vibe-backend/internal/config"
	"vibe-backend/internal/database"
	"vibe-backend/internal/models"
	"vibe-backend/internal/router"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Initialize logger
	log := initLogger(cfg)
	defer log.Sync()

	log.Info("Starting server",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
	)

	// Initialize database with retry logic
	var db *database.PostgresDB
	var redisCache *cache.RedisCache
	
	// Try to connect to database with retries
	maxRetries := 5
	retryDelay := 2 * time.Second
	for i := 0; i < maxRetries; i++ {
		var err error
		db, err = database.NewPostgres(cfg.DatabaseURL, log)
		if err == nil {
			// Auto-migrate models
			if err := db.DB.AutoMigrate(
				&models.Pomodoro{},
				&models.VideoAnalysis{},
				&models.Chapter{},
				&models.Transcription{},
				&models.KeyPoint{},
			); err != nil {
				log.Error("Failed to auto-migrate database", zap.Error(err))
				db.Close()
				db = nil
			} else {
				log.Info("Database migration completed")
				break
			}
		}
		if i < maxRetries-1 {
			log.Warn("Failed to connect to database, retrying...",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
			)
			time.Sleep(retryDelay)
		} else {
			log.Error("Failed to connect to database after retries, continuing without database",
				zap.Error(err),
			)
		}
	}
	// Try to connect to Redis with retries
	for i := 0; i < maxRetries; i++ {
		var err error
		redisCache, err = cache.NewRedis(cfg.RedisURL, log)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			log.Warn("Failed to connect to Redis, retrying...",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
			)
			time.Sleep(retryDelay)
		} else {
			log.Error("Failed to connect to Redis after retries, continuing without Redis",
				zap.Error(err),
			)
		}
	}

	// Setup cleanup
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				log.Error("Error closing database", zap.Error(err))
			}
		}()
	}
	if redisCache != nil {
		defer func() {
			if err := redisCache.Close(); err != nil {
				log.Error("Error closing Redis", zap.Error(err))
			}
		}()
	}

	// Initialize router
	r := router.New(cfg, db, redisCache, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server stopped")
}

// initLogger initializes the Zap logger based on environment.
func initLogger(cfg *config.Config) *zap.Logger {
	var logConfig zap.Config

	if cfg.IsProduction() {
		logConfig = zap.NewProductionConfig()
	} else {
		logConfig = zap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	switch cfg.LogLevel {
	case "debug":
		logConfig.Level.SetLevel(zap.DebugLevel)
	case "info":
		logConfig.Level.SetLevel(zap.InfoLevel)
	case "warn":
		logConfig.Level.SetLevel(zap.WarnLevel)
	case "error":
		logConfig.Level.SetLevel(zap.ErrorLevel)
	}

	log, err := logConfig.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	return log
}
