package manufacturing

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	httputil "erp-system/pkg/http"
)

// Handler handles HTTP requests for the manufacturing module.
type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func tenantFromCtx(c *fiber.Ctx) uint {
	v, _ := c.Locals("tenant_id").(uint)
	return v
}

func userFromCtx(c *fiber.Ctx) uint {
	v, _ := c.Locals("user_id").(uint)
	return v
}

func validateStruct(v *validator.Validate, s interface{}) map[string]interface{} {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	errs := map[string]interface{}{}
	for _, e := range err.(validator.ValidationErrors) {
		errs[e.Field()] = e.Tag()
	}
	return errs
}

func parseID(c *fiber.Ctx, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Params(param), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ── Cost Estimates ────────────────────────────────────────────────────────────

func (h *Handler) ListEstimates(c *fiber.Ctx) error {
	rows, err := h.svc.ListEstimates(c.Context(), tenantFromCtx(c), c.Query("status"))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list cost estimates")
	}
	return httputil.Success(c, "cost estimates retrieved", rows)
}

func (h *Handler) GetEstimate(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	est, err := h.svc.GetEstimate(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "cost estimate not found")
	}
	return httputil.Success(c, "cost estimate retrieved", est)
}

func (h *Handler) CreateEstimate(c *fiber.Ctx) error {
	var req CreateEstimateRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	est, err := h.svc.CreateEstimate(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "cost estimate created", est)
}

func (h *Handler) UpdateEstimate(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateEstimateRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	est, err := h.svc.UpdateEstimate(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "cost estimate updated", est)
}

func (h *Handler) DeleteEstimate(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteEstimate(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) ApproveEstimate(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	est, err := h.svc.ApproveEstimate(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "cost estimate approved", est)
}

func (h *Handler) CancelEstimate(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	est, err := h.svc.CancelEstimate(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "cost estimate cancelled", est)
}

// ── Cost Estimate Lines ───────────────────────────────────────────────────────

func (h *Handler) ListEstimateLines(c *fiber.Ctx) error {
	estimateID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lines, err := h.svc.ListEstimateLines(c.Context(), tenantFromCtx(c), estimateID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "cost estimate lines retrieved", lines)
}

func (h *Handler) AddEstimateLine(c *fiber.Ctx) error {
	estimateID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddEstimateLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddEstimateLine(c.Context(), tenantFromCtx(c), estimateID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "cost estimate line added", line)
}

func (h *Handler) DeleteEstimateLine(c *fiber.Ctx) error {
	estimateID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "lineId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteEstimateLine(c.Context(), tenantFromCtx(c), estimateID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Production Plans ──────────────────────────────────────────────────────────

func (h *Handler) ListPlans(c *fiber.Ctx) error {
	rows, err := h.svc.ListPlans(c.Context(), tenantFromCtx(c), c.Query("status"))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list production plans")
	}
	return httputil.Success(c, "production plans retrieved", rows)
}

func (h *Handler) GetPlan(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	plan, err := h.svc.GetPlan(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "production plan not found")
	}
	return httputil.Success(c, "production plan retrieved", plan)
}

func (h *Handler) CreatePlan(c *fiber.Ctx) error {
	var req CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	plan, err := h.svc.CreatePlan(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "production plan created", plan)
}

func (h *Handler) UpdatePlan(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	plan, err := h.svc.UpdatePlan(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production plan updated", plan)
}

func (h *Handler) DeletePlan(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeletePlan(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) ReleasePlan(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	plan, err := h.svc.ReleasePlan(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production plan released", plan)
}

func (h *Handler) CancelPlan(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	plan, err := h.svc.CancelPlan(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production plan cancelled", plan)
}

func (h *Handler) CreateOrderFromPlan(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	order, err := h.svc.CreateOrderFromPlan(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "production order created from plan", order)
}

// ── Production Plan Inputs ────────────────────────────────────────────────────

func (h *Handler) ListPlanInputs(c *fiber.Ctx) error {
	planID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	inputs, err := h.svc.ListPlanInputs(c.Context(), tenantFromCtx(c), planID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "plan inputs retrieved", inputs)
}

func (h *Handler) AddPlanInput(c *fiber.Ctx) error {
	planID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddPlanInputRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	input, err := h.svc.AddPlanInput(c.Context(), tenantFromCtx(c), planID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "plan input added", input)
}

func (h *Handler) UpdatePlanInput(c *fiber.Ctx) error {
	planID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid plan id")
	}
	inputID, err := parseID(c, "inputId")
	if err != nil {
		return httputil.BadRequest(c, "invalid input id")
	}
	var req UpdatePlanInputRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	input, err := h.svc.UpdatePlanInput(c.Context(), tenantFromCtx(c), planID, inputID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "plan input updated", input)
}

func (h *Handler) DeletePlanInput(c *fiber.Ctx) error {
	planID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid plan id")
	}
	inputID, err := parseID(c, "inputId")
	if err != nil {
		return httputil.BadRequest(c, "invalid input id")
	}
	if err := h.svc.DeletePlanInput(c.Context(), tenantFromCtx(c), planID, inputID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Production Orders ─────────────────────────────────────────────────────────

func (h *Handler) ListOrders(c *fiber.Ctx) error {
	rows, err := h.svc.ListOrders(c.Context(), tenantFromCtx(c), c.Query("status"))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list production orders")
	}
	return httputil.Success(c, "production orders retrieved", rows)
}

func (h *Handler) GetOrder(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	order, err := h.svc.GetOrder(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "production order not found")
	}
	return httputil.Success(c, "production order retrieved", order)
}

// GetOrderDashboard is the aggregation endpoint the FE opens when someone
// clicks a Manufacturing Order — one packet with the order + all linked MRs,
// issues, transfers, GRNs and QCs plus a totals block already normalised to
// the order's UOM.
func (h *Handler) GetOrderDashboard(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	dash, err := h.svc.GetDashboard(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, err.Error())
	}
	return httputil.Success(c, "dashboard retrieved", dash)
}

func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	order, err := h.svc.CreateOrder(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "production order created", order)
}

func (h *Handler) UpdateOrder(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	order, err := h.svc.UpdateOrder(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production order updated", order)
}

func (h *Handler) DeleteOrder(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteOrder(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) StartOrder(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	order, err := h.svc.StartOrder(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production order started", order)
}

func (h *Handler) CompleteOrder(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	order, err := h.svc.CompleteOrder(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production order completed", order)
}

func (h *Handler) CancelOrder(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	order, err := h.svc.CancelOrder(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production order cancelled", order)
}

// ── Production Outputs ────────────────────────────────────────────────────────

func (h *Handler) ListOutputs(c *fiber.Ctx) error {
	orderID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	outputs, err := h.svc.ListOutputs(c.Context(), tenantFromCtx(c), orderID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production outputs retrieved", outputs)
}

func (h *Handler) AddOutput(c *fiber.Ctx) error {
	orderID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddOutputRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	output, err := h.svc.AddOutput(c.Context(), tenantFromCtx(c), orderID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "production output added", output)
}

func (h *Handler) DeleteOutput(c *fiber.Ctx) error {
	orderID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	outputID, err := parseID(c, "outputId")
	if err != nil {
		return httputil.BadRequest(c, "invalid output id")
	}
	if err := h.svc.DeleteOutput(c.Context(), tenantFromCtx(c), orderID, outputID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Production Resources ──────────────────────────────────────────────────────

func (h *Handler) ListResources(c *fiber.Ctx) error {
	orderID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	resources, err := h.svc.ListResources(c.Context(), tenantFromCtx(c), orderID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "production resources retrieved", resources)
}

func (h *Handler) AddResource(c *fiber.Ctx) error {
	orderID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddResourceRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	resource, err := h.svc.AddResource(c.Context(), tenantFromCtx(c), orderID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "production resource added", resource)
}

func (h *Handler) DeleteResource(c *fiber.Ctx) error {
	orderID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	resourceID, err := parseID(c, "resourceId")
	if err != nil {
		return httputil.BadRequest(c, "invalid resource id")
	}
	if err := h.svc.DeleteResource(c.Context(), tenantFromCtx(c), orderID, resourceID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Post Costs ────────────────────────────────────────────────────────────────

func (h *Handler) ListPostCosts(c *fiber.Ctx) error {
	rows, err := h.svc.ListPostCosts(c.Context(), tenantFromCtx(c), c.Query("status"))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list post-cost records")
	}
	return httputil.Success(c, "post-cost records retrieved", rows)
}

func (h *Handler) GetPostCost(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pc, err := h.svc.GetPostCost(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "post-cost record not found")
	}
	return httputil.Success(c, "post-cost record retrieved", pc)
}

func (h *Handler) CreatePostCost(c *fiber.Ctx) error {
	var req CreatePostCostRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	pc, err := h.svc.CreatePostCost(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "post-cost record created", pc)
}

func (h *Handler) FinalizePostCost(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pc, err := h.svc.FinalizePostCost(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "post-cost record finalized", pc)
}
