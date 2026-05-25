package hr

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

	// Job Positions
	router.Get("/job-positions", canRead, h.ListJobPositions)
	router.Post("/job-positions", canWrite, auditMW, h.CreateJobPosition)
	router.Put("/job-positions/:id", canWrite, auditMW, h.UpdateJobPosition)

	// Employees
	router.Get("/employees", canRead, h.ListEmployees)
	router.Post("/employees", canWrite, auditMW, h.CreateEmployee)
	router.Get("/employees/:id", canRead, h.GetEmployee)
	router.Put("/employees/:id", canWrite, auditMW, h.UpdateEmployee)
	router.Delete("/employees/:id", canDelete, auditMW, h.DeleteEmployee)
}
