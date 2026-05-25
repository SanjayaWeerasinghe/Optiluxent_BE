package manufacturing

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	httputil "erp-system/pkg/http"
)

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

// ── BOMs ──────────────────────────────────────────────────────────────────────

func (h *Handler) ListBOMs(c *fiber.Ctx) error {
	rows, err := h.svc.ListBOMs(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list BOMs")
	}
	return httputil.Success(c, "BOMs retrieved", rows)
}

func (h *Handler) GetBOM(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	b, err := h.svc.GetBOM(c.Context(), tenantFromCtx(c), uint(id))
	if err != nil {
		return httputil.NotFound(c, "BOM not found")
	}
	return httputil.Success(c, "BOM retrieved", b)
}

func (h *Handler) CreateBOM(c *fiber.Ctx) error {
	var req CreateBOMRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	b, err := h.svc.CreateBOM(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "BOM created", b)
}

func (h *Handler) UpdateBOM(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateBOMRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	b, err := h.svc.UpdateBOM(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "BOM updated", b)
}

func (h *Handler) DeleteBOM(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteBOM(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) AddBOMLine(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	bomID, err := strconv.ParseUint(c.Params("bomId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid bom id")
	}
	var req AddBOMLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddBOMLine(c.Context(), tenantID, uint(bomID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "BOM line added", line)
}

func (h *Handler) DeleteBOMLine(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteBOMLine(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Work Centers ──────────────────────────────────────────────────────────────

func (h *Handler) ListWorkCenters(c *fiber.Ctx) error {
	rows, err := h.svc.ListWorkCenters(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list work centers")
	}
	return httputil.Success(c, "work centers retrieved", rows)
}

func (h *Handler) CreateWorkCenter(c *fiber.Ctx) error {
	var req CreateWorkCenterRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	wc, err := h.svc.CreateWorkCenter(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "work center created", wc)
}

func (h *Handler) UpdateWorkCenter(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateWorkCenterRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	wc, err := h.svc.UpdateWorkCenter(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "work center updated", wc)
}

func (h *Handler) DeleteWorkCenter(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteWorkCenter(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Routings ──────────────────────────────────────────────────────────────────

func (h *Handler) ListRoutings(c *fiber.Ctx) error {
	rows, err := h.svc.ListRoutings(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list routings")
	}
	return httputil.Success(c, "routings retrieved", rows)
}

func (h *Handler) GetRouting(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	rt, err := h.svc.GetRouting(c.Context(), tenantFromCtx(c), uint(id))
	if err != nil {
		return httputil.NotFound(c, "routing not found")
	}
	return httputil.Success(c, "routing retrieved", rt)
}

func (h *Handler) CreateRouting(c *fiber.Ctx) error {
	var req CreateRoutingRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	rt, err := h.svc.CreateRouting(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "routing created", rt)
}

func (h *Handler) DeleteRouting(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteRouting(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) AddRoutingOperation(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	routingID, err := strconv.ParseUint(c.Params("routingId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid routing id")
	}
	var req AddOperationRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	op, err := h.svc.AddRoutingOperation(c.Context(), tenantID, uint(routingID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "operation added", op)
}

func (h *Handler) DeleteRoutingOperation(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteRoutingOperation(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}
