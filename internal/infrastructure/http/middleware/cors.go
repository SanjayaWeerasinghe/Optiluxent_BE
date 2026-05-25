package middleware

import (
	"strings"

	"erp-system/internal/infrastructure/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns a CORS middleware configured from the config
func CORS(cfg *config.Config) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CORS.AllowOrigins, ","),
		AllowMethods:     strings.Join(cfg.CORS.AllowMethods, ","),
		AllowHeaders:     strings.Join(cfg.CORS.AllowHeaders, ","),
		AllowCredentials: cfg.CORS.AllowCredentials,
		ExposeHeaders:    strings.Join(cfg.CORS.ExposeHeaders, ","),
		MaxAge:           cfg.CORS.MaxAge,
	})
}
