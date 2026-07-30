package manufacturing

import (
	"github.com/gofiber/fiber/v2"

	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/domain/rbac"
	"erp-system/internal/infrastructure/http/middleware"
)

func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW    := middleware.AuditLog(auditLogger)
	canRead    := middleware.RequirePermission(enforcer, "manufacturing", "read")
	canWrite   := middleware.RequirePermission(enforcer, "manufacturing", "write")
	canDelete  := middleware.RequirePermission(enforcer, "manufacturing", "delete")
	canApprove := middleware.RequirePermission(enforcer, "manufacturing", "approve")

	// ── Cost Estimates (Pre-Costing) ────────────────────────────────────────────
	estimates := router.Group("/pre-costs")
	estimates.Get("", canRead, h.ListEstimates)
	estimates.Post("", canWrite, auditMW, h.CreateEstimate)
	estimates.Get("/:id", canRead, h.GetEstimate)
	estimates.Put("/:id", canWrite, auditMW, h.UpdateEstimate)
	estimates.Delete("/:id", canDelete, auditMW, h.DeleteEstimate)

	// Estimate workflow
	estimates.Post("/:id/approve", canApprove, auditMW, h.ApproveEstimate)
	estimates.Post("/:id/cancel", canWrite, auditMW, h.CancelEstimate)

	// Estimate lines
	estimates.Get("/:id/lines", canRead, h.ListEstimateLines)
	estimates.Post("/:id/lines", canWrite, auditMW, h.AddEstimateLine)
	estimates.Delete("/:id/lines/:lineId", canDelete, auditMW, h.DeleteEstimateLine)

	// ── Production Plans ────────────────────────────────────────────────────────
	plans := router.Group("/plans")
	plans.Get("", canRead, h.ListPlans)
	plans.Post("", canWrite, auditMW, h.CreatePlan)
	plans.Get("/:id", canRead, h.GetPlan)
	plans.Put("/:id", canWrite, auditMW, h.UpdatePlan)
	plans.Delete("/:id", canDelete, auditMW, h.DeletePlan)

	// Plan workflow
	plans.Post("/:id/release", canApprove, auditMW, h.ReleasePlan)
	plans.Post("/:id/cancel", canWrite, auditMW, h.CancelPlan)
	// Spawn a draft Production from a released Plan (copies product/qty/inputs).
	plans.Post("/:id/create-production", canWrite, auditMW, h.CreateOrderFromPlan)

	// Plan Inputs — anticipated chemicals/resources; copied into MO Resources
	// when a Production is manually created from a released Plan.
	plans.Get("/:id/inputs", canRead, h.ListPlanInputs)
	plans.Post("/:id/inputs", canWrite, auditMW, h.AddPlanInput)
	plans.Put("/:id/inputs/:inputId", canWrite, auditMW, h.UpdatePlanInput)
	plans.Delete("/:id/inputs/:inputId", canDelete, auditMW, h.DeletePlanInput)

	// ── Production Orders ───────────────────────────────────────────────────────
	orders := router.Group("/orders")
	orders.Get("", canRead, h.ListOrders)
	orders.Post("", canWrite, auditMW, h.CreateOrder)
	orders.Get("/:id", canRead, h.GetOrder)
	orders.Get("/:id/dashboard", canRead, h.GetOrderDashboard)
	orders.Put("/:id", canWrite, auditMW, h.UpdateOrder)
	orders.Delete("/:id", canDelete, auditMW, h.DeleteOrder)

	// Order workflow
	orders.Post("/:id/start", canWrite, auditMW, h.StartOrder)
	orders.Post("/:id/complete", canApprove, auditMW, h.CompleteOrder)
	orders.Post("/:id/cancel", canWrite, auditMW, h.CancelOrder)

	// Order outputs
	orders.Get("/:id/outputs", canRead, h.ListOutputs)
	orders.Post("/:id/outputs", canWrite, auditMW, h.AddOutput)
	orders.Delete("/:id/outputs/:outputId", canDelete, auditMW, h.DeleteOutput)

	// Order resources
	orders.Get("/:id/resources", canRead, h.ListResources)
	orders.Post("/:id/resources", canWrite, auditMW, h.AddResource)
	orders.Delete("/:id/resources/:resourceId", canDelete, auditMW, h.DeleteResource)

	// ── Post Costs ──────────────────────────────────────────────────────────────
	postCosts := router.Group("/post-costs")
	postCosts.Get("", canRead, h.ListPostCosts)
	postCosts.Post("", canWrite, auditMW, h.CreatePostCost)
	postCosts.Get("/:id", canRead, h.GetPostCost)

	// Post-cost workflow
	postCosts.Post("/:id/finalize", canApprove, auditMW, h.FinalizePostCost)
}
