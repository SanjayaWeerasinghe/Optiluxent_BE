package procurement

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

// ── Purchase Requests ─────────────────────────────────────────────────────────

func (h *Handler) ListPRs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountPRs(c.Context(), tenantID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count purchase requests")
	}
	rows, err := h.svc.ListPRs(c.Context(), tenantID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list purchase requests")
	}
	return httputil.SuccessWithMeta(c, "purchase requests retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetPR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pr, err := h.svc.GetPR(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "purchase request not found")
	}
	return httputil.Success(c, "purchase request retrieved", pr)
}

func (h *Handler) CreatePR(c *fiber.Ctx) error {
	var req CreatePRRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	pr, err := h.svc.CreatePR(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "purchase request created", pr)
}

func (h *Handler) UpdatePR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdatePRRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	pr, err := h.svc.UpdatePR(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase request updated", pr)
}

func (h *Handler) DeletePR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeletePR(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) SubmitPR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pr, err := h.svc.SubmitPR(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase request submitted for approval", pr)
}

func (h *Handler) ApprovePR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pr, err := h.svc.ApprovePR(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase request approved", pr)
}

func (h *Handler) RejectPR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pr, err := h.svc.RejectPR(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase request rejected", pr)
}

func (h *Handler) CancelPR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	pr, err := h.svc.CancelPR(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase request cancelled", pr)
}

// ── PR Items ──────────────────────────────────────────────────────────────────

func (h *Handler) ListPRItems(c *fiber.Ctx) error {
	prID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	items, err := h.svc.ListPRItems(c.Context(), tenantFromCtx(c), prID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "pr items retrieved", items)
}

func (h *Handler) GetPRItem(c *fiber.Ctx) error {
	_, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	item, err := h.svc.GetPRItem(c.Context(), tenantFromCtx(c), itemID)
	if err != nil {
		return httputil.NotFound(c, "item not found")
	}
	return httputil.Success(c, "pr item retrieved", item)
}

func (h *Handler) AddPRItem(c *fiber.Ctx) error {
	prID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddPRItemRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.AddPRItem(c.Context(), tenantFromCtx(c), prID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "item added", item)
}

func (h *Handler) UpdatePRItem(c *fiber.Ctx) error {
	prID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	var req UpdatePRItemRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.UpdatePRItem(c.Context(), tenantFromCtx(c), prID, itemID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "item updated", item)
}

func (h *Handler) DeletePRItem(c *fiber.Ctx) error {
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	if err := h.svc.DeletePRItem(c.Context(), tenantFromCtx(c), itemID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Purchase Orders ───────────────────────────────────────────────────────────

func (h *Handler) ListPOs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var supplierID *uint
	if s := c.Query("supplier_id"); s != "" {
		if sid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(sid)
			supplierID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountPOs(c.Context(), tenantID, status, supplierID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count purchase orders")
	}
	rows, err := h.svc.ListPOs(c.Context(), tenantID, status, supplierID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list purchase orders")
	}
	return httputil.SuccessWithMeta(c, "purchase orders retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetPO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	po, err := h.svc.GetPO(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "purchase order not found")
	}
	return httputil.Success(c, "purchase order retrieved", po)
}

func (h *Handler) CreatePO(c *fiber.Ctx) error {
	var req CreatePORequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	po, err := h.svc.CreatePO(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "purchase order created", po)
}

func (h *Handler) UpdatePO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdatePORequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	po, err := h.svc.UpdatePO(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase order updated", po)
}

func (h *Handler) DeletePO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeletePO(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) ConfirmPO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	po, err := h.svc.ConfirmPO(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase order confirmed", po)
}

func (h *Handler) CancelPO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	po, err := h.svc.CancelPO(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "purchase order cancelled", po)
}

// ── PO Items ──────────────────────────────────────────────────────────────────

func (h *Handler) ListPOItems(c *fiber.Ctx) error {
	poID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	items, err := h.svc.ListPOItems(c.Context(), tenantFromCtx(c), poID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "po items retrieved", items)
}

func (h *Handler) GetPOItem(c *fiber.Ctx) error {
	_, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	item, err := h.svc.GetPOItem(c.Context(), tenantFromCtx(c), itemID)
	if err != nil {
		return httputil.NotFound(c, "item not found")
	}
	return httputil.Success(c, "po item retrieved", item)
}

func (h *Handler) AddPOItem(c *fiber.Ctx) error {
	poID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddPOItemRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.AddPOItem(c.Context(), tenantFromCtx(c), poID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "item added", item)
}

func (h *Handler) UpdatePOItem(c *fiber.Ctx) error {
	poID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	var req UpdatePOItemRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.UpdatePOItem(c.Context(), tenantFromCtx(c), poID, itemID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "item updated", item)
}

func (h *Handler) DeletePOItem(c *fiber.Ctx) error {
	poID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	if err := h.svc.DeletePOItem(c.Context(), tenantFromCtx(c), poID, itemID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Goods Receipts ────────────────────────────────────────────────────────────

func (h *Handler) ListGRNs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var poID *uint
	if s := c.Query("po_id"); s != "" {
		if pid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(pid)
			poID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountGRNs(c.Context(), tenantID, status, poID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count goods receipts")
	}
	rows, err := h.svc.ListGRNs(c.Context(), tenantID, status, poID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list goods receipts")
	}
	return httputil.SuccessWithMeta(c, "goods receipts retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetGRN(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	grn, err := h.svc.GetGRN(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "goods receipt not found")
	}
	return httputil.Success(c, "goods receipt retrieved", grn)
}

func (h *Handler) CreateGRN(c *fiber.Ctx) error {
	var req CreateGRNRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	grn, err := h.svc.CreateGRN(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "goods receipt created", grn)
}

func (h *Handler) ConfirmGRN(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ConfirmGRN(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	grn, _ := h.svc.GetGRN(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "goods receipt confirmed", grn)
}

func (h *Handler) CancelGRN(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelGRN(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	grn, _ := h.svc.GetGRN(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "goods receipt cancelled", grn)
}

// ── GRN Items ─────────────────────────────────────────────────────────────────

func (h *Handler) ListGRNItems(c *fiber.Ctx) error {
	grnID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	items, err := h.svc.ListGRNItems(c.Context(), tenantFromCtx(c), grnID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "grn items retrieved", items)
}

func (h *Handler) GetGRNItem(c *fiber.Ctx) error {
	_, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	item, err := h.svc.GetGRNItem(c.Context(), tenantFromCtx(c), itemID)
	if err != nil {
		return httputil.NotFound(c, "item not found")
	}
	return httputil.Success(c, "grn item retrieved", item)
}

func (h *Handler) AddGRNItem(c *fiber.Ctx) error {
	grnID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddGRNItemRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.AddGRNItem(c.Context(), tenantFromCtx(c), grnID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "item added", item)
}

func (h *Handler) UpdateGRNItem(c *fiber.Ctx) error {
	grnID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	var req UpdateGRNItemRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.UpdateGRNItem(c.Context(), tenantFromCtx(c), grnID, itemID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "item updated", item)
}

func (h *Handler) AutoInvoiceFromGRN(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	inv, err := h.svc.AutoInvoiceFromGRN(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "invoice auto-generated from goods receipt", inv)
}

// ── Purchase Invoices ─────────────────────────────────────────────────────────

func (h *Handler) ListInvoices(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var supplierID *uint
	if s := c.Query("supplier_id"); s != "" {
		if sid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(sid)
			supplierID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountInvoices(c.Context(), tenantID, status, supplierID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count invoices")
	}
	rows, err := h.svc.ListInvoices(c.Context(), tenantID, status, supplierID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list invoices")
	}
	return httputil.SuccessWithMeta(c, "invoices retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetInvoice(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	inv, err := h.svc.GetInvoice(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "invoice not found")
	}
	return httputil.Success(c, "invoice retrieved", inv)
}

func (h *Handler) CreateInvoice(c *fiber.Ctx) error {
	var req CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	inv, err := h.svc.CreateInvoice(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "invoice created", inv)
}

func (h *Handler) ListInvoiceLines(c *fiber.Ctx) error {
	invoiceID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	inv, err := h.svc.GetInvoice(c.Context(), tenantFromCtx(c), invoiceID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "invoice lines retrieved", inv.Lines)
}

func (h *Handler) AddInvoiceLine(c *fiber.Ctx) error {
	invoiceID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddInvoiceLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddInvoiceLine(c.Context(), tenantFromCtx(c), invoiceID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) DeleteInvoiceLine(c *fiber.Ctx) error {
	invoiceID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid invoice id")
	}
	lineID, err := parseID(c, "lineId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteInvoiceLine(c.Context(), tenantFromCtx(c), invoiceID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) RecordPayment(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req RecordPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	inv, err := h.svc.RecordPayment(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "payment recorded", inv)
}

func (h *Handler) PostInvoice(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	inv, err := h.svc.PostInvoice(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "invoice posted", inv)
}

func (h *Handler) CancelInvoice(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	inv, err := h.svc.CancelInvoice(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "invoice cancelled", inv)
}
