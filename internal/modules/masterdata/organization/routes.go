package organization

import (
	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/infrastructure/http/middleware"
	"erp-system/internal/domain/rbac"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes mounts all Organization routes onto the given router.
// The router must already have auth + tenant middleware applied.
func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW := middleware.AuditLog(auditLogger)
	canRead   := middleware.RequirePermission(enforcer, "masterdata", "read")
	canWrite  := middleware.RequirePermission(enforcer, "masterdata", "write")
	canDelete := middleware.RequirePermission(enforcer, "masterdata", "delete")

	// Company profile (global, single record)
	router.Get("/company",  canRead,  h.GetCompany)
	router.Put("/company",  canWrite, auditMW, h.SaveCompany)

	// Departments
	depts := router.Group("/departments")
	depts.Get("",        canRead,   h.ListDepartments)
	depts.Post("",       canWrite,  auditMW, h.CreateDepartment)
	depts.Get("/:id",    canRead,   h.GetDepartment)
	depts.Put("/:id",    canWrite,  auditMW, h.UpdateDepartment)
	depts.Delete("/:id", canDelete, auditMW, h.DeleteDepartment)

	// Fiscal years
	fy := router.Group("/fiscal-years")
	fy.Get("",            canRead,  h.ListFiscalYears)
	fy.Post("",           canWrite, auditMW, h.CreateFiscalYear)
	fy.Put("/:id/close",  canWrite, auditMW, h.CloseFiscalYear)

	// Accounting periods
	ap := router.Group("/accounting-periods")
	ap.Get("",            canRead,  h.ListAccountingPeriods)
	ap.Post("/:id/close", canWrite, auditMW, h.CloseAccountingPeriod)

	// Document sequences
	ds := router.Group("/document-sequences")
	ds.Get("",      canRead,  h.ListDocumentSequences)
	ds.Post("",     canWrite, auditMW, h.CreateDocumentSequence)
	ds.Put("/:id",  canWrite, auditMW, h.UpdateDocumentSequence)

	// Countries + states (reference data, read-only via API)
	router.Get("/countries",            canRead, h.ListCountries)
	router.Get("/countries/:id/states", canRead, h.ListStates)
}
