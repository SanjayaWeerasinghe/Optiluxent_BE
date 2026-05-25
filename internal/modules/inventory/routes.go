package inventory

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW    := middleware.AuditLog(auditLogger)
	canRead    := middleware.RequirePermission(enforcer, "inventory", "read")
	canWrite   := middleware.RequirePermission(enforcer, "inventory", "write")
	canDelete  := middleware.RequirePermission(enforcer, "inventory", "delete")
	canApprove := middleware.RequirePermission(enforcer, "inventory", "approve")

	// ── Material Requests ───────────────────────────────────────────────────────
	mrs := router.Group("/material-requests")
	mrs.Get("", canRead, h.ListMRs)
	mrs.Post("", canWrite, auditMW, h.CreateMR)
	mrs.Get("/:id", canRead, h.GetMR)
	mrs.Put("/:id", canWrite, auditMW, h.UpdateMR)
	mrs.Post("/:id/approve", canApprove, auditMW, h.ApproveMR)
	mrs.Post("/:id/reject", canApprove, auditMW, h.RejectMR)

	// MR Items
	mrs.Get("/:id/items", canRead, h.ListMRLines)
	mrs.Post("/:id/items", canWrite, auditMW, h.AddMRLine)
	mrs.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateMRLine)
	mrs.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteMRLine)

	// ── Goods Transfers ─────────────────────────────────────────────────────────
	transfers := router.Group("/transfers")
	transfers.Get("", canRead, h.ListTransfers)
	transfers.Post("", canWrite, auditMW, h.CreateTransfer)
	transfers.Get("/:id", canRead, h.GetTransfer)
	transfers.Put("/:id", canWrite, auditMW, h.UpdateTransfer)
	transfers.Post("/:id/send", canApprove, auditMW, h.SendTransfer)
	transfers.Post("/:id/receive", canApprove, auditMW, h.ReceiveTransfer)

	// Transfer Items
	transfers.Get("/:id/items", canRead, h.ListTransferLines)
	transfers.Post("/:id/items", canWrite, auditMW, h.AddTransferLine)
	transfers.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateTransferLine)
	transfers.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteTransferLine)

	// ── Goods Issues ────────────────────────────────────────────────────────────
	issues := router.Group("/issues")
	issues.Get("", canRead, h.ListIssues)
	issues.Post("", canWrite, auditMW, h.CreateIssue)
	issues.Get("/:id", canRead, h.GetIssue)
	issues.Put("/:id", canWrite, auditMW, h.UpdateIssue)
	issues.Post("/:id/confirm", canApprove, auditMW, h.ConfirmIssue)

	// Issue Items
	issues.Get("/:id/items", canRead, h.ListIssueLines)
	issues.Post("/:id/items", canWrite, auditMW, h.AddIssueLine)
	issues.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateIssueLine)
	issues.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteIssueLine)

	// ── Stock Adjustments ───────────────────────────────────────────────────────
	adjustments := router.Group("/adjustments")
	adjustments.Get("", canRead, h.ListAdjustments)
	adjustments.Post("", canWrite, auditMW, h.CreateAdjustment)
	adjustments.Get("/:id", canRead, h.GetAdjustment)
	adjustments.Put("/:id", canWrite, auditMW, h.UpdateAdjustment)
	adjustments.Post("/:id/confirm", canApprove, auditMW, h.ConfirmAdjustment)

	// Adjustment Items
	adjustments.Get("/:id/items", canRead, h.ListAdjustmentLines)
	adjustments.Post("/:id/items", canWrite, auditMW, h.AddAdjustmentLine)
	adjustments.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateAdjustmentLine)
	adjustments.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteAdjustmentLine)

	// ── Quality Checks ──────────────────────────────────────────────────────────
	qcs := router.Group("/quality-checks")
	qcs.Get("", canRead, h.ListQualityChecks)
	qcs.Post("", canWrite, auditMW, h.CreateQualityCheck)
	qcs.Get("/:id", canRead, h.GetQualityCheck)
	qcs.Post("/:id/start", canApprove, auditMW, h.StartQualityCheck)
	qcs.Post("/:id/submit", canApprove, auditMW, h.SubmitQualityCheck)

	// QC Items
	qcs.Get("/:id/items", canRead, h.ListQCLines)
	qcs.Post("/:id/items", canWrite, auditMW, h.AddQCLine)
	qcs.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateQCLine)

	// ── Stock Balances ──────────────────────────────────────────────────────────
	router.Get("/stock", canRead, h.GetStockBalance)
}
