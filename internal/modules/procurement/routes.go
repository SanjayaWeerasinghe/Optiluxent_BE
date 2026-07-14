package procurement

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW    := middleware.AuditLog(auditLogger)
	canRead    := middleware.RequirePermission(enforcer, "procurement", "read")
	canWrite   := middleware.RequirePermission(enforcer, "procurement", "write")
	canDelete  := middleware.RequirePermission(enforcer, "procurement", "delete")
	canApprove := middleware.RequirePermission(enforcer, "procurement", "approve")

	// ── Purchase Requests ───────────────────────────────────────────────────────
	prs := router.Group("/purchase-requests")
	prs.Get("", canRead, h.ListPRs)
	prs.Post("", canWrite, auditMW, h.CreatePR)
	prs.Get("/:id", canRead, h.GetPR)
	prs.Put("/:id", canWrite, auditMW, h.UpdatePR)
	prs.Delete("/:id", canDelete, auditMW, h.DeletePR)

	// PR workflow
	prs.Post("/:id/submit", canWrite, auditMW, h.SubmitPR)
	prs.Post("/:id/approve", canApprove, auditMW, h.ApprovePR)
	prs.Post("/:id/reject", canApprove, auditMW, h.RejectPR)
	prs.Post("/:id/cancel", canWrite, auditMW, h.CancelPR)

	// PR Items (PRI)
	prs.Get("/:id/items", canRead, h.ListPRItems)
	prs.Post("/:id/items", canWrite, auditMW, h.AddPRItem)
	prs.Get("/:id/items/:itemId", canRead, h.GetPRItem)
	prs.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdatePRItem)
	prs.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeletePRItem)

	// ── Purchase Orders ─────────────────────────────────────────────────────────
	pos := router.Group("/purchase-orders")
	pos.Get("", canRead, h.ListPOs)
	pos.Post("", canWrite, auditMW, h.CreatePO)
	pos.Get("/:id", canRead, h.GetPO)
	pos.Put("/:id", canWrite, auditMW, h.UpdatePO)
	pos.Delete("/:id", canDelete, auditMW, h.DeletePO)

	// PO workflow
	pos.Post("/:id/confirm", canApprove, auditMW, h.ConfirmPO)
	pos.Post("/:id/cancel", canWrite, auditMW, h.CancelPO)

	// PO Items (POI)
	pos.Get("/:id/items", canRead, h.ListPOItems)
	pos.Post("/:id/items", canWrite, auditMW, h.AddPOItem)
	pos.Get("/:id/items/:itemId", canRead, h.GetPOItem)
	pos.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdatePOItem)
	pos.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeletePOItem)

	// ── Goods Receipts ──────────────────────────────────────────────────────────
	grns := router.Group("/goods-receipts")
	grns.Get("", canRead, h.ListGRNs)
	grns.Post("", canWrite, auditMW, h.CreateGRN)
	grns.Get("/:id", canRead, h.GetGRN)

	// GRN workflow
	grns.Post("/:id/confirm", canApprove, auditMW, h.ConfirmGRN)
	grns.Post("/:id/cancel", canWrite, auditMW, h.CancelGRN)

	// GRN Items (GRI)
	grns.Get("/:id/items", canRead, h.ListGRNItems)
	grns.Post("/:id/items", canWrite, auditMW, h.AddGRNItem)
	grns.Get("/:id/items/:itemId", canRead, h.GetGRNItem)
	grns.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateGRNItem)

	// ── Purchase Invoices ───────────────────────────────────────────────────────
	invs := router.Group("/purchase-invoices")
	invs.Get("", canRead, h.ListInvoices)
	invs.Post("", canWrite, auditMW, h.CreateInvoice)
	invs.Get("/:id", canRead, h.GetInvoice)

	// Invoice lines
	invs.Post("/:id/lines", canWrite, auditMW, h.AddInvoiceLine)
	invs.Delete("/:id/lines/:lineId", canDelete, auditMW, h.DeleteInvoiceLine)

	// Invoice workflow
	invs.Post("/:id/post", canApprove, auditMW, h.PostInvoice)
	invs.Post("/:id/pay", canWrite, auditMW, h.RecordPayment)
	invs.Post("/:id/cancel", canWrite, auditMW, h.CancelInvoice)
}
