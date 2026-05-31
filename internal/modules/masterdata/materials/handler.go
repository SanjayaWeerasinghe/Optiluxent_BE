package materials

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

func parseID(c *fiber.Ctx, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Params(param), 10, 64)
	return uint(id), err
}

// ── Core Material ─────────────────────────────────────────────────────────────

func (h *Handler) List(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	materialType := c.Query("material_type")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.Count(c.Context(), tenantID, materialType)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count materials")
	}
	rows, err := h.svc.List(c.Context(), tenantID, materialType, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list materials")
	}
	return httputil.SuccessWithMeta(c, "materials retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	d, err := h.svc.Get(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "material not found")
	}
	return httputil.Success(c, "material retrieved", d)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	d, err := h.svc.Create(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "material created", d)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	d, err := h.svc.Update(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "material updated", d)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Purchasing ────────────────────────────────────────────────────────────────

func (h *Handler) GetPurchasing(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	p, err := h.svc.GetPurchasing(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, err.Error())
	}
	return httputil.Success(c, "purchasing retrieved", p)
}

func (h *Handler) UpsertPurchasing(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpsertPurchasingRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	p, err := h.svc.UpsertPurchasing(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchasing saved", p)
}

// ── Manufacturing ─────────────────────────────────────────────────────────────

func (h *Handler) GetManufacturing(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	m, err := h.svc.GetManufacturing(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, err.Error())
	}
	return httputil.Success(c, "manufacturing retrieved", m)
}

func (h *Handler) UpsertManufacturing(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpsertManufacturingRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	m, err := h.svc.UpsertManufacturing(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "manufacturing saved", m)
}

// ── Warehouse ─────────────────────────────────────────────────────────────────

func (h *Handler) GetWarehouse(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	w, err := h.svc.GetWarehouse(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, err.Error())
	}
	return httputil.Success(c, "warehouse retrieved", w)
}

func (h *Handler) UpsertWarehouse(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpsertWarehouseRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	w, err := h.svc.UpsertWarehouse(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "warehouse saved", w)
}

// ── Vendors ───────────────────────────────────────────────────────────────────

func (h *Handler) ListVendors(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	rows, err := h.svc.ListVendors(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "vendors retrieved", rows)
}

func (h *Handler) AddVendor(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req VendorRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	v, err := h.svc.AddVendor(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "vendor added", v)
}

func (h *Handler) UpdateVendor(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	vid, err := parseID(c, "vid")
	if err != nil {
		return httputil.BadRequest(c, "invalid vendor id")
	}
	var req VendorRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	v, err := h.svc.UpdateVendor(c.Context(), tenantFromCtx(c), id, vid, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "vendor updated", v)
}

func (h *Handler) DeleteVendor(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	vid, err := parseID(c, "vid")
	if err != nil {
		return httputil.BadRequest(c, "invalid vendor id")
	}
	if err := h.svc.DeleteVendor(c.Context(), tenantFromCtx(c), id, vid); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) ReplaceVendors(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var reqs []VendorRequest
	if err := c.BodyParser(&reqs); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	rows, err := h.svc.ReplaceVendors(c.Context(), tenantFromCtx(c), id, reqs)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "vendors replaced", rows)
}

// ── Measurements ──────────────────────────────────────────────────────────────

func (h *Handler) ListMeasurements(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	rows, err := h.svc.ListMeasurements(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "measurements retrieved", rows)
}

func (h *Handler) AddMeasurement(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req MeasurementRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	m, err := h.svc.AddMeasurement(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "measurement added", m)
}

func (h *Handler) UpdateMeasurement(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	mid, err := parseID(c, "mid")
	if err != nil {
		return httputil.BadRequest(c, "invalid measurement id")
	}
	var req MeasurementRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	m, err := h.svc.UpdateMeasurement(c.Context(), tenantFromCtx(c), id, mid, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "measurement updated", m)
}

func (h *Handler) DeleteMeasurement(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	mid, err := parseID(c, "mid")
	if err != nil {
		return httputil.BadRequest(c, "invalid measurement id")
	}
	if err := h.svc.DeleteMeasurement(c.Context(), tenantFromCtx(c), id, mid); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) ReplaceMeasurements(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var reqs []MeasurementRequest
	if err := c.BodyParser(&reqs); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	rows, err := h.svc.ReplaceMeasurements(c.Context(), tenantFromCtx(c), id, reqs)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "measurements replaced", rows)
}
