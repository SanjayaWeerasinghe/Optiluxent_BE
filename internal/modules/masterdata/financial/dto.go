package financial

import "time"

// ── Currency ──────────────────────────────────────────────────────────────────

type CreateCurrencyRequest struct {
	Code   string `json:"code"   validate:"required,len=3"`
	Name   string `json:"name"   validate:"required,min=2,max=100"`
	Symbol string `json:"symbol" validate:"required,max=10"`
}

type UpdateCurrencyRequest struct {
	Name     string `json:"name"      validate:"omitempty,min=2,max=100"`
	Symbol   string `json:"symbol"    validate:"omitempty,max=10"`
	IsActive *bool  `json:"is_active"`
}

// ── Exchange Rate ─────────────────────────────────────────────────────────────

type CreateExchangeRateRequest struct {
	FromCurrencyID uint      `json:"from_currency_id" validate:"required"`
	ToCurrencyID   uint      `json:"to_currency_id"   validate:"required"`
	Rate           float64   `json:"rate"             validate:"required,gt=0"`
	EffectiveDate  time.Time `json:"effective_date"   validate:"required"`
	Source         string    `json:"source"           validate:"max=30"`
}

// ── Chart of Accounts ─────────────────────────────────────────────────────────

type CreateCoARequest struct {
	Code        string `json:"code"         validate:"required,min=1,max=20"`
	Name        string `json:"name"         validate:"required,min=2,max=200"`
	AccountType string `json:"account_type" validate:"required"`
	ParentID    *uint  `json:"parent_id"`
	CurrencyID  uint   `json:"currency_id"  validate:"required"`
}

type UpdateCoARequest struct {
	Name         string `json:"name"          validate:"omitempty,min=2,max=200"`
	IsControlled *bool  `json:"is_controlled"`
	IsActive     *bool  `json:"is_active"`
}

// ── Cost Center ───────────────────────────────────────────────────────────────

type CreateCostCenterRequest struct {
	Code         string `json:"code"          validate:"required,min=1,max=20"`
	Name         string `json:"name"          validate:"required,min=2,max=100"`
	DepartmentID *uint  `json:"department_id"`
}

type UpdateCostCenterRequest struct {
	Name         string `json:"name"          validate:"omitempty,min=2,max=100"`
	DepartmentID *uint  `json:"department_id"`
	IsActive     *bool  `json:"is_active"`
}

// ── Payment Term ──────────────────────────────────────────────────────────────

type CreatePaymentTermRequest struct {
	Code            string  `json:"code"             validate:"required,min=1,max=20"`
	Name            string  `json:"name"             validate:"required,min=2,max=100"`
	DueDays         int     `json:"due_days"         validate:"min=0"`
	DiscountDays    int     `json:"discount_days"    validate:"min=0"`
	DiscountPercent float64 `json:"discount_percent" validate:"min=0,max=100"`
}

type UpdatePaymentTermRequest struct {
	Name            string  `json:"name"             validate:"omitempty,min=2,max=100"`
	DueDays         int     `json:"due_days"         validate:"min=0"`
	DiscountDays    int     `json:"discount_days"    validate:"min=0"`
	DiscountPercent float64 `json:"discount_percent" validate:"min=0,max=100"`
	IsActive        *bool   `json:"is_active"`
}

// ── Bank ──────────────────────────────────────────────────────────────────────

type CreateBankRequest struct {
	Name       string `json:"name"        validate:"required,min=2,max=200"`
	BranchName string `json:"branch_name" validate:"max=200"`
	SwiftCode  string `json:"swift_code"  validate:"max=11"`
	Address    string `json:"address"`
}

// ── Company Bank Account ──────────────────────────────────────────────────────

type CreateBankAccountRequest struct {
	BankID        uint   `json:"bank_id"        validate:"required"`
	AccountNumber string `json:"account_number" validate:"required,min=1,max=50"`
	AccountName   string `json:"account_name"   validate:"required,min=2,max=200"`
	CurrencyID    uint   `json:"currency_id"    validate:"required"`
	GLAccountID   uint   `json:"gl_account_id"  validate:"required"`
	IsDefault     bool   `json:"is_default"`
}

type UpdateBankAccountRequest struct {
	AccountName string `json:"account_name" validate:"omitempty,min=2,max=200"`
	IsDefault   *bool  `json:"is_default"`
	IsActive    *bool  `json:"is_active"`
}

// ── Tax Code ──────────────────────────────────────────────────────────────────

type CreateTaxCodeRequest struct {
	Code        string  `json:"code"          validate:"required,min=1,max=20"`
	Name        string  `json:"name"          validate:"required,min=2,max=100"`
	TaxType     string  `json:"tax_type"      validate:"required"`
	Rate        float64 `json:"rate"          validate:"min=0"`
	GLAccountID uint    `json:"gl_account_id" validate:"required"`
}

type UpdateTaxCodeRequest struct {
	Name     string  `json:"name"      validate:"omitempty,min=2,max=100"`
	Rate     float64 `json:"rate"      validate:"min=0"`
	IsActive *bool   `json:"is_active"`
}

// ── Tax Group ─────────────────────────────────────────────────────────────────

type CreateTaxGroupRequest struct {
	Name       string `json:"name"        validate:"required,min=2,max=100"`
	TaxCodeIDs []uint `json:"tax_code_ids"`
}

type UpdateTaxGroupRequest struct {
	Name       string `json:"name"        validate:"omitempty,min=2,max=100"`
	TaxCodeIDs []uint `json:"tax_code_ids"`
	IsActive   *bool  `json:"is_active"`
}
