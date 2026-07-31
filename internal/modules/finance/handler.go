package finance

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

// ── Payments ────────────────────────────────────────────────────────────────

func (h *Handler) RecordPayment(c *fiber.Ctx) error {
	var req RecordPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	p, err := h.svc.RecordPayment(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "payment recorded", p)
}

func (h *Handler) ListPayments(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	invoiceKind := c.Query("invoice_kind")
	invoiceID, _ := strconv.ParseUint(c.Query("invoice_id", "0"), 10, 64)
	partyID, _ := strconv.ParseUint(c.Query("party_id", "0"), 10, 64)
	direction := c.Query("direction")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountPayments(c.Context(), tenantID, invoiceKind, uint(invoiceID), uint(partyID), direction)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	rows, err := h.svc.ListPayments(c.Context(), tenantID,
		invoiceKind, uint(invoiceID), uint(partyID), direction, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.SuccessWithMeta(c, "payments retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetPayment(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	p, err := h.svc.GetPayment(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "payment not found")
	}
	return httputil.Success(c, "payment retrieved", p)
}

// ── Journal Entries ─────────────────────────────────────────────────────────

func (h *Handler) ListJournalEntries(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	sourceType := c.Query("source_type")
	sourceID, _ := strconv.ParseUint(c.Query("source_id", "0"), 10, 64)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountJournalEntries(c.Context(), tenantID, sourceType, uint(sourceID))
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	rows, err := h.svc.ListJournalEntries(c.Context(), tenantID,
		sourceType, uint(sourceID), limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.SuccessWithMeta(c, "journal entries retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetJournalEntry(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	je, err := h.svc.GetJournalEntry(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "journal entry not found")
	}
	return httputil.Success(c, "journal entry retrieved", je)
}

// ── Reports ─────────────────────────────────────────────────────────────────

func (h *Handler) AgingReport(c *fiber.Ctx) error {
	kind := c.Query("kind", "AR")
	asOf := c.Query("as_of")
	rows, err := h.svc.AgingReport(c.Context(), tenantFromCtx(c), kind, asOf)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "aging report retrieved", rows)
}

func (h *Handler) PartyOutstanding(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid party id")
	}
	kind := c.Query("kind", "AR")
	total, err := h.svc.PartyOutstanding(c.Context(), tenantFromCtx(c), id, kind)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "outstanding retrieved", PartyOutstanding{PartyID: id, Kind: kind, Outstanding: total})
}

// ── Settings ────────────────────────────────────────────────────────────────

func (h *Handler) GetSettings(c *fiber.Ctx) error {
	s, err := h.svc.GetSettings(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "finance settings retrieved", s)
}

func (h *Handler) UpdateSettings(c *fiber.Ctx) error {
	var req UpdateSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	s, err := h.svc.UpdateSettings(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "finance settings saved", s)
}
