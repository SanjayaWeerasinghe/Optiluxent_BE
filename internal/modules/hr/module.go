package hr

import (
	"context"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/events"
	"erp-system/internal/modules"

	"github.com/gofiber/fiber/v2"
)

// Module wires the HR service (family, emergency contacts, attendance,
// salary history, CV uploads) into the app. Existing employee CRUD
// stays in masterdata/hr — this module handles the *extensions*.
type Module struct {
	enforcer    rbac.Enforcer
	auditLogger *auditinfra.Logger
	svc         *Service
}

func New(enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) *Module {
	return &Module{enforcer: enforcer, auditLogger: auditLogger}
}

func (m *Module) Name() string           { return "hr" }
func (m *Module) Dependencies() []string { return []string{"masterdata"} }

func (m *Module) Service() *Service { return m.svc }

func (m *Module) Initialize(deps modules.Dependencies) error {
	m.svc = NewService(NewRepository(deps.DB))
	// Create the upload directory root at boot so the first CV upload
	// doesn't race a missing folder. Non-fatal if it fails — the handler
	// re-creates the per-employee subdirectory anyway.
	_, _ = EnsureUploadDir()
	return nil
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	RegisterRoutes(router, NewHandler(m.svc), m.enforcer, m.auditLogger)
}

func (m *Module) RegisterEvents(_ events.EventBus) {}

func (m *Module) Shutdown(_ context.Context) error { return nil }
