package inventory

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	httputil "erp-system/pkg/http"
)

// Handler holds the service and validator for all inventory HTTP handlers.
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

func parseOptionalUint(s string) *uint {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil
	}
	u := uint(v)
	return &u
}

// ── Material Requests ─────────────────────────────────────────────────────────

func (h *Handler) ListMRs(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountMRs(c.Context(), tenantID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count material requests")
	}
	rows, err := h.svc.ListMRs(c.Context(), tenantID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list material requests")
	}
	return httputil.SuccessWithMeta(c, "material requests retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetMR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	mr, err := h.svc.GetMR(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "material request not found")
	}
	return httputil.Success(c, "material request retrieved", mr)
}

func (h *Handler) CreateMR(c *fiber.Ctx) error {
	var req CreateMRRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	mr, err := h.svc.CreateMR(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "material request created", mr)
}

func (h *Handler) UpdateMR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateMRRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	mr, err := h.svc.UpdateMR(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "material request updated", mr)
}

func (h *Handler) SubmitMR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.SubmitMR(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	mr, _ := h.svc.GetMR(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "material request submitted", mr)
}

func (h *Handler) CancelMR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelMR(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	mr, _ := h.svc.GetMR(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "material request cancelled", mr)
}

func (h *Handler) ApproveMR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ApproveMR(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	mr, _ := h.svc.GetMR(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "material request approved", mr)
}

func (h *Handler) RejectMR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req RejectMRRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	if err := h.svc.RejectMR(c.Context(), tenantFromCtx(c), id, userFromCtx(c), req.Reason); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	mr, _ := h.svc.GetMR(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "material request rejected", mr)
}

func (h *Handler) ListMRLines(c *fiber.Ctx) error {
	mrID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lines, err := h.svc.ListMRLines(c.Context(), tenantFromCtx(c), mrID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "material request lines retrieved", lines)
}

func (h *Handler) AddMRLine(c *fiber.Ctx) error {
	mrID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddMRLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddMRLine(c.Context(), tenantFromCtx(c), mrID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) UpdateMRLine(c *fiber.Ctx) error {
	mrID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	var req UpdateMRLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.UpdateMRLine(c.Context(), tenantFromCtx(c), mrID, lineID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "line updated", line)
}

func (h *Handler) DeleteMRLine(c *fiber.Ctx) error {
	mrID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteMRLine(c.Context(), tenantFromCtx(c), mrID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Goods Transfers ───────────────────────────────────────────────────────────

func (h *Handler) ListTransfers(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountTransfers(c.Context(), tenantID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count goods transfers")
	}
	rows, err := h.svc.ListTransfers(c.Context(), tenantID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list goods transfers")
	}
	return httputil.SuccessWithMeta(c, "goods transfers retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetTransfer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	t, err := h.svc.GetTransfer(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "goods transfer not found")
	}
	return httputil.Success(c, "goods transfer retrieved", t)
}

func (h *Handler) CreateTransfer(c *fiber.Ctx) error {
	var req CreateTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	t, err := h.svc.CreateTransfer(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "goods transfer created", t)
}

func (h *Handler) UpdateTransfer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	t, err := h.svc.UpdateTransfer(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "goods transfer updated", t)
}

func (h *Handler) SendTransfer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.SendTransfer(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	t, _ := h.svc.GetTransfer(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "goods transfer sent", t)
}

func (h *Handler) ReceiveTransfer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ReceiveTransfer(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	t, _ := h.svc.GetTransfer(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "goods transfer received", t)
}

func (h *Handler) ListTransferLines(c *fiber.Ctx) error {
	transferID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lines, err := h.svc.ListTransferLines(c.Context(), tenantFromCtx(c), transferID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "transfer lines retrieved", lines)
}

func (h *Handler) AddTransferLine(c *fiber.Ctx) error {
	transferID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddTransferLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddTransferLine(c.Context(), tenantFromCtx(c), transferID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) UpdateTransferLine(c *fiber.Ctx) error {
	transferID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	var req UpdateTransferLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.UpdateTransferLine(c.Context(), tenantFromCtx(c), transferID, lineID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "line updated", line)
}

func (h *Handler) DeleteTransferLine(c *fiber.Ctx) error {
	transferID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteTransferLine(c.Context(), tenantFromCtx(c), transferID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Goods Issues ──────────────────────────────────────────────────────────────

func (h *Handler) ListIssues(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	status := c.Query("status")
	reason := c.Query("reason")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountIssues(c.Context(), tenantID, status, reason)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count goods issues")
	}
	rows, err := h.svc.ListIssues(c.Context(), tenantID, status, reason, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list goods issues")
	}
	return httputil.SuccessWithMeta(c, "goods issues retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetIssue(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	gi, err := h.svc.GetIssue(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "goods issue not found")
	}
	return httputil.Success(c, "goods issue retrieved", gi)
}

func (h *Handler) CreateIssue(c *fiber.Ctx) error {
	var req CreateIssueRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	gi, err := h.svc.CreateIssue(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "goods issue created", gi)
}

func (h *Handler) UpdateIssue(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateIssueRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	gi, err := h.svc.UpdateIssue(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "goods issue updated", gi)
}

func (h *Handler) ConfirmIssue(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ConfirmIssue(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	gi, _ := h.svc.GetIssue(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "goods issue confirmed", gi)
}

func (h *Handler) CancelIssue(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelIssue(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	gi, _ := h.svc.GetIssue(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "goods issue cancelled", gi)
}

func (h *Handler) ListIssueLines(c *fiber.Ctx) error {
	issueID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lines, err := h.svc.ListIssueLines(c.Context(), tenantFromCtx(c), issueID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "goods issue lines retrieved", lines)
}

func (h *Handler) AddIssueLine(c *fiber.Ctx) error {
	issueID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddIssueLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddIssueLine(c.Context(), tenantFromCtx(c), issueID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) UpdateIssueLine(c *fiber.Ctx) error {
	issueID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	var req UpdateIssueLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.UpdateIssueLine(c.Context(), tenantFromCtx(c), issueID, lineID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "line updated", line)
}

func (h *Handler) DeleteIssueLine(c *fiber.Ctx) error {
	issueID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteIssueLine(c.Context(), tenantFromCtx(c), issueID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Stock Adjustments ─────────────────────────────────────────────────────────

func (h *Handler) ListAdjustments(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountAdjustments(c.Context(), tenantID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count stock adjustments")
	}
	rows, err := h.svc.ListAdjustments(c.Context(), tenantID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list stock adjustments")
	}
	return httputil.SuccessWithMeta(c, "stock adjustments retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetAdjustment(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	sa, err := h.svc.GetAdjustment(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "stock adjustment not found")
	}
	return httputil.Success(c, "stock adjustment retrieved", sa)
}

func (h *Handler) CreateAdjustment(c *fiber.Ctx) error {
	var req CreateAdjustmentRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	sa, err := h.svc.CreateAdjustment(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "stock adjustment created", sa)
}

func (h *Handler) UpdateAdjustment(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateAdjustmentRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	sa, err := h.svc.UpdateAdjustment(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "stock adjustment updated", sa)
}

func (h *Handler) ConfirmAdjustment(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.ConfirmAdjustment(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	sa, _ := h.svc.GetAdjustment(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "stock adjustment confirmed", sa)
}

func (h *Handler) CancelAdjustment(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.CancelAdjustment(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	sa, _ := h.svc.GetAdjustment(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "stock adjustment cancelled", sa)
}

func (h *Handler) ListAdjustmentLines(c *fiber.Ctx) error {
	adjID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lines, err := h.svc.ListAdjustmentLines(c.Context(), tenantFromCtx(c), adjID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "adjustment lines retrieved", lines)
}

func (h *Handler) AddAdjustmentLine(c *fiber.Ctx) error {
	adjID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddAdjustmentLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddAdjustmentLine(c.Context(), tenantFromCtx(c), adjID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) UpdateAdjustmentLine(c *fiber.Ctx) error {
	adjID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	var req UpdateAdjustmentLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.UpdateAdjustmentLine(c.Context(), tenantFromCtx(c), adjID, lineID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "line updated", line)
}

func (h *Handler) DeleteAdjustmentLine(c *fiber.Ctx) error {
	adjID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	if err := h.svc.DeleteAdjustmentLine(c.Context(), tenantFromCtx(c), adjID, lineID); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Quality Checks ────────────────────────────────────────────────────────────

func (h *Handler) ListQualityChecks(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	status := c.Query("status")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountQualityChecks(c.Context(), tenantID, status)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count quality checks")
	}
	rows, err := h.svc.ListQualityChecks(c.Context(), tenantID, status, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list quality checks")
	}
	return httputil.SuccessWithMeta(c, "quality checks retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

func (h *Handler) GetQualityCheck(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	qc, err := h.svc.GetQualityCheck(c.Context(), tenantFromCtx(c), id)
	if err != nil {
		return httputil.NotFound(c, "quality check not found")
	}
	return httputil.Success(c, "quality check retrieved", qc)
}

func (h *Handler) CreateQualityCheck(c *fiber.Ctx) error {
	var req CreateQualityCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	qc, err := h.svc.CreateQualityCheck(c.Context(), tenantFromCtx(c), userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "quality check created", qc)
}

func (h *Handler) StartQualityCheck(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	qc, err := h.svc.StartQualityCheck(c.Context(), tenantFromCtx(c), id, userFromCtx(c))
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quality check started", qc)
}

func (h *Handler) SubmitQualityCheck(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.SubmitQualityCheck(c.Context(), tenantFromCtx(c), id, userFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	qc, _ := h.svc.GetQualityCheck(c.Context(), tenantFromCtx(c), id)
	return httputil.Success(c, "quality check submitted", qc)
}

func (h *Handler) ListQCLines(c *fiber.Ctx) error {
	checkID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lines, err := h.svc.ListQCLines(c.Context(), tenantFromCtx(c), checkID)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "quality check lines retrieved", lines)
}

func (h *Handler) AddQCLine(c *fiber.Ctx) error {
	checkID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req AddQCLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.AddQCLine(c.Context(), tenantFromCtx(c), checkID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "line added", line)
}

func (h *Handler) UpdateQCLine(c *fiber.Ctx) error {
	checkID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	lineID, err := parseID(c, "itemId")
	if err != nil {
		return httputil.BadRequest(c, "invalid line id")
	}
	var req UpdateQCLineRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	line, err := h.svc.UpdateQCLine(c.Context(), tenantFromCtx(c), checkID, lineID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "line updated", line)
}

// ── Stock Balances ────────────────────────────────────────────────────────────

func (h *Handler) GetStockBalance(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	warehouseID := parseOptionalUint(c.Query("warehouse_id"))
	productID := parseOptionalUint(c.Query("product_id"))

	rows, err := h.svc.GetStockBalance(c.Context(), tenantID, warehouseID, productID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to retrieve stock balances")
	}
	// Decorate each row with SUM(active allocations) so the FE can render
	// Available = Quantity − Reserved. The allocation service handles nil
	// gracefully so this stays working if wiring isn't in place yet.
	if alloc := h.svc.Allocation(); alloc != nil {
		for i := range rows {
			r := &rows[i]
			reserved, _ := alloc.ReservedByProductWarehouse(c.Context(), tenantID, r.ProductID, r.WarehouseID)
			r.ReservedQty = reserved
		}
	}
	return httputil.Success(c, "stock balances retrieved", rows)
}

// ListAllocations — flat list of ACTIVE stock reservations for the FE
// Allocations section. Optional filters: product_id, warehouse_id,
// source_type, source_doc_id.
func (h *Handler) ListAllocations(c *fiber.Ctx) error {
	if h.svc.Allocation() == nil {
		return httputil.InternalServerError(c, "allocation service not wired")
	}
	f := AllocationFilters{
		ProductID:   parseOptionalUint(c.Query("product_id")),
		WarehouseID: parseOptionalUint(c.Query("warehouse_id")),
		SourceType:  c.Query("source_type"),
		SourceDocID: parseOptionalUint(c.Query("source_doc_id")),
	}
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.Allocation().CountActive(c.Context(), tenantID, f)
	if err != nil {
		return httputil.InternalServerError(c, "failed to count allocations")
	}
	rows, err := h.svc.Allocation().ListActive(c.Context(), tenantID, f, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "failed to retrieve allocations")
	}
	return httputil.SuccessWithMeta(c, "allocations retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}
