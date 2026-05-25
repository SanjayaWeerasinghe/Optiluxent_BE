package featureflags

import (
	"strings"

	ff "erp-system/internal/infrastructure/featureflags"
	httputil "erp-system/pkg/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	store    ff.Store
	validate *validator.Validate
}

func NewHandler(store ff.Store) *Handler {
	return &Handler{
		store:    store,
		validate: validator.New(),
	}
}

// GET /api/v1/feature-flags
func (h *Handler) List(c *fiber.Ctx) error {
	tenantID := tenantIDFromCtx(c)
	flags, err := h.store.List(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list feature flags")
	}
	resp := make([]FlagResponse, len(flags))
	for i, f := range flags {
		resp[i] = toFlagResponse(f)
	}
	return httputil.Success(c, "Feature flags retrieved", resp)
}

// GET /api/v1/feature-flags/:name
func (h *Handler) Get(c *fiber.Ctx) error {
	name := c.Params("name")
	tenantID := tenantIDFromCtx(c)

	flag, err := h.store.Get(c.UserContext(), tenantID, name)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get feature flag")
	}
	if flag == nil {
		return httputil.NotFound(c, "Feature flag not found")
	}
	return httputil.Success(c, "Feature flag retrieved", toFlagResponse(flag))
}

// POST /api/v1/feature-flags
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	tenantID := tenantIDFromCtx(c)
	flag := &ff.FeatureFlag{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Rules:       req.Rules,
	}
	if len(flag.Rules) == 0 {
		flag.Rules = []byte("{}")
	}

	if err := h.store.Create(c.UserContext(), flag); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return httputil.BadRequest(c, "Feature flag with this name already exists")
		}
		return httputil.InternalServerError(c, "Failed to create feature flag")
	}
	return httputil.Created(c, "Feature flag created", toFlagResponse(flag))
}

// PUT /api/v1/feature-flags/:name/enabled
func (h *Handler) SetEnabled(c *fiber.Ctx) error {
	name := c.Params("name")
	tenantID := tenantIDFromCtx(c)

	var req SetEnabledRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	existing, err := h.store.Get(c.UserContext(), tenantID, name)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get feature flag")
	}
	if existing == nil {
		return httputil.NotFound(c, "Feature flag not found")
	}

	if err := h.store.Set(c.UserContext(), tenantID, name, req.Enabled); err != nil {
		return httputil.InternalServerError(c, "Failed to update feature flag")
	}
	return httputil.Success(c, "Feature flag updated", nil)
}

// DELETE /api/v1/feature-flags/:name
func (h *Handler) Delete(c *fiber.Ctx) error {
	name := c.Params("name")
	tenantID := tenantIDFromCtx(c)

	existing, err := h.store.Get(c.UserContext(), tenantID, name)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get feature flag")
	}
	if existing == nil {
		return httputil.NotFound(c, "Feature flag not found")
	}

	if err := h.store.Delete(c.UserContext(), tenantID, name); err != nil {
		return httputil.InternalServerError(c, "Failed to delete feature flag")
	}
	return httputil.Success(c, "Feature flag deleted", nil)
}

// POST /api/v1/feature-flags/:name/invalidate
func (h *Handler) Invalidate(c *fiber.Ctx) error {
	name := c.Params("name")
	tenantID := tenantIDFromCtx(c)

	if err := h.store.Invalidate(c.UserContext(), tenantID, name); err != nil {
		return httputil.InternalServerError(c, "Failed to invalidate cache")
	}
	return httputil.Success(c, "Cache invalidated", nil)
}

func tenantIDFromCtx(c *fiber.Ctx) *uint {
	if id, ok := c.Locals("tenant_id").(uint); ok && id != 0 {
		return &id
	}
	return nil
}

func toFlagResponse(f *ff.FeatureFlag) FlagResponse {
	return FlagResponse{
		ID:          f.ID,
		TenantID:    f.TenantID,
		Name:        f.Name,
		Description: f.Description,
		Enabled:     f.Enabled,
		Rules:       f.Rules,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

func validateStruct(v *validator.Validate, s interface{}) map[string]interface{} {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	fields := make(map[string]interface{})
	for _, e := range err.(validator.ValidationErrors) {
		fields[strings.ToLower(e.Field())] = e.Tag()
	}
	return fields
}
