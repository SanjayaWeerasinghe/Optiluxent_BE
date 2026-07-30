package finance

import (
	"context"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/events"
	"erp-system/internal/modules"

	"github.com/gofiber/fiber/v2"
)

// Module wires the finance service into the app.
type Module struct {
	enforcer    rbac.Enforcer
	auditLogger *auditinfra.Logger
	svc         *Service
}

func New(enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) *Module {
	return &Module{enforcer: enforcer, auditLogger: auditLogger}
}

func (m *Module) Name() string           { return "finance" }
func (m *Module) Dependencies() []string { return []string{"masterdata", "procurement", "sales"} }

// Service exposes the underlying service so main.go can inject cross-module
// adapters (procurement + sales appliers, invoice lookups).
func (m *Module) Service() *Service { return m.svc }

func (m *Module) Initialize(deps modules.Dependencies) error {
	m.svc = NewService(NewRepository(deps.DB))
	return nil
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	RegisterRoutes(router, NewHandler(m.svc), m.enforcer, m.auditLogger)
}

func (m *Module) RegisterEvents(_ events.EventBus) {}

func (m *Module) Shutdown(_ context.Context) error { return nil }
