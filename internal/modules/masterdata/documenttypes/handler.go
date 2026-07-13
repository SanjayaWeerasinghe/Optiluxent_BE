package documenttypes

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

func parseID(c *fiber.Ctx, key string) (uint, error) {
	id, err := strconv.ParseUint(c.Params(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ── Types ────────────────────────────────────────────────────────────────────

func (h *Handler) ListTypes(c *fiber.Ctx) error {
	model := c.Query("model")
	rows, err := h.svc.ListTypes(c.Context(), tenantFromCtx(c), model)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list document types")
	}
	return httputil.Success(c, "document types retrieved", rows)
}

func (h *Handler) GetType(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	t, err := h.svc.GetType(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "document type not found")
	}
	return httputil.Success(c, "document type retrieved", t)
}

func (h *Handler) CreateType(c *fiber.Ctx) error {
	var req CreateTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	t, err := h.svc.CreateType(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Created(c, "document type created", t)
}

func (h *Handler) UpdateType(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	t, err := h.svc.UpdateType(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "document type updated", t)
}

func (h *Handler) DeleteType(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteType(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Fields ───────────────────────────────────────────────────────────────────

func (h *Handler) ListFields(c *fiber.Ctx) error {
	typeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	rows, err := h.svc.ListFields(c.Context(), tenantFromCtx(c), typeID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list fields")
	}
	return httputil.Success(c, "fields retrieved", rows)
}

func (h *Handler) CreateField(c *fiber.Ctx) error {
	typeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req CreateFieldRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	f, err := h.svc.CreateField(c.Context(), tenantFromCtx(c), typeID, &req)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Created(c, "field created", f)
}

func (h *Handler) UpdateField(c *fiber.Ctx) error {
	fieldID, err := parseID(c, "fieldId")
	if err != nil {
		return httputil.BadRequest(c, "invalid field id")
	}
	var req UpdateFieldRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	f, err := h.svc.UpdateField(c.Context(), tenantFromCtx(c), fieldID, &req)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "field updated", f)
}

func (h *Handler) DeleteField(c *fiber.Ctx) error {
	fieldID, err := parseID(c, "fieldId")
	if err != nil {
		return httputil.BadRequest(c, "invalid field id")
	}
	if err := h.svc.DeleteField(c.Context(), tenantFromCtx(c), fieldID); err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Doc-side values ──────────────────────────────────────────────────────────

// GET /documents/:kind/:id/fields?type_id=N
func (h *Handler) GetDocValues(c *fiber.Ctx) error {
	docKind := c.Params("kind")
	docID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid doc id")
	}
	typeID, _ := strconv.ParseUint(c.Query("type_id"), 10, 64)
	if typeID == 0 {
		return httputil.Success(c, "no type — no fields", []ValueResponse{})
	}
	rows, err := h.svc.GetValuesForDoc(c.Context(), tenantFromCtx(c), docKind, docID, uint(typeID))
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "values retrieved", rows)
}

// POST /documents/:kind/:id/fields — batch upsert
func (h *Handler) UpsertDocValues(c *fiber.Ctx) error {
	docKind := c.Params("kind")
	docID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid doc id")
	}
	var req UpsertValuesRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if err := h.svc.UpsertValuesForDoc(c.Context(), tenantFromCtx(c), docKind, docID, req.Values); err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "values saved", nil)
}
