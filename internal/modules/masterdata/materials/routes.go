package materials

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

	// Core
	router.Get("/", canRead, h.List)
	router.Post("/", canWrite, auditMW, h.Create)
	router.Get("/:id", canRead, h.Get)
	router.Put("/:id", canWrite, auditMW, h.Update)
	router.Delete("/:id", canDelete, auditMW, h.Delete)

	// Purchasing
	router.Get("/:id/purchasing", canRead, h.GetPurchasing)
	router.Put("/:id/purchasing", canWrite, auditMW, h.UpsertPurchasing)

	// Manufacturing
	router.Get("/:id/manufacturing", canRead, h.GetManufacturing)
	router.Put("/:id/manufacturing", canWrite, auditMW, h.UpsertManufacturing)

	// Warehouse
	router.Get("/:id/warehouse", canRead, h.GetWarehouse)
	router.Put("/:id/warehouse", canWrite, auditMW, h.UpsertWarehouse)

	// Vendors
	router.Get("/:id/vendors", canRead, h.ListVendors)
	router.Post("/:id/vendors", canWrite, auditMW, h.AddVendor)
	router.Put("/:id/vendors/:vid", canWrite, auditMW, h.UpdateVendor)
	router.Delete("/:id/vendors/:vid", canDelete, auditMW, h.DeleteVendor)
	router.Post("/:id/vendors/replace", canWrite, auditMW, h.ReplaceVendors)

	// Measurements
	router.Get("/:id/measurements", canRead, h.ListMeasurements)
	router.Post("/:id/measurements", canWrite, auditMW, h.AddMeasurement)
	router.Put("/:id/measurements/:mid", canWrite, auditMW, h.UpdateMeasurement)
	router.Delete("/:id/measurements/:mid", canDelete, auditMW, h.DeleteMeasurement)
	router.Post("/:id/measurements/replace", canWrite, auditMW, h.ReplaceMeasurements)
}
