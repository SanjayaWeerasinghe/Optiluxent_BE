package contacts

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

	// Parties
	router.Get("/parties", canRead, h.ListParties)
	router.Post("/parties", canWrite, auditMW, h.CreateParty)
	router.Get("/parties/:id", canRead, h.GetParty)
	router.Put("/parties/:id", canWrite, auditMW, h.UpdateParty)
	router.Delete("/parties/:id", canDelete, auditMW, h.DeleteParty)

	// Contact persons (nested under party)
	router.Get("/parties/:partyId/contacts", canRead, h.ListContactPersons)
	router.Post("/parties/:partyId/contacts", canWrite, auditMW, h.CreateContactPerson)
	router.Put("/contacts/:id", canWrite, auditMW, h.UpdateContactPerson)
	router.Delete("/contacts/:id", canDelete, auditMW, h.DeleteContactPerson)

	// Addresses (nested under party)
	router.Get("/parties/:partyId/addresses", canRead, h.ListAddresses)
	router.Post("/parties/:partyId/addresses", canWrite, auditMW, h.CreateAddress)
	router.Put("/addresses/:id", canWrite, auditMW, h.UpdateAddress)
	router.Delete("/addresses/:id", canDelete, auditMW, h.DeleteAddress)
}
