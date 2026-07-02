package financial

import (
	"strconv"
	"strings"
	"time"

	httputil "erp-system/pkg/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler exposes HTTP endpoints for the Financial submodule.
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
	fields := make(map[string]interface{})
	for _, fe := range err.(validator.ValidationErrors) {
		fields[strings.ToLower(fe.Field())] = fe.Tag()
	}
	return fields
}

// ── Currencies ────────────────────────────────────────────────────────────────

// GET /financial/currencies
func (h *Handler) ListCurrencies(c *fiber.Ctx) error {
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountCurrencies(c.UserContext())
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count currencies")
	}
	list, err := h.svc.ListCurrencies(c.UserContext(), limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list currencies")
	}
	return httputil.SuccessWithMeta(c, "Currencies retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/currencies
func (h *Handler) CreateCurrency(c *fiber.Ctx) error {
	var req CreateCurrencyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	cur := &Currency{Code: req.Code, Name: req.Name, Symbol: req.Symbol}
	if err := h.svc.CreateCurrency(c.UserContext(), cur); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Currency created", cur)
}

// PUT /financial/currencies/:id
func (h *Handler) UpdateCurrency(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid currency ID")
	}
	var req UpdateCurrencyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	cur, err := h.svc.GetCurrency(c.UserContext(), uint(id))
	if err != nil || cur == nil {
		return httputil.NotFound(c, "Currency not found")
	}

	if req.Name != "" {
		cur.Name = req.Name
	}
	if req.Symbol != "" {
		cur.Symbol = req.Symbol
	}
	if req.IsActive != nil {
		cur.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateCurrency(c.UserContext(), cur); err != nil {
		return httputil.InternalServerError(c, "Failed to update currency")
	}
	return httputil.Success(c, "Currency updated", cur)
}

// PUT /financial/currencies/:id/set-base
func (h *Handler) SetBaseCurrency(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid currency ID")
	}
	if err := h.svc.SetBaseCurrency(c.UserContext(), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "Base currency updated", nil)
}

// ── Exchange Rates ────────────────────────────────────────────────────────────

// GET /financial/exchange-rates
func (h *Handler) ListExchangeRates(c *fiber.Ctx) error {
	var fromCurrID *uint
	if raw := c.Query("from_currency_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			u := uint(v)
			fromCurrID = &u
		}
	}
	var from, to *time.Time
	if raw := c.Query("from"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			from = &t
		}
	}
	if raw := c.Query("to"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			to = &t
		}
	}
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountExchangeRates(c.UserContext(), tenantFromCtx(c), fromCurrID, from, to)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count exchange rates")
	}
	list, err := h.svc.ListExchangeRates(c.UserContext(), tenantFromCtx(c), fromCurrID, from, to, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list exchange rates")
	}
	return httputil.SuccessWithMeta(c, "Exchange rates retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// GET /financial/exchange-rates/latest
func (h *Handler) GetLatestRates(c *fiber.Ctx) error {
	list, err := h.svc.GetLatestRates(c.UserContext(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get latest rates")
	}
	return httputil.Success(c, "Latest exchange rates retrieved", list)
}

// POST /financial/exchange-rates
func (h *Handler) CreateExchangeRate(c *fiber.Ctx) error {
	var req CreateExchangeRateRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	rate := &ExchangeRate{
		TenantID:       tenantFromCtx(c),
		FromCurrencyID: req.FromCurrencyID,
		ToCurrencyID:   req.ToCurrencyID,
		Rate:           req.Rate,
		EffectiveDate:  req.EffectiveDate,
		Source:         source,
	}
	if err := h.svc.CreateExchangeRate(c.UserContext(), rate); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "Exchange rate added", rate)
}

// ── Chart of Accounts ─────────────────────────────────────────────────────────

// GET /financial/chart-of-accounts
func (h *Handler) ListCoA(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountCoA(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count chart of accounts")
	}
	list, err := h.svc.ListCoA(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list chart of accounts")
	}
	return httputil.SuccessWithMeta(c, "Chart of accounts retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/chart-of-accounts
func (h *Handler) CreateCoA(c *fiber.Ctx) error {
	var req CreateCoARequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	level := 1
	if req.ParentID != nil {
		level = 2
	}

	a := &ChartOfAccount{
		TenantID:    tenantFromCtx(c),
		Code:        req.Code,
		Name:        req.Name,
		AccountType: req.AccountType,
		ParentID:    req.ParentID,
		CurrencyID:  req.CurrencyID,
		Level:       level,
	}
	if err := h.svc.CreateCoA(c.UserContext(), a); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Account created", a)
}

// GET /financial/chart-of-accounts/:id
func (h *Handler) GetCoA(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid account ID")
	}
	a, err := h.svc.GetCoA(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get account")
	}
	if a == nil {
		return httputil.NotFound(c, "Account not found")
	}
	return httputil.Success(c, "Account retrieved", a)
}

// PUT /financial/chart-of-accounts/:id
func (h *Handler) UpdateCoA(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid account ID")
	}
	var req UpdateCoARequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	a, err := h.svc.GetCoA(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || a == nil {
		return httputil.NotFound(c, "Account not found")
	}

	if req.Name != "" {
		a.Name = req.Name
	}
	if req.IsControlled != nil {
		a.IsControlled = *req.IsControlled
	}
	if req.IsActive != nil {
		a.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateCoA(c.UserContext(), a); err != nil {
		return httputil.InternalServerError(c, "Failed to update account")
	}
	return httputil.Success(c, "Account updated", a)
}

// DELETE /financial/chart-of-accounts/:id
func (h *Handler) DeleteCoA(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid account ID")
	}
	if err := h.svc.DeleteCoA(c.UserContext(), uint(id), tenantFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "Account deactivated", nil)
}

// ── Cost Centers ──────────────────────────────────────────────────────────────

// GET /financial/cost-centers
func (h *Handler) ListCostCenters(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountCostCenters(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count cost centers")
	}
	list, err := h.svc.ListCostCenters(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list cost centers")
	}
	return httputil.SuccessWithMeta(c, "Cost centers retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/cost-centers
func (h *Handler) CreateCostCenter(c *fiber.Ctx) error {
	var req CreateCostCenterRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	cc := &CostCenter{
		TenantID:     tenantFromCtx(c),
		Code:         req.Code,
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
	}
	if err := h.svc.CreateCostCenter(c.UserContext(), cc); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Cost center created", cc)
}

// PUT /financial/cost-centers/:id
func (h *Handler) UpdateCostCenter(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid cost center ID")
	}
	var req UpdateCostCenterRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	cc, err := h.svc.GetCostCenter(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || cc == nil {
		return httputil.NotFound(c, "Cost center not found")
	}

	if req.Name != "" {
		cc.Name = req.Name
	}
	if req.DepartmentID != nil {
		cc.DepartmentID = req.DepartmentID
	}
	if req.IsActive != nil {
		cc.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateCostCenter(c.UserContext(), cc); err != nil {
		return httputil.InternalServerError(c, "Failed to update cost center")
	}
	return httputil.Success(c, "Cost center updated", cc)
}

// ── Payment Terms ─────────────────────────────────────────────────────────────

// GET /financial/payment-terms
func (h *Handler) ListPaymentTerms(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountPaymentTerms(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count payment terms")
	}
	list, err := h.svc.ListPaymentTerms(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list payment terms")
	}
	return httputil.SuccessWithMeta(c, "Payment terms retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/payment-terms
func (h *Handler) CreatePaymentTerm(c *fiber.Ctx) error {
	var req CreatePaymentTermRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	pt := &PaymentTerm{
		TenantID:        tenantFromCtx(c),
		Code:            req.Code,
		Name:            req.Name,
		DueDays:         req.DueDays,
		DiscountDays:    req.DiscountDays,
		DiscountPercent: req.DiscountPercent,
	}
	if err := h.svc.CreatePaymentTerm(c.UserContext(), pt); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Payment term created", pt)
}

// PUT /financial/payment-terms/:id
func (h *Handler) UpdatePaymentTerm(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid payment term ID")
	}
	var req UpdatePaymentTermRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	pt, err := h.svc.GetPaymentTerm(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || pt == nil {
		return httputil.NotFound(c, "Payment term not found")
	}

	if req.Name != "" {
		pt.Name = req.Name
	}
	if req.DueDays >= 0 {
		pt.DueDays = req.DueDays
	}
	if req.DiscountDays >= 0 {
		pt.DiscountDays = req.DiscountDays
	}
	if req.DiscountPercent >= 0 {
		pt.DiscountPercent = req.DiscountPercent
	}
	if req.IsActive != nil {
		pt.IsActive = *req.IsActive
	}

	if err := h.svc.UpdatePaymentTerm(c.UserContext(), pt); err != nil {
		return httputil.InternalServerError(c, "Failed to update payment term")
	}
	return httputil.Success(c, "Payment term updated", pt)
}

// ── Banks ─────────────────────────────────────────────────────────────────────

// GET /financial/banks
func (h *Handler) ListBanks(c *fiber.Ctx) error {
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountBanks(c.UserContext())
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count banks")
	}
	list, err := h.svc.ListBanks(c.UserContext(), limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list banks")
	}
	return httputil.SuccessWithMeta(c, "Banks retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/banks
func (h *Handler) CreateBank(c *fiber.Ctx) error {
	var req CreateBankRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	b := &Bank{
		Name:       req.Name,
		BranchName: req.BranchName,
		SwiftCode:  req.SwiftCode,
		Address:    req.Address,
	}
	if err := h.svc.CreateBank(c.UserContext(), b); err != nil {
		return httputil.InternalServerError(c, "Failed to create bank")
	}
	return httputil.Created(c, "Bank created", b)
}

// ── Company Bank Accounts ─────────────────────────────────────────────────────

// GET /financial/bank-accounts
func (h *Handler) ListBankAccounts(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountBankAccounts(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count bank accounts")
	}
	list, err := h.svc.ListBankAccounts(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list bank accounts")
	}
	return httputil.SuccessWithMeta(c, "Bank accounts retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/bank-accounts
func (h *Handler) CreateBankAccount(c *fiber.Ctx) error {
	var req CreateBankAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	ba := &CompanyBankAccount{
		TenantID:      tenantFromCtx(c),
		BankID:        req.BankID,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		CurrencyID:    req.CurrencyID,
		GLAccountID:   req.GLAccountID,
		IsDefault:     req.IsDefault,
	}
	if err := h.svc.CreateBankAccount(c.UserContext(), ba); err != nil {
		return httputil.InternalServerError(c, "Failed to create bank account")
	}
	return httputil.Created(c, "Bank account created", ba)
}

// PUT /financial/bank-accounts/:id
func (h *Handler) UpdateBankAccount(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid bank account ID")
	}
	var req UpdateBankAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	ba, err := h.svc.GetBankAccount(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || ba == nil {
		return httputil.NotFound(c, "Bank account not found")
	}

	if req.AccountName != "" {
		ba.AccountName = req.AccountName
	}
	if req.IsDefault != nil {
		ba.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		ba.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateBankAccount(c.UserContext(), ba); err != nil {
		return httputil.InternalServerError(c, "Failed to update bank account")
	}
	return httputil.Success(c, "Bank account updated", ba)
}

// ── Tax Codes ─────────────────────────────────────────────────────────────────

// GET /financial/tax-codes
func (h *Handler) ListTaxCodes(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountTaxCodes(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count tax codes")
	}
	list, err := h.svc.ListTaxCodes(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list tax codes")
	}
	return httputil.SuccessWithMeta(c, "Tax codes retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/tax-codes
func (h *Handler) CreateTaxCode(c *fiber.Ctx) error {
	var req CreateTaxCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	t := &TaxCode{
		TenantID:    tenantFromCtx(c),
		Code:        req.Code,
		Name:        req.Name,
		TaxType:     req.TaxType,
		Rate:        req.Rate,
		GLAccountID: req.GLAccountID,
	}
	if err := h.svc.CreateTaxCode(c.UserContext(), t); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Tax code created", t)
}

// PUT /financial/tax-codes/:id
func (h *Handler) UpdateTaxCode(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid tax code ID")
	}
	var req UpdateTaxCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	t, err := h.svc.GetTaxCode(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || t == nil {
		return httputil.NotFound(c, "Tax code not found")
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Rate >= 0 {
		t.Rate = req.Rate
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateTaxCode(c.UserContext(), t); err != nil {
		return httputil.InternalServerError(c, "Failed to update tax code")
	}
	return httputil.Success(c, "Tax code updated", t)
}

// ── Tax Groups ────────────────────────────────────────────────────────────────

// GET /financial/tax-groups
func (h *Handler) ListTaxGroups(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountTaxGroups(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count tax groups")
	}
	list, err := h.svc.ListTaxGroups(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list tax groups")
	}
	return httputil.SuccessWithMeta(c, "Tax groups retrieved", list, httputil.Paginate(page, perPage, int(total)))
}

// POST /financial/tax-groups
func (h *Handler) CreateTaxGroup(c *fiber.Ctx) error {
	var req CreateTaxGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	taxCodes := make([]TaxCode, len(req.TaxCodeIDs))
	for i, tcID := range req.TaxCodeIDs {
		taxCodes[i] = TaxCode{ID: tcID}
	}

	tg := &TaxGroup{
		TenantID: tenantFromCtx(c),
		Name:     req.Name,
		TaxCodes: taxCodes,
	}
	if err := h.svc.CreateTaxGroup(c.UserContext(), tg); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Tax group created", tg)
}

// PUT /financial/tax-groups/:id
func (h *Handler) UpdateTaxGroup(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid tax group ID")
	}
	var req UpdateTaxGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	tg, err := h.svc.GetTaxGroup(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || tg == nil {
		return httputil.NotFound(c, "Tax group not found")
	}

	if req.Name != "" {
		tg.Name = req.Name
	}
	if req.IsActive != nil {
		tg.IsActive = *req.IsActive
	}
	if req.TaxCodeIDs != nil {
		taxCodes := make([]TaxCode, len(req.TaxCodeIDs))
		for i, tcID := range req.TaxCodeIDs {
			taxCodes[i] = TaxCode{ID: tcID}
		}
		tg.TaxCodes = taxCodes
	}

	if err := h.svc.UpdateTaxGroup(c.UserContext(), tg); err != nil {
		return httputil.InternalServerError(c, "Failed to update tax group")
	}
	return httputil.Success(c, "Tax group updated", tg)
}
