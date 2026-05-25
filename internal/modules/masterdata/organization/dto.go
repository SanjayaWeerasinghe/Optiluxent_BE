package organization

import "time"

// ── Company ──────────────────────────────────────────────────────────────────

type SaveCompanyRequest struct {
	Name            string `json:"name"             validate:"required,min=2,max=200"`
	LegalName       string `json:"legal_name"       validate:"max=200"`
	TaxRegNumber    string `json:"tax_reg_number"   validate:"max=50"`
	Logo            string `json:"logo"`
	Email           string `json:"email"            validate:"omitempty,email"`
	Phone           string `json:"phone"            validate:"max=50"`
	Website         string `json:"website"          validate:"max=255"`
	AddressLine1    string `json:"address_line1"    validate:"max=255"`
	AddressLine2    string `json:"address_line2"    validate:"max=255"`
	City            string `json:"city"             validate:"max=100"`
	StateID         *uint  `json:"state_id"`
	CountryID       *uint  `json:"country_id"`
	PostalCode      string `json:"postal_code"      validate:"max=20"`
	BaseCurrencyID  *uint  `json:"base_currency_id"`
	FiscalYearStart int    `json:"fiscal_year_start" validate:"min=1,max=12"`
}

// ── Department ───────────────────────────────────────────────────────────────

type CreateDepartmentRequest struct {
	Code              string `json:"code"  validate:"required,min=1,max=20"`
	Name              string `json:"name"  validate:"required,min=1,max=100"`
	ParentID          *uint  `json:"parent_id"`
	ManagerEmployeeID *uint  `json:"manager_employee_id"`
}

type UpdateDepartmentRequest struct {
	Name              string `json:"name"               validate:"omitempty,min=1,max=100"`
	ParentID          *uint  `json:"parent_id"`
	ManagerEmployeeID *uint  `json:"manager_employee_id"`
	IsActive          *bool  `json:"is_active"`
}

// ── Fiscal Year ───────────────────────────────────────────────────────────────

type CreateFiscalYearRequest struct {
	Name      string    `json:"name"       validate:"required,min=2,max=50"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date"   validate:"required"`
}

// ── Document Sequence ─────────────────────────────────────────────────────────

type CreateDocumentSequenceRequest struct {
	DocumentType string `json:"document_type" validate:"required,min=2,max=50"`
	Prefix       string `json:"prefix"        validate:"required,min=1,max=10"`
	NextNumber   int    `json:"next_number"   validate:"omitempty,min=1"`
	Padding      int    `json:"padding"       validate:"omitempty,min=1,max=10"`
	Suffix       string `json:"suffix"        validate:"max=20"`
}

type UpdateDocumentSequenceRequest struct {
	Prefix     string `json:"prefix"      validate:"omitempty,min=1,max=10"`
	NextNumber int    `json:"next_number" validate:"omitempty,min=1"`
	Padding    int    `json:"padding"     validate:"omitempty,min=1,max=10"`
	Suffix     string `json:"suffix"      validate:"max=20"`
}
