package inventory

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

	// Warehouses
	router.Get("/warehouses", canRead, h.ListWarehouses)
	router.Post("/warehouses", canWrite, auditMW, h.CreateWarehouse)
	router.Get("/warehouses/:id", canRead, h.GetWarehouse)
	router.Put("/warehouses/:id", canWrite, auditMW, h.UpdateWarehouse)
	router.Delete("/warehouses/:id", canDelete, auditMW, h.DeleteWarehouse)

	// Storage locations (nested under warehouse)
	router.Get("/warehouses/:warehouseId/locations", canRead, h.ListLocations)
	router.Post("/warehouses/:warehouseId/locations", canWrite, auditMW, h.CreateLocation)
	router.Put("/locations/:id", canWrite, auditMW, h.UpdateLocation)

	// Stock ledger
	router.Get("/stock/balance", canRead, h.GetStockBalance)
	router.Get("/stock/ledger", canRead, h.ListStockLedger)
	router.Post("/stock/entries", canWrite, auditMW, h.CreateStockEntry)
}
