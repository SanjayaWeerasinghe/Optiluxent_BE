package handlers

import (
	"context"
	"time"

	"erp-system/internal/infrastructure/cache/redis"
	"erp-system/internal/infrastructure/database/postgres"
	httputil "erp-system/pkg/http"
	"erp-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	logger logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(log logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: log,
	}
}

// HealthStatus represents the health status response
type HealthStatus struct {
	Status    string             `json:"status"`
	Timestamp string             `json:"timestamp"`
	Services  map[string]Service `json:"services,omitempty"`
}

// Service represents a service health status
type Service struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Basic handles the basic health check endpoint
// GET /health
func (h *HealthHandler) Basic(c *fiber.Ctx) error {
	return httputil.Success(c, "Service is healthy", HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Live handles the liveness probe endpoint
// GET /health/live
// This checks if the application is running (doesn't check dependencies)
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return httputil.Success(c, "Service is alive", HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready handles the readiness probe endpoint
// GET /health/ready
// This checks if the application is ready to serve requests (checks all dependencies)
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	services := make(map[string]Service)
	allHealthy := true

	// Check PostgreSQL
	if err := postgres.HealthCheck(ctx); err != nil {
		services["postgresql"] = Service{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		allHealthy = false
		h.logger.Error("PostgreSQL health check failed", logger.Err(err))
	} else {
		services["postgresql"] = Service{
			Status: "healthy",
		}
	}

	// Check Redis
	if err := redis.HealthCheck(ctx); err != nil {
		services["redis"] = Service{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		allHealthy = false
		h.logger.Error("Redis health check failed", logger.Err(err))
	} else {
		services["redis"] = Service{
			Status: "healthy",
		}
	}

	status := "ok"
	if !allHealthy {
		status = "degraded"
	}

	response := HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	if !allHealthy {
		return httputil.JSON(c, fiber.StatusServiceUnavailable, fiber.Map{
			"success": false,
			"data":    response,
		})
	}

	return httputil.Success(c, "Service is ready", response)
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Basic)
	router.Get("/live", h.Live)
	router.Get("/ready", h.Ready)
}
