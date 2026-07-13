package masterdata

import (
	"context"
	"os"
	"strconv"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/events"
	"erp-system/internal/modules"
	"erp-system/internal/modules/masterdata/contacts"
	"erp-system/internal/modules/masterdata/documenttypes"
	"erp-system/internal/modules/masterdata/financial"
	"erp-system/internal/modules/masterdata/hr"
	"erp-system/internal/modules/masterdata/inventory"
	"erp-system/internal/modules/masterdata/manufacturing"
	"erp-system/internal/modules/masterdata/materials"
	"erp-system/internal/modules/masterdata/mmcategories"
	"erp-system/internal/modules/masterdata/mrp"
	"erp-system/internal/modules/masterdata/organization"
	"erp-system/internal/modules/masterdata/products"

	"github.com/gofiber/fiber/v2"
)

type Module struct {
	enforcer    rbac.Enforcer
	auditLogger *auditinfra.Logger

	orgService    *organization.Service
	finService    *financial.Service
	conService    *contacts.Service
	prodService   *products.Service
	invService    *inventory.Service
	mfgService    *manufacturing.Service
	hrService     *hr.Service
	mrpService    *mrp.Service
	matService    *materials.Service
	mmCatService  *mmcategories.Service
	docTypeService *documenttypes.Service
}

// DocumentTypeService exposes the doctypes service so main.go can wire it
// into procurement/inventory as a DocumentTypeResolver adapter.
func (m *Module) DocumentTypeService() *documenttypes.Service { return m.docTypeService }

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
	m.mrpService = mrp.NewService(mrp.NewRepository(deps.DB))
	m.matService = materials.NewService(materials.NewRepository(deps.DB))

	maxDepth := 5
	if v := os.Getenv("MM_CATEGORY_MAX_DEPTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxDepth = n
		}
	}
	m.mmCatService = mmcategories.NewService(mmcategories.NewRepository(deps.DB), maxDepth)
	m.docTypeService = documenttypes.NewService(documenttypes.NewRepository(deps.DB))
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
	mrp.RegisterRoutes(router.Group("/mrp"), mrp.NewHandler(m.mrpService), m.enforcer, m.auditLogger)
	materials.RegisterRoutes(router.Group("/materials"), materials.NewHandler(m.matService), m.enforcer, m.auditLogger)
	mmcategories.RegisterRoutes(router.Group("/material-categories"), mmcategories.NewHandler(m.mmCatService), m.enforcer, m.auditLogger)

	documenttypes.RegisterRoutes(router.Group("/document-types"), documenttypes.NewHandler(m.docTypeService), m.enforcer, m.auditLogger)
}

// DocTypeHandler exposes a handler instance so main.go can also mount the
// per-document value endpoints on a sibling URL group (/api/v1/documents/...).
// The code lives in this module because the storage is here, but the URL is
// intentionally not nested under /masterdata for readability.
func (m *Module) DocTypeHandler() *documenttypes.Handler {
	return documenttypes.NewHandler(m.docTypeService)
}

func (m *Module) RegisterEvents(_ events.EventBus) {}

func (m *Module) Shutdown(_ context.Context) error { return nil }
