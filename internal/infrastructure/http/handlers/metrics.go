package handlers

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler handles Prometheus metrics endpoint
type MetricsHandler struct {
	registry *prometheus.Registry
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	// Create a new Prometheus registry
	registry := prometheus.NewRegistry()

	// Register default collectors
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return &MetricsHandler{
		registry: registry,
	}
}

// Handler returns the Prometheus metrics handler
func (m *MetricsHandler) Handler() fiber.Handler {
	// Create Prometheus HTTP handler
	promHandler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	// Adapt the Prometheus handler to Fiber
	return adaptor.HTTPHandler(promHandler)
}

// GetRegistry returns the Prometheus registry for registering custom metrics
func (m *MetricsHandler) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RegisterRoutes registers metrics routes
func (m *MetricsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", m.Handler())
}
