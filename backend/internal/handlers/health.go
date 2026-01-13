package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"vibe-backend/internal/cache"
	"vibe-backend/internal/database"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db    *database.PostgresDB
	cache *cache.RedisCache
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *database.PostgresDB, cache *cache.RedisCache) *HealthHandler {
	return &HealthHandler{
		db:    db,
		cache: cache,
	}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Services  map[string]string `json:"services,omitempty"`
}

// Health returns basic health status (for load balancer).
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   "2026-01-13-v12-clean-health",
	})
}

// Ready returns detailed readiness status (checks dependencies).
// Note: Redis is optional, so we don't fail if it's unavailable.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	dbHealthy := true

	// Check database (required)
	if h.db == nil {
		services["database"] = "unavailable: not connected"
		dbHealthy = false
	} else if err := h.db.Ping(ctx); err != nil {
		services["database"] = "unhealthy: " + err.Error()
		dbHealthy = false
	} else {
		services["database"] = "healthy"
	}

	// Check Redis (optional - don't fail if unavailable)
	if h.cache == nil {
		services["cache"] = "disabled"
	} else if err := h.cache.Ping(ctx); err != nil {
		services["cache"] = "unavailable: " + err.Error()
	} else {
		services["cache"] = "healthy"
	}

	// Only database is required for the service to be "ready"
	status := "ok"
	statusCode := http.StatusOK
	if !dbHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Services:  services,
	})
}
