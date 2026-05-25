package products

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

	// Categories
	router.Get("/categories", canRead, h.ListCategories)
	router.Post("/categories", canWrite, auditMW, h.CreateCategory)
	router.Put("/categories/:id", canWrite, auditMW, h.UpdateCategory)
	router.Delete("/categories/:id", canDelete, auditMW, h.DeleteCategory)

	// UOMs
	router.Get("/uoms", canRead, h.ListUOMs)
	router.Post("/uoms", canWrite, auditMW, h.CreateUOM)
	router.Put("/uoms/:id", canWrite, auditMW, h.UpdateUOM)

	// Products
	router.Get("/", canRead, h.ListProducts)
	router.Post("/", canWrite, auditMW, h.CreateProduct)
	router.Get("/:id", canRead, h.GetProduct)
	router.Put("/:id", canWrite, auditMW, h.UpdateProduct)
	router.Delete("/:id", canDelete, auditMW, h.DeleteProduct)

	// Variants (nested under product)
	router.Get("/:productId/variants", canRead, h.ListVariants)
	router.Post("/:productId/variants", canWrite, auditMW, h.CreateVariant)

	// Prices (nested under product)
	router.Get("/:productId/prices", canRead, h.ListPrices)
	router.Post("/:productId/prices", canWrite, auditMW, h.CreatePrice)
	router.Delete("/prices/:id", canDelete, auditMW, h.DeletePrice)
}
