package hr

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

// ── Job Positions ─────────────────────────────────────────────────────────────

func (h *Handler) ListJobPositions(c *fiber.Ctx) error {
	rows, err := h.svc.ListJobPositions(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list job positions")
	}
	return httputil.Success(c, "job positions retrieved", rows)
}

func (h *Handler) CreateJobPosition(c *fiber.Ctx) error {
	var req CreateJobPositionRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	jp, err := h.svc.CreateJobPosition(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "job position created", jp)
}

func (h *Handler) UpdateJobPosition(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateJobPositionRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	jp, err := h.svc.UpdateJobPosition(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "job position updated", jp)
}

// ── Employees ─────────────────────────────────────────────────────────────────

func (h *Handler) ListEmployees(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	activeOnly := c.Query("active_only") == "true"
	var deptID *uint
	if deptStr := c.Query("department_id"); deptStr != "" {
		if d, err := strconv.ParseUint(deptStr, 10, 64); err == nil {
			v := uint(d)
			deptID = &v
		}
	}
	rows, err := h.svc.ListEmployees(c.Context(), tenantID, activeOnly, deptID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list employees")
	}
	return httputil.Success(c, "employees retrieved", rows)
}

func (h *Handler) GetEmployee(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	e, err := h.svc.GetEmployee(c.Context(), tenantFromCtx(c), uint(id))
	if err != nil {
		return httputil.NotFound(c, "employee not found")
	}
	return httputil.Success(c, "employee retrieved", e)
}

func (h *Handler) CreateEmployee(c *fiber.Ctx) error {
	var req CreateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	e, err := h.svc.CreateEmployee(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "employee created", e)
}

func (h *Handler) UpdateEmployee(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	e, err := h.svc.UpdateEmployee(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "employee updated", e)
}

func (h *Handler) DeleteEmployee(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteEmployee(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}
