package mmcategories

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

func (h *Handler) List(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.Count(c.Context(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count categories")
	}
	rows, err := h.svc.List(c.Context(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list categories")
	}
	return httputil.SuccessWithMeta(c, "categories retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	cat, err := h.svc.Create(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "category created", cat)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	cat, err := h.svc.Update(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "category updated", cat)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) Config(c *fiber.Ctx) error {
	return httputil.Success(c, "config retrieved", fiber.Map{
		"max_depth": h.svc.MaxDepth(),
	})
}
