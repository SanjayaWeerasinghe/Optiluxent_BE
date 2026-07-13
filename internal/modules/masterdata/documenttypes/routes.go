package documenttypes

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

// RegisterRoutes wires the document-types CRUD under
// /api/v1/masterdata/document-types.
func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW   := middleware.AuditLog(auditLogger)
	canRead   := middleware.RequirePermission(enforcer, "masterdata", "read")
	canWrite  := middleware.RequirePermission(enforcer, "masterdata", "write")
	canDelete := middleware.RequirePermission(enforcer, "masterdata", "delete")

	// Types
	router.Get("",     canRead,   h.ListTypes)
	router.Post("",    canWrite,  auditMW, h.CreateType)
	router.Get("/:id", canRead,   h.GetType)
	router.Put("/:id", canWrite,  auditMW, h.UpdateType)
	router.Delete("/:id", canDelete, auditMW, h.DeleteType)

	// Fields
	router.Get("/:id/fields",              canRead,  h.ListFields)
	router.Post("/:id/fields",             canWrite, auditMW, h.CreateField)
	router.Put("/:id/fields/:fieldId",     canWrite, auditMW, h.UpdateField)
	router.Delete("/:id/fields/:fieldId",  canDelete, auditMW, h.DeleteField)
}

// RegisterDocValueRoutes wires the per-document value endpoints under
// /api/v1/documents/:kind/:id/fields. It's mounted separately so the URL
// reads naturally (documents, not masterdata/documents).
func RegisterDocValueRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW  := middleware.AuditLog(auditLogger)
	canRead  := middleware.RequirePermission(enforcer, "masterdata", "read")
	canWrite := middleware.RequirePermission(enforcer, "masterdata", "write")

	router.Get("/:kind/:id/fields",  canRead,  h.GetDocValues)
	router.Post("/:kind/:id/fields", canWrite, auditMW, h.UpsertDocValues)
}
