package masterdata

import (
	"context"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/events"
	"erp-system/internal/modules"
	"erp-system/internal/modules/masterdata/contacts"
	"erp-system/internal/modules/masterdata/financial"
	"erp-system/internal/modules/masterdata/hr"
	"erp-system/internal/modules/masterdata/inventory"
	"erp-system/internal/modules/masterdata/manufacturing"
	"erp-system/internal/modules/masterdata/organization"
	"erp-system/internal/modules/masterdata/products"

	"github.com/gofiber/fiber/v2"
)

type Module struct {
	enforcer    rbac.Enforcer
	auditLogger *auditinfra.Logger

	orgService  *organization.Service
	finService  *financial.Service
	conService  *contacts.Service
	prodService *products.Service
	invService  *inventory.Service
	mfgService  *manufacturing.Service
	hrService   *hr.Service
}

func New(enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) *Module {
	return &Module{
		enforcer:    enforcer,
		auditLogger: auditLogger,
	}
}

func (m *Module) Name() string           { return "masterdata" }
func (m *Module) Dependencies() []string { return []string{} }

func (m *Module) Initialize(deps modules.Dependencies) error {
	m.orgService = organization.NewService(organization.NewRepository(deps.DB))
	m.finService = financial.NewService(financial.NewRepository(deps.DB))
	m.conService = contacts.NewService(contacts.NewRepository(deps.DB))
	m.prodService = products.NewService(products.NewRepository(deps.DB))
	m.invService = inventory.NewService(inventory.NewRepository(deps.DB))
	m.mfgService = manufacturing.NewService(manufacturing.NewRepository(deps.DB))
	m.hrService = hr.NewService(hr.NewRepository(deps.DB))
	return nil
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	organization.RegisterRoutes(router.Group("/organization"), organization.NewHandler(m.orgService), m.enforcer, m.auditLogger)
	financial.RegisterRoutes(router.Group("/financial"), financial.NewHandler(m.finService), m.enforcer, m.auditLogger)
	contacts.RegisterRoutes(router.Group("/contacts"), contacts.NewHandler(m.conService), m.enforcer, m.auditLogger)
	products.RegisterRoutes(router.Group("/products"), products.NewHandler(m.prodService), m.enforcer, m.auditLogger)
	inventory.RegisterRoutes(router.Group("/inventory"), inventory.NewHandler(m.invService), m.enforcer, m.auditLogger)
	manufacturing.RegisterRoutes(router.Group("/manufacturing"), manufacturing.NewHandler(m.mfgService), m.enforcer, m.auditLogger)
	hr.RegisterRoutes(router.Group("/hr"), hr.NewHandler(m.hrService), m.enforcer, m.auditLogger)
}

func (m *Module) RegisterEvents(_ events.EventBus) {}

func (m *Module) Shutdown(_ context.Context) error { return nil }
