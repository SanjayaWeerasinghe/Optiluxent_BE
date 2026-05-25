package products

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

// ── Categories ────────────────────────────────────────────────────────────────

func (h *Handler) ListCategories(c *fiber.Ctx) error {
	rows, err := h.svc.ListCategories(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list categories")
	}
	return httputil.Success(c, "categories retrieved", rows)
}

func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	cat, err := h.svc.CreateCategory(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "category created", cat)
}

func (h *Handler) UpdateCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	cat, err := h.svc.UpdateCategory(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "category updated", cat)
}

func (h *Handler) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteCategory(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── UOMs ──────────────────────────────────────────────────────────────────────

func (h *Handler) ListUOMs(c *fiber.Ctx) error {
	rows, err := h.svc.ListUOMs(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list UOMs")
	}
	return httputil.Success(c, "UOMs retrieved", rows)
}

func (h *Handler) CreateUOM(c *fiber.Ctx) error {
	var req CreateUOMRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	u, err := h.svc.CreateUOM(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "UOM created", u)
}

func (h *Handler) UpdateUOM(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateUOMRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	u, err := h.svc.UpdateUOM(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "UOM updated", u)
}

// ── Products ──────────────────────────────────────────────────────────────────

func (h *Handler) ListProducts(c *fiber.Ctx) error {
	productType := c.Query("product_type")
	activeOnly := c.Query("active_only") == "true"
	rows, err := h.svc.ListProducts(c.Context(), tenantFromCtx(c), productType, activeOnly)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "products retrieved", rows)
}

func (h *Handler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	p, err := h.svc.GetProduct(c.Context(), tenantFromCtx(c), uint(id))
	if err != nil {
		return httputil.NotFound(c, "product not found")
	}
	return httputil.Success(c, "product retrieved", p)
}

func (h *Handler) CreateProduct(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	p, err := h.svc.CreateProduct(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "product created", p)
}

func (h *Handler) UpdateProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	p, err := h.svc.UpdateProduct(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "product updated", p)
}

func (h *Handler) DeleteProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteProduct(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Variants ──────────────────────────────────────────────────────────────────

func (h *Handler) ListVariants(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("productId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid product id")
	}
	rows, err := h.svc.ListVariants(c.Context(), tenantFromCtx(c), uint(productID))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list variants")
	}
	return httputil.Success(c, "variants retrieved", rows)
}

func (h *Handler) CreateVariant(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	productID, err := strconv.ParseUint(c.Params("productId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid product id")
	}
	var req CreateVariantRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	v, err := h.svc.CreateVariant(c.Context(), tenantID, uint(productID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "variant created", v)
}

// ── Prices ────────────────────────────────────────────────────────────────────

func (h *Handler) ListPrices(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("productId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid product id")
	}
	rows, err := h.svc.ListPrices(c.Context(), tenantFromCtx(c), uint(productID))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list prices")
	}
	return httputil.Success(c, "prices retrieved", rows)
}

func (h *Handler) CreatePrice(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	productID, err := strconv.ParseUint(c.Params("productId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid product id")
	}
	var req CreatePriceRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	p, err := h.svc.CreatePrice(c.Context(), tenantID, uint(productID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "price created", p)
}

func (h *Handler) DeletePrice(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeletePrice(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}
