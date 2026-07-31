package sales

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

// ── Sales Quotations ──────────────────────────────────────────────────────────

func (h *Handler) ListSQs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var customerID *uint
	if s := c.Query("customer_id"); s != "" {
		if cid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(cid)
			customerID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountSQs(c.Context(), tenantID, customerID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count quotations")
	}
	rows, err := h.svc.ListSQs(c.Context(), tenantID, customerID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list quotations")
	}
	return httputil.SuccessWithMeta(c, "quotations retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	sq, err := h.svc.GetSQ(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "quotation not found")
	}
	return httputil.Success(c, "quotation retrieved", sq)
}

func (h *Handler) CreateSQ(c *fiber.Ctx) error {
	var req CreateSQRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	sq, err := h.svc.CreateSQ(c.Context(), tenantFromCtx(c), userFromCtx(c), req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "quotation created", sq)
}

func (h *Handler) UpdateSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateSQRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	sq, err := h.svc.UpdateSQ(c.Context(), tenantFromCtx(c), id, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation updated", sq)
}

func (h *Handler) DeleteSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteSQ(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) SubmitSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	sq, err := h.svc.SubmitSQ(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation sent", sq)
}

func (h *Handler) AcceptSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	sq, so, err := h.svc.AcceptSQ(c.Context(), tenantFromCtx(c), userFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation accepted", fiber.Map{
		"quotation": sq,
		"sales_order": fiber.Map{
			"id":   so.ID,
			"code": so.Code,
		},
	})
}

func (h *Handler) RejectSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	sq, err := h.svc.RejectSQ(c.Context(), tenantFromCtx(c), userFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation rejected", sq)
}

func (h *Handler) CancelSQ(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	sq, err := h.svc.CancelSQ(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation cancelled", sq)
}

// ── SQ Lines ──────────────────────────────────────────────────────────────────

func (h *Handler) ListSQLines(c *fiber.Ctx) error {
	sqID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	rows, err := h.svc.ListSQLines(c.Context(), tenantFromCtx(c), sqID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation lines retrieved", rows)
}

func (h *Handler) GetSQLine(c *fiber.Ctx) error {
	sqID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	line, err := h.svc.GetSQLine(c.Context(), tenantFromCtx(c), sqID, lineID)
	if err != nil {
		return httputil.NotFound(c, "line not found")
	}
	return httputil.Success(c, "quotation line retrieved", line)
}

func (h *Handler) AddSQLine(c *fiber.Ctx) error {
	sqID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddSQLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddSQLine(c.Context(), tenantFromCtx(c), sqID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "quotation line added", line)
}

func (h *Handler) UpdateSQLine(c *fiber.Ctx) error {
	sqID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	var req UpdateSQLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	line, err := h.svc.UpdateSQLine(c.Context(), tenantFromCtx(c), sqID, lineID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quotation line updated", line)
}

func (h *Handler) DeleteSQLine(c *fiber.Ctx) error {
	sqID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	if err := h.svc.DeleteSQLine(c.Context(), tenantFromCtx(c), sqID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Sales Orders ──────────────────────────────────────────────────────────────

func (h *Handler) ListSOs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var customerID *uint
	if s := c.Query("customer_id"); s != "" {
		if cid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(cid)
			customerID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountSOs(c.Context(), tenantID, customerID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count sales orders")
	}
	rows, err := h.svc.ListSOs(c.Context(), tenantID, customerID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list sales orders")
	}
	return httputil.SuccessWithMeta(c, "sales orders retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetSO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	so, err := h.svc.GetSO(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "sales order not found")
	}
	return httputil.Success(c, "sales order retrieved", so)
}

func (h *Handler) CreateSO(c *fiber.Ctx) error {
	var req CreateSORequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	so, err := h.svc.CreateSO(c.Context(), tenantFromCtx(c), userFromCtx(c), req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "sales order created", so)
}

func (h *Handler) UpdateSO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateSORequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	so, err := h.svc.UpdateSO(c.Context(), tenantFromCtx(c), id, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "sales order updated", so)
}

func (h *Handler) DeleteSO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteSO(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) ConfirmSO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ConfirmSO(c.Context(), tenantFromCtx(c), userFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	so, _ := h.svc.GetSO(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "sales order confirmed", so)
}

func (h *Handler) CancelSO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelSO(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	so, _ := h.svc.GetSO(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "sales order cancelled", so)
}

// ── SO Lines ──────────────────────────────────────────────────────────────────

func (h *Handler) ListSOLines(c *fiber.Ctx) error {
	soID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	items, err := h.svc.ListSOLines(c.Context(), tenantFromCtx(c), soID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "so items retrieved", items)
}

func (h *Handler) GetSOLine(c *fiber.Ctx) error {
	soID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	item, err := h.svc.GetSOLine(c.Context(), tenantFromCtx(c), soID, itemID)
	if err != nil {
		return httputil.NotFound(c, "item not found")
	}
	return httputil.Success(c, "so item retrieved", item)
}

func (h *Handler) AddSOLine(c *fiber.Ctx) error {
	soID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddSOLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.AddSOLine(c.Context(), tenantFromCtx(c), soID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "item added", item)
}

func (h *Handler) UpdateSOLine(c *fiber.Ctx) error {
	soID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	var req UpdateSOLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.UpdateSOLine(c.Context(), tenantFromCtx(c), soID, itemID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "item updated", item)
}

func (h *Handler) DeleteSOLine(c *fiber.Ctx) error {
	soID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	if err := h.svc.DeleteSOLine(c.Context(), tenantFromCtx(c), soID, itemID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Delivery Orders ───────────────────────────────────────────────────────────

func (h *Handler) ListDOs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var soID *uint
	if s := c.Query("so_id"); s != "" {
		if sid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(sid)
			soID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountDOs(c.Context(), tenantID, soID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count delivery orders")
	}
	rows, err := h.svc.ListDOs(c.Context(), tenantID, soID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list delivery orders")
	}
	return httputil.SuccessWithMeta(c, "delivery orders retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetDO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	do, err := h.svc.GetDO(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "delivery order not found")
	}
	return httputil.Success(c, "delivery order retrieved", do)
}

func (h *Handler) CreateDO(c *fiber.Ctx) error {
	var req CreateDORequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	do, err := h.svc.CreateDO(c.Context(), tenantFromCtx(c), userFromCtx(c), req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "delivery order created", do)
}

func (h *Handler) UpdateDO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateDORequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	do, err := h.svc.UpdateDO(c.Context(), tenantFromCtx(c), id, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "delivery order updated", do)
}

func (h *Handler) ConfirmDO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ConfirmDO(c.Context(), tenantFromCtx(c), userFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	do, _ := h.svc.GetDO(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "delivery order confirmed", do)
}

func (h *Handler) CancelDO(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelDO(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	do, _ := h.svc.GetDO(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "delivery order cancelled", do)
}

// ── DO Lines ──────────────────────────────────────────────────────────────────

func (h *Handler) ListDOLines(c *fiber.Ctx) error {
	doID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	items, err := h.svc.ListDOLines(c.Context(), tenantFromCtx(c), doID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "do items retrieved", items)
}

func (h *Handler) GetDOLine(c *fiber.Ctx) error {
	doID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	item, err := h.svc.GetDOLine(c.Context(), tenantFromCtx(c), doID, itemID)
	if err != nil {
		return httputil.NotFound(c, "item not found")
	}
	return httputil.Success(c, "do item retrieved", item)
}

func (h *Handler) AddDOLine(c *fiber.Ctx) error {
	doID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddDOLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.AddDOLine(c.Context(), tenantFromCtx(c), doID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "item added", item)
}

func (h *Handler) UpdateDOLine(c *fiber.Ctx) error {
	doID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	var req UpdateDOLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	item, err := h.svc.UpdateDOLine(c.Context(), tenantFromCtx(c), doID, itemID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "item updated", item)
}

func (h *Handler) DeleteDOLine(c *fiber.Ctx) error {
	doID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	itemID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid item id")
	}
	if err := h.svc.DeleteDOLine(c.Context(), tenantFromCtx(c), doID, itemID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Sales Invoices ────────────────────────────────────────────────────────────

func (h *Handler) ListSIs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	var customerID *uint
	if s := c.Query("customer_id"); s != "" {
		if cid, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(cid)
			customerID = &v
		}
	}
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountSIs(c.Context(), tenantID, customerID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count sales invoices")
	}
	rows, err := h.svc.ListSIs(c.Context(), tenantID, customerID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list sales invoices")
	}
	return httputil.SuccessWithMeta(c, "sales invoices retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetSI(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	si, err := h.svc.GetSI(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "sales invoice not found")
	}
	return httputil.Success(c, "sales invoice retrieved", si)
}

func (h *Handler) CreateSI(c *fiber.Ctx) error {
	var req CreateSIRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	si, err := h.svc.CreateSI(c.Context(), tenantFromCtx(c), userFromCtx(c), req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "sales invoice created", si)
}

func (h *Handler) ListSILines(c *fiber.Ctx) error {
	invoiceID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	items, err := h.svc.ListSILines(c.Context(), tenantFromCtx(c), invoiceID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "si lines retrieved", items)
}

func (h *Handler) AddSILine(c *fiber.Ctx) error {
	invoiceID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddSILineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddSILine(c.Context(), tenantFromCtx(c), invoiceID, req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) DeleteSILine(c *fiber.Ctx) error {
	invoiceID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid invoice id")
	}
	lineID, err := parseID(c, "lineId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteSILine(c.Context(), tenantFromCtx(c), invoiceID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

func (h *Handler) PostSI(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.PostSI(c.Context(), tenantFromCtx(c), userFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	si, _ := h.svc.GetSI(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "sales invoice posted", si)
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
	if err := h.svc.RecordPayment(c.Context(), tenantFromCtx(c), id, req); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	si, _ := h.svc.GetSI(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "payment recorded", si)
}

func (h *Handler) CancelSI(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelSI(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	si, _ := h.svc.GetSI(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "sales invoice cancelled", si)
}
