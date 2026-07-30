package hr

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW := middleware.AuditLog(auditLogger)
	// HR permissions borrow from the masterdata namespace for now — same
	// audience (HR ops) as employee master data. Split into a dedicated
	// "hr" domain later if role separation calls for it.
	canRead := middleware.RequirePermission(enforcer, "masterdata", "read")
	canWrite := middleware.RequirePermission(enforcer, "masterdata", "write")
	canDelete := middleware.RequirePermission(enforcer, "masterdata", "delete")

	// Cross-employee attendance list (for the "Attendance" tab on the HR
	// dashboard). Filters via ?employee_id=&from=&to=.
	router.Get("/attendance", canRead, h.ListAttendance)

	// Everything else is nested under an employee for a natural REST shape.
	emp := router.Group("/employees/:id")

	// Family
	emp.Get("/family", canRead, h.ListFamily)
	emp.Post("/family", canWrite, auditMW, h.AddFamily)
	emp.Put("/family/:familyId", canWrite, auditMW, h.UpdateFamily)
	emp.Delete("/family/:familyId", canDelete, auditMW, h.DeleteFamily)

	// Emergency contacts
	emp.Get("/emergency", canRead, h.ListEmergency)
	emp.Post("/emergency", canWrite, auditMW, h.AddEmergency)
	emp.Put("/emergency/:contactId", canWrite, auditMW, h.UpdateEmergency)
	emp.Delete("/emergency/:contactId", canDelete, auditMW, h.DeleteEmergency)

	// Attendance (nested list is useful for the per-employee tab)
	emp.Get("/attendance", canRead, h.ListAttendanceForEmployee)
	emp.Post("/attendance", canWrite, auditMW, h.UpsertAttendance)

	// Salary history + revision
	emp.Get("/salary-history", canRead, h.ListSalaryHistory)
	emp.Post("/salary-history", canWrite, auditMW, h.AddSalaryRevision)

	// CV upload — multipart form, field name "file".
	emp.Post("/cv", canWrite, auditMW, h.UploadCV)
}
