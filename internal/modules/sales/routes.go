package sales

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW    := middleware.AuditLog(auditLogger)
	canRead    := middleware.RequirePermission(enforcer, "sales", "read")
	canWrite   := middleware.RequirePermission(enforcer, "sales", "write")
	canDelete  := middleware.RequirePermission(enforcer, "sales", "delete")
	canApprove := middleware.RequirePermission(enforcer, "sales", "approve")

	// ── Sales Quotations ────────────────────────────────────────────────────────
	sqs := router.Group("/quotations")
	sqs.Get("", canRead, h.ListSQs)
	sqs.Post("", canWrite, auditMW, h.CreateSQ)
	sqs.Get("/:id", canRead, h.GetSQ)
	sqs.Put("/:id", canWrite, auditMW, h.UpdateSQ)
	sqs.Delete("/:id", canDelete, auditMW, h.DeleteSQ)

	// SQ workflow
	sqs.Post("/:id/submit", canWrite, auditMW, h.SubmitSQ)
	sqs.Post("/:id/accept", canApprove, auditMW, h.AcceptSQ)
	sqs.Post("/:id/reject", canApprove, auditMW, h.RejectSQ)
	sqs.Post("/:id/cancel", canWrite, auditMW, h.CancelSQ)

	// SQ Lines
	sqs.Get("/:id/items", canRead, h.ListSQLines)
	sqs.Post("/:id/items", canWrite, auditMW, h.AddSQLine)
	sqs.Get("/:id/items/:itemId", canRead, h.GetSQLine)
	sqs.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateSQLine)
	sqs.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteSQLine)

	// ── Sales Orders ────────────────────────────────────────────────────────────
	sos := router.Group("/sales-orders")
	sos.Get("", canRead, h.ListSOs)
	sos.Post("", canWrite, auditMW, h.CreateSO)
	sos.Get("/:id", canRead, h.GetSO)
	sos.Put("/:id", canWrite, auditMW, h.UpdateSO)
	sos.Delete("/:id", canDelete, auditMW, h.DeleteSO)

	// SO workflow
	sos.Post("/:id/confirm", canApprove, auditMW, h.ConfirmSO)
	sos.Post("/:id/cancel", canWrite, auditMW, h.CancelSO)

	// SO Lines
	sos.Get("/:id/items", canRead, h.ListSOLines)
	sos.Post("/:id/items", canWrite, auditMW, h.AddSOLine)
	sos.Get("/:id/items/:itemId", canRead, h.GetSOLine)
	sos.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateSOLine)
	sos.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteSOLine)

	// ── Delivery Orders ─────────────────────────────────────────────────────────
	dos := router.Group("/deliveries")
	dos.Get("", canRead, h.ListDOs)
	dos.Post("", canWrite, auditMW, h.CreateDO)
	dos.Get("/:id", canRead, h.GetDO)
	dos.Put("/:id", canWrite, auditMW, h.UpdateDO)

	// DO workflow
	dos.Post("/:id/confirm", canApprove, auditMW, h.ConfirmDO)
	dos.Post("/:id/cancel", canWrite, auditMW, h.CancelDO)

	// DO Lines
	dos.Get("/:id/items", canRead, h.ListDOLines)
	dos.Post("/:id/items", canWrite, auditMW, h.AddDOLine)
	dos.Get("/:id/items/:itemId", canRead, h.GetDOLine)
	dos.Put("/:id/items/:itemId", canWrite, auditMW, h.UpdateDOLine)
	dos.Delete("/:id/items/:itemId", canDelete, auditMW, h.DeleteDOLine)

	// ── Sales Invoices ──────────────────────────────────────────────────────────
	invs := router.Group("/invoices")
	invs.Get("", canRead, h.ListSIs)
	invs.Post("", canWrite, auditMW, h.CreateSI)
	invs.Get("/:id", canRead, h.GetSI)

	// Invoice lines
	invs.Get("/:id/lines", canRead, h.ListSILines)
	invs.Post("/:id/lines", canWrite, auditMW, h.AddSILine)
	invs.Delete("/:id/lines/:lineId", canDelete, auditMW, h.DeleteSILine)

	// Invoice workflow
	invs.Post("/:id/post", canApprove, auditMW, h.PostSI)
	invs.Post("/:id/pay", canWrite, auditMW, h.RecordPayment)
	invs.Post("/:id/cancel", canWrite, auditMW, h.CancelSI)
}
