package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"erp-system/internal/infrastructure/config"
	"erp-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Server represents the HTTP server
type Server struct {
	app    *fiber.App
	config *config.Config
	logger logger.Logger
}

// New creates a new HTTP server instance
func New(cfg *config.Config, log logger.Logger) *Server {
	// Create Fiber app with custom configuration
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ServerHeader: "ERP-System",

		// Body limits
		BodyLimit: cfg.Server.BodyLimit,

		// Read/Write timeouts
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,

		// JSON encoder/decoder
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,

		// Error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a fiber.*Error
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Log the error
			log.Error("HTTP error",
				logger.String("path", c.Path()),
				logger.String("method", c.Method()),
				logger.Int("status", code),
				logger.Err(err))

			// Send custom error response
			return c.Status(code).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    code,
					"message": err.Error(),
				},
			})
		},

		// Disable startup message (we'll log it ourselves)
		DisableStartupMessage: true,

		// Enable prefork for production (use all CPU cores)
		Prefork: cfg.Server.Prefork,

		// Case sensitive routing
		CaseSensitive: true,

		// Strict routing (trailing slash matters)
		StrictRouting: true,

		// Enable printing routes on startup
		EnablePrintRoutes: cfg.App.Debug,
	})

	// Set up global middleware (order matters!)
	setupGlobalMiddleware(app, cfg, log)

	return &Server{
		app:    app,
		config: cfg,
		logger: log,
	}
}

// setupGlobalMiddleware sets up middleware that applies to all routes
func setupGlobalMiddleware(app *fiber.App, cfg *config.Config, log logger.Logger) {
	// Recover from panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.App.Debug,
	}))

	// Compression
	if cfg.Server.EnableCompression {
		app.Use(compress.New(compress.Config{
			Level: compress.LevelDefault, // Default compression level
		}))
	}
}

// App returns the Fiber app instance
func (s *Server) App() *fiber.App {
	return s.app
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Get server address
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	// Log server startup
	s.logger.Info("Starting HTTP server",
		logger.String("host", s.config.Server.Host),
		logger.Int("port", s.config.Server.Port),
		logger.String("environment", s.config.App.Environment),
		logger.Bool("debug", s.config.App.Debug))

	// Start server
	if err := s.app.Listen(addr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// StartWithGracefulShutdown starts the server and handles graceful shutdown
func (s *Server) StartWithGracefulShutdown() error {
	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		serverErrors <- s.Start()
	}()

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		s.logger.Info("Shutting down server",
			logger.String("signal", sig.String()))

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := s.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		s.logger.Info("Server stopped gracefully")
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Create a channel to signal when shutdown is complete
	done := make(chan error, 1)

	// Perform shutdown in goroutine
	go func() {
		done <- s.app.ShutdownWithContext(ctx)
	}()

	// Wait for shutdown to complete or context to timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
}

// GetPort returns the server port
func (s *Server) GetPort() int {
	return s.config.Server.Port
}
