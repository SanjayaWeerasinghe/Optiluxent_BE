package manufacturing

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

	// Bill of Materials
	router.Get("/boms", canRead, h.ListBOMs)
	router.Post("/boms", canWrite, auditMW, h.CreateBOM)
	router.Get("/boms/:id", canRead, h.GetBOM)
	router.Put("/boms/:id", canWrite, auditMW, h.UpdateBOM)
	router.Delete("/boms/:id", canDelete, auditMW, h.DeleteBOM)
	router.Post("/boms/:bomId/lines", canWrite, auditMW, h.AddBOMLine)
	router.Delete("/bom-lines/:id", canDelete, auditMW, h.DeleteBOMLine)

	// Work Centers
	router.Get("/work-centers", canRead, h.ListWorkCenters)
	router.Post("/work-centers", canWrite, auditMW, h.CreateWorkCenter)
	router.Put("/work-centers/:id", canWrite, auditMW, h.UpdateWorkCenter)
	router.Delete("/work-centers/:id", canDelete, auditMW, h.DeleteWorkCenter)

	// Routings
	router.Get("/routings", canRead, h.ListRoutings)
	router.Post("/routings", canWrite, auditMW, h.CreateRouting)
	router.Get("/routings/:id", canRead, h.GetRouting)
	router.Delete("/routings/:id", canDelete, auditMW, h.DeleteRouting)
	router.Post("/routings/:routingId/operations", canWrite, auditMW, h.AddRoutingOperation)
	router.Delete("/routing-operations/:id", canDelete, auditMW, h.DeleteRoutingOperation)
}
