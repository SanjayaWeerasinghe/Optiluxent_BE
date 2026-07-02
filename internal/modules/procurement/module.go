package procurement

import (
	"context"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/events"
	"erp-system/internal/modules"

	"github.com/gofiber/fiber/v2"
)

type Module struct {
	enforcer    rbac.Enforcer
	auditLogger *auditinfra.Logger
	svc         *Service
}

func New(enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) *Module {
	return &Module{enforcer: enforcer, auditLogger: auditLogger}
}

func (m *Module) Name() string           { return "procurement" }
func (m *Module) Dependencies() []string { return []string{"masterdata", "inventory"} }

// Service returns the underlying procurement service so wiring code (main.go)
// can inject cross-module collaborators like the QC auto-creator.
func (m *Module) Service() *Service { return m.svc }

func (m *Module) Initialize(deps modules.Dependencies) error {
	m.svc = NewService(NewRepository(deps.DB, deps.LedgerDB))
	return nil
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	RegisterRoutes(router, NewHandler(m.svc), m.enforcer, m.auditLogger)
}

func (m *Module) RegisterEvents(bus events.EventBus) {
	if m.svc != nil {
		m.svc.SetEventBus(bus)
	}
}

func (m *Module) Shutdown(_ context.Context) error { return nil }
