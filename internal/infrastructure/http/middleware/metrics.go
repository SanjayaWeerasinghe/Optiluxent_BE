package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics returns a Fiber middleware that records HTTP request metrics.
func PrometheusMetrics(registry *prometheus.Registry) fiber.Handler {
	httpRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeRequests := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Number of currently in-flight HTTP requests.",
	})

	registry.MustRegister(httpRequests, httpDuration, activeRequests)

	return func(c *fiber.Ctx) error {
		start := time.Now()
		activeRequests.Inc()
		defer activeRequests.Dec()

		err := c.Next()

		path := c.Route().Path
		method := c.Method()
		status := strconv.Itoa(c.Response().StatusCode())
		elapsed := time.Since(start).Seconds()

		httpRequests.WithLabelValues(method, path, status).Inc()
		httpDuration.WithLabelValues(method, path).Observe(elapsed)

		return err
	}
}
