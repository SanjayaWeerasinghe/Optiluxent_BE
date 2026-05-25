package modules

import (
	"context"
	"database/sql"

	"erp-system/internal/infrastructure/events"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Dependencies bundles the shared infrastructure available to every module.
type Dependencies struct {
	DB       *gorm.DB
	LedgerDB *sql.DB
	EventBus events.EventBus
}

// Module is the interface every ERP module must satisfy.
type Module interface {
	// Name returns the unique module identifier (e.g. "masterdata", "sales").
	Name() string

	// Dependencies returns names of modules this module depends on.
	Dependencies() []string

	// Initialize performs any setup needed before routes are registered.
	Initialize(deps Dependencies) error

	// RegisterRoutes mounts the module's HTTP routes on the provided router.
	RegisterRoutes(router fiber.Router)

	// RegisterEvents subscribes to domain events via the event bus.
	RegisterEvents(bus events.EventBus)

	// Shutdown performs cleanup on application shutdown.
	Shutdown(ctx context.Context) error
}
