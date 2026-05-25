package contacts

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

// ── Parties ──────────────────────────────────────────────────────────────────

func (h *Handler) ListParties(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	partyType := c.Query("party_type")
	activeOnly := c.Query("active_only") == "true"

	rows, err := h.svc.ListParties(c.Context(), tenantID, partyType, activeOnly)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "parties retrieved", rows)
}

func (h *Handler) GetParty(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	p, err := h.svc.GetParty(c.Context(), tenantID, uint(id))
	if err != nil {
		return httputil.NotFound(c, "party not found")
	}
	return httputil.Success(c, "party retrieved", p)
}

func (h *Handler) CreateParty(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var req CreatePartyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	p, err := h.svc.CreateParty(c.Context(), tenantID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "party created", p)
}

func (h *Handler) UpdateParty(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdatePartyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	p, err := h.svc.UpdateParty(c.Context(), tenantID, uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "party updated", p)
}

func (h *Handler) DeleteParty(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteParty(c.Context(), tenantID, uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Contact Persons ───────────────────────────────────────────────────────────

func (h *Handler) ListContactPersons(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	partyID, err := strconv.ParseUint(c.Params("partyId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid party id")
	}
	rows, err := h.svc.ListContactPersons(c.Context(), tenantID, uint(partyID))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list contact persons")
	}
	return httputil.Success(c, "contact persons retrieved", rows)
}

func (h *Handler) CreateContactPerson(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	partyID, err := strconv.ParseUint(c.Params("partyId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid party id")
	}
	var req CreateContactPersonRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	cp, err := h.svc.CreateContactPerson(c.Context(), tenantID, uint(partyID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "contact person created", cp)
}

func (h *Handler) UpdateContactPerson(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateContactPersonRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	cp, err := h.svc.UpdateContactPerson(c.Context(), tenantID, uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "contact person updated", cp)
}

func (h *Handler) DeleteContactPerson(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteContactPerson(c.Context(), tenantID, uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Addresses ─────────────────────────────────────────────────────────────────

func (h *Handler) ListAddresses(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	partyID, err := strconv.ParseUint(c.Params("partyId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid party id")
	}
	rows, err := h.svc.ListAddresses(c.Context(), tenantID, uint(partyID))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list addresses")
	}
	return httputil.Success(c, "addresses retrieved", rows)
}

func (h *Handler) CreateAddress(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	partyID, err := strconv.ParseUint(c.Params("partyId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid party id")
	}
	var req CreateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	a, err := h.svc.CreateAddress(c.Context(), tenantID, uint(partyID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "address created", a)
}

func (h *Handler) UpdateAddress(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	a, err := h.svc.UpdateAddress(c.Context(), tenantID, uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "address updated", a)
}

func (h *Handler) DeleteAddress(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteAddress(c.Context(), tenantID, uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}
