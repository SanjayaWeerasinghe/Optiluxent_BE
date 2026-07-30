package finance

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW := middleware.AuditLog(auditLogger)
	canRead := middleware.RequirePermission(enforcer, "finance", "read")
	canWrite := middleware.RequirePermission(enforcer, "finance", "write")

	// Payments
	router.Post("/payments", canWrite, auditMW, h.RecordPayment)
	router.Get("/payments", canRead, h.ListPayments)
	router.Get("/payments/:id", canRead, h.GetPayment)

	// Journal entries
	router.Get("/journal-entries", canRead, h.ListJournalEntries)
	router.Get("/journal-entries/:id", canRead, h.GetJournalEntry)

	// Reports
	router.Get("/reports/aging", canRead, h.AgingReport)
	router.Get("/parties/:id/outstanding", canRead, h.PartyOutstanding)

	// Settings
	router.Get("/settings", canRead, h.GetSettings)
	router.Put("/settings", canWrite, auditMW, h.UpdateSettings)
}
