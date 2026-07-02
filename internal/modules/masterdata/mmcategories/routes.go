package mmcategories

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/infrastructure/http/middleware"
	"erp-system/internal/domain/rbac"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW   := middleware.AuditLog(auditLogger)
	canRead   := middleware.RequirePermission(enforcer, "masterdata", "read")
	canWrite  := middleware.RequirePermission(enforcer, "masterdata", "write")
	canDelete := middleware.RequirePermission(enforcer, "masterdata", "delete")

	router.Get("/config", canRead, h.Config)
	router.Get("/", canRead, h.List)
	router.Post("/", canWrite, auditMW, h.Create)
	router.Put("/:id", canWrite, auditMW, h.Update)
	router.Delete("/:id", canDelete, auditMW, h.Delete)
}
