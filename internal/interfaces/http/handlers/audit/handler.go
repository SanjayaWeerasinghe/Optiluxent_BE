package audit

import (
	"strconv"

	domainaudit "erp-system/internal/domain/audit"
	httputil "erp-system/pkg/http"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	repo domainaudit.Repository
}

func NewHandler(repo domainaudit.Repository) *Handler {
	return &Handler{repo: repo}
}

// GET /api/v1/audit-logs
func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit > 200 {
		limit = 200
	}

	filter := domainaudit.ListFilter{
		Resource:   c.Query("resource"),
		ResourceID: c.Query("resource_id"),
		Action:     c.Query("action"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		Limit:      limit,
		Offset:     offset,
	}

	if tidStr := c.Query("tenant_id"); tidStr != "" {
		if tid, err := strconv.ParseUint(tidStr, 10, 64); err == nil {
			tidUint := uint(tid)
			filter.TenantID = &tidUint
		}
	}
	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			uidUint := uint(uid)
			filter.UserID = &uidUint
		}
	}

	// Non-super-admin can only see their own tenant's logs.
	role, _ := c.Locals("user_role").(string)
	if role != "super_admin" {
		if tid, ok := c.Locals("tenant_id").(uint); ok {
			filter.TenantID = &tid
		}
	}

	logs, total, err := h.repo.List(c.UserContext(), filter)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list audit logs")
	}

	return httputil.Success(c, "Audit logs retrieved", fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
