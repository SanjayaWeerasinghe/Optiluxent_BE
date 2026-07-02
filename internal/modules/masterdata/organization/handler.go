package organization

import (
	"strconv"
	"strings"

	httputil "erp-system/pkg/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler exposes HTTP endpoints for the Organization submodule.
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

// GET /organization/company
func (h *Handler) GetCompany(c *fiber.Ctx) error {
	company, err := h.svc.GetCompany(c.UserContext())
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get company")
	}
	if company == nil {
		return httputil.NotFound(c, "Company not configured")
	}
	return httputil.Success(c, "Company retrieved", company)
}

// PUT /organization/company
func (h *Handler) SaveCompany(c *fiber.Ctx) error {
	var req SaveCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	existing, err := h.svc.GetCompany(c.UserContext())
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get company")
	}
	company := &Company{}
	if existing != nil {
		company = existing
	}

	company.Name = req.Name
	company.LegalName = req.LegalName
	company.TaxRegNumber = req.TaxRegNumber
	company.Logo = req.Logo
	company.Email = req.Email
	company.Phone = req.Phone
	company.Website = req.Website
	company.AddressLine1 = req.AddressLine1
	company.AddressLine2 = req.AddressLine2
	company.City = req.City
	company.StateID = req.StateID
	company.CountryID = req.CountryID
	company.PostalCode = req.PostalCode
	company.BaseCurrencyID = req.BaseCurrencyID
	if req.FiscalYearStart > 0 {
		company.FiscalYearStart = req.FiscalYearStart
	}

	if err := h.svc.SaveCompany(c.UserContext(), company); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "Company saved", company)
}

// GET /organization/departments
func (h *Handler) ListDepartments(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountDepartments(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count departments")
	}
	depts, err := h.svc.ListDepartments(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list departments")
	}
	return httputil.SuccessWithMeta(c, "Departments retrieved", depts, httputil.Paginate(page, perPage, int(total)))
}

// POST /organization/departments
func (h *Handler) CreateDepartment(c *fiber.Ctx) error {
	var req CreateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	d := &Department{
		TenantID:          tenantFromCtx(c),
		Code:              req.Code,
		Name:              req.Name,
		ParentID:          req.ParentID,
		ManagerEmployeeID: req.ManagerEmployeeID,
	}
	if err := h.svc.CreateDepartment(c.UserContext(), d); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Department created", d)
}

// GET /organization/departments/:id
func (h *Handler) GetDepartment(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid department ID")
	}
	d, err := h.svc.GetDepartment(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get department")
	}
	if d == nil {
		return httputil.NotFound(c, "Department not found")
	}
	return httputil.Success(c, "Department retrieved", d)
}

// PUT /organization/departments/:id
func (h *Handler) UpdateDepartment(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid department ID")
	}
	var req UpdateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	d, err := h.svc.GetDepartment(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || d == nil {
		return httputil.NotFound(c, "Department not found")
	}

	if req.Name != "" {
		d.Name = req.Name
	}
	if req.ParentID != nil {
		d.ParentID = req.ParentID
	}
	if req.ManagerEmployeeID != nil {
		d.ManagerEmployeeID = req.ManagerEmployeeID
	}
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateDepartment(c.UserContext(), d); err != nil {
		return httputil.InternalServerError(c, "Failed to update department")
	}
	return httputil.Success(c, "Department updated", d)
}

// DELETE /organization/departments/:id
func (h *Handler) DeleteDepartment(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid department ID")
	}
	if err := h.svc.DeleteDepartment(c.UserContext(), uint(id), tenantFromCtx(c)); err != nil {
		return httputil.NotFound(c, err.Error())
	}
	return httputil.Success(c, "Department deactivated", nil)
}

// GET /organization/fiscal-years
func (h *Handler) ListFiscalYears(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountFiscalYears(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count fiscal years")
	}
	fys, err := h.svc.ListFiscalYears(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list fiscal years")
	}
	return httputil.SuccessWithMeta(c, "Fiscal years retrieved", fys, httputil.Paginate(page, perPage, int(total)))
}

// POST /organization/fiscal-years
func (h *Handler) CreateFiscalYear(c *fiber.Ctx) error {
	var req CreateFiscalYearRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	fy := &FiscalYear{
		TenantID:  tenantFromCtx(c),
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	if err := h.svc.CreateFiscalYear(c.UserContext(), fy); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "Fiscal year created", fy)
}

// PUT /organization/fiscal-years/:id/close
func (h *Handler) CloseFiscalYear(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid fiscal year ID")
	}
	if err := h.svc.CloseFiscalYear(c.UserContext(), uint(id), tenantFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "Fiscal year closed", nil)
}

// GET /organization/accounting-periods
func (h *Handler) ListAccountingPeriods(c *fiber.Ctx) error {
	var fyID *uint
	if raw := c.Query("fiscal_year_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			u := uint(v)
			fyID = &u
		}
	}
	periods, err := h.svc.ListAccountingPeriods(c.UserContext(), tenantFromCtx(c), fyID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list accounting periods")
	}
	return httputil.Success(c, "Accounting periods retrieved", periods)
}

// POST /organization/accounting-periods/:id/close
func (h *Handler) CloseAccountingPeriod(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid accounting period ID")
	}
	if err := h.svc.CloseAccountingPeriod(c.UserContext(), uint(id), tenantFromCtx(c)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "Accounting period closed", nil)
}

// GET /organization/document-sequences
func (h *Handler) ListDocumentSequences(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountDocumentSequences(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to count document sequences")
	}
	seqs, err := h.svc.ListDocumentSequences(c.UserContext(), tenantID, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list document sequences")
	}
	return httputil.SuccessWithMeta(c, "Document sequences retrieved", seqs, httputil.Paginate(page, perPage, int(total)))
}

// POST /organization/document-sequences
func (h *Handler) CreateDocumentSequence(c *fiber.Ctx) error {
	var req CreateDocumentSequenceRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	padding := req.Padding
	if padding == 0 {
		padding = 5
	}
	nextNumber := req.NextNumber
	if nextNumber == 0 {
		nextNumber = 1
	}

	ds := &DocumentSequence{
		TenantID:     tenantFromCtx(c),
		DocumentType: req.DocumentType,
		Prefix:       req.Prefix,
		NextNumber:   nextNumber,
		Padding:      padding,
		Suffix:       req.Suffix,
	}
	if err := h.svc.CreateDocumentSequence(c.UserContext(), ds); err != nil {
		return httputil.Conflict(c, err.Error())
	}
	return httputil.Created(c, "Document sequence created", ds)
}

// PUT /organization/document-sequences/:id
func (h *Handler) UpdateDocumentSequence(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid sequence ID")
	}
	var req UpdateDocumentSequenceRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	ds, err := h.svc.GetDocumentSequence(c.UserContext(), uint(id), tenantFromCtx(c))
	if err != nil || ds == nil {
		return httputil.NotFound(c, "Document sequence not found")
	}

	if req.Prefix != "" {
		ds.Prefix = req.Prefix
	}
	if req.NextNumber > 0 {
		ds.NextNumber = req.NextNumber
	}
	if req.Padding > 0 {
		ds.Padding = req.Padding
	}
	ds.Suffix = req.Suffix

	if err := h.svc.UpdateDocumentSequence(c.UserContext(), ds); err != nil {
		return httputil.InternalServerError(c, "Failed to update document sequence")
	}
	return httputil.Success(c, "Document sequence updated", ds)
}

// GET /organization/countries
func (h *Handler) ListCountries(c *fiber.Ctx) error {
	countries, err := h.svc.ListCountries(c.UserContext())
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list countries")
	}
	return httputil.Success(c, "Countries retrieved", countries)
}

// GET /organization/countries/:id/states
func (h *Handler) ListStates(c *fiber.Ctx) error {
	countryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid country ID")
	}
	states, err := h.svc.ListStates(c.UserContext(), uint(countryID))
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list states")
	}
	return httputil.Success(c, "States retrieved", states)
}
