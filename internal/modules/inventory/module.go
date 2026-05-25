package inventory

import (
	"context"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/events"
	"erp-system/internal/modules"

	"github.com/gofiber/fiber/v2"
)

// Module is the inventory ERP module, managing stock movements, quality checks,
// and material requests.
type Module struct {
	enforcer    rbac.Enforcer
	auditLogger *auditinfra.Logger
	svc         *Service
}

func New(enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) *Module {
	return &Module{enforcer: enforcer, auditLogger: auditLogger}
}

func (m *Module) Name() string           { return "inventory" }
func (m *Module) Dependencies() []string { return []string{"masterdata"} }

func (m *Module) Initialize(deps modules.Dependencies) error {
	m.svc = NewService(NewRepository(deps.DB, deps.LedgerDB))
	return nil
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	RegisterRoutes(router, NewHandler(m.svc), m.enforcer, m.auditLogger)
}

func (m *Module) RegisterEvents(_ events.EventBus) {
	// No event subscriptions for inventory at this time.
}

func (m *Module) Shutdown(_ context.Context) error { return nil }
