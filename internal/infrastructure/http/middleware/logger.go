package middleware

import (
	"time"

	"erp-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// Logger returns a request logging middleware
func Logger(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Get request ID from context
		requestID, _ := c.Locals("request_id").(string)

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Log request details
		fields := []logger.Field{
			logger.String("request_id", requestID),
			logger.String("method", c.Method()),
			logger.String("path", c.Path()),
			logger.Int("status", c.Response().StatusCode()),
			logger.Duration("duration", duration),
			logger.String("ip", c.IP()),
			logger.String("user_agent", c.Get("User-Agent")),
		}

		// Add error to log if present
		if err != nil {
			fields = append(fields, logger.Err(err))
			log.Error("HTTP request error", fields...)
		} else {
			// Log based on status code
			statusCode := c.Response().StatusCode()
			if statusCode >= 500 {
				log.Error("HTTP request completed with server error", fields...)
			} else if statusCode >= 400 {
				log.Warn("HTTP request completed with client error", fields...)
			} else {
				log.Info("HTTP request completed", fields...)
			}
		}

		return err
	}
}
