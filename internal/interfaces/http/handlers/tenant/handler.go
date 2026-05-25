package tenant

import (
	"strconv"
	"strings"

	domaintenant "erp-system/internal/domain/tenant"
	httputil "erp-system/pkg/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	tenantRepo domaintenant.Repository
	validate   *validator.Validate
}

func NewHandler(tenantRepo domaintenant.Repository) *Handler {
	return &Handler{
		tenantRepo: tenantRepo,
		validate:   validator.New(),
	}
}

// POST /api/v1/tenants
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	existing, _ := h.tenantRepo.GetBySlug(c.UserContext(), req.Slug)
	if existing != nil {
		return httputil.Conflict(c, "A tenant with this slug already exists")
	}

	plan := req.Plan
	if plan == "" {
		plan = domaintenant.PlanStandard
	}

	tenant := &domaintenant.Tenant{
		Name:   req.Name,
		Slug:   req.Slug,
		Plan:   plan,
		Status: domaintenant.StatusActive,
		Config: req.Config,
	}
	if len(tenant.Config) == 0 {
		tenant.Config = []byte("{}")
	}

	if err := h.tenantRepo.Create(c.UserContext(), tenant); err != nil {
		return httputil.InternalServerError(c, "Failed to create tenant")
	}

	return httputil.Created(c, "Tenant created", toTenantResponse(tenant))
}

// GET /api/v1/tenants
func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit > 100 {
		limit = 100
	}

	tenants, total, err := h.tenantRepo.List(c.UserContext(), limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list tenants")
	}

	resp := make([]TenantResponse, len(tenants))
	for i, t := range tenants {
		resp[i] = toTenantResponse(t)
	}

	return httputil.Success(c, "Tenants retrieved", ListTenantsResponse{
		Tenants: resp,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// GET /api/v1/tenants/:id
func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid tenant ID")
	}

	tenant, err := h.tenantRepo.GetByID(c.UserContext(), uint(id))
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get tenant")
	}
	if tenant == nil {
		return httputil.NotFound(c, "Tenant not found")
	}

	return httputil.Success(c, "Tenant retrieved", toTenantResponse(tenant))
}

// PUT /api/v1/tenants/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid tenant ID")
	}

	var req UpdateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	tenant, err := h.tenantRepo.GetByID(c.UserContext(), uint(id))
	if err != nil || tenant == nil {
		return httputil.NotFound(c, "Tenant not found")
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Plan != "" {
		tenant.Plan = req.Plan
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}
	if len(req.Config) > 0 {
		tenant.Config = req.Config
	}

	if err := h.tenantRepo.Update(c.UserContext(), tenant); err != nil {
		return httputil.InternalServerError(c, "Failed to update tenant")
	}

	return httputil.Success(c, "Tenant updated", toTenantResponse(tenant))
}

func toTenantResponse(t *domaintenant.Tenant) TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		Plan:      t.Plan,
		Status:    t.Status,
		Config:    t.Config,
		CreatedAt: t.CreatedAt,
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
