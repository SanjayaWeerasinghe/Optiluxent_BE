package financial

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Service handles business logic for the Financial submodule.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ── Currencies ────────────────────────────────────────────────────────────────

func (s *Service) ListCurrencies(ctx context.Context) ([]*Currency, error) {
	return s.repo.ListCurrencies(ctx)
}

func (s *Service) GetCurrency(ctx context.Context, id uint) (*Currency, error) {
	return s.repo.GetCurrency(ctx, id)
}

func (s *Service) CreateCurrency(ctx context.Context, c *Currency) error {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	if err := s.repo.CreateCurrency(ctx, c); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("currency code already exists")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateCurrency(ctx context.Context, c *Currency) error {
	return s.repo.UpdateCurrency(ctx, c)
}

func (s *Service) SetBaseCurrency(ctx context.Context, id uint) error {
	existing, err := s.repo.GetCurrency(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("currency not found")
	}
	return s.repo.SetBaseCurrency(ctx, id)
}

// ── Exchange Rates ────────────────────────────────────────────────────────────

func (s *Service) ListExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time) ([]*ExchangeRate, error) {
	return s.repo.ListExchangeRates(ctx, tenantID, fromCurrencyID, from, to)
}

func (s *Service) GetLatestRates(ctx context.Context, tenantID uint) ([]*ExchangeRate, error) {
	return s.repo.GetLatestRates(ctx, tenantID)
}

func (s *Service) CreateExchangeRate(ctx context.Context, r *ExchangeRate) error {
	if r.Rate <= 0 {
		return errors.New("rate must be greater than zero")
	}
	if r.FromCurrencyID == r.ToCurrencyID {
		return errors.New("from_currency and to_currency must differ")
	}
	return s.repo.CreateExchangeRate(ctx, r)
}

// ── Chart of Accounts ─────────────────────────────────────────────────────────

var validAccountTypes = map[string]bool{
	"ASSET": true, "LIABILITY": true, "EQUITY": true, "REVENUE": true, "EXPENSE": true,
}

func (s *Service) ListCoA(ctx context.Context, tenantID uint) ([]*ChartOfAccount, error) {
	return s.repo.ListCoA(ctx, tenantID)
}

func (s *Service) GetCoA(ctx context.Context, id, tenantID uint) (*ChartOfAccount, error) {
	return s.repo.GetCoA(ctx, id, tenantID)
}

func (s *Service) CreateCoA(ctx context.Context, a *ChartOfAccount) error {
	if !validAccountTypes[strings.ToUpper(a.AccountType)] {
		return errors.New("invalid account_type: must be ASSET, LIABILITY, EQUITY, REVENUE, or EXPENSE")
	}
	a.AccountType = strings.ToUpper(a.AccountType)
	if err := s.repo.CreateCoA(ctx, a); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("account code already exists in this tenant")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateCoA(ctx context.Context, a *ChartOfAccount) error {
	return s.repo.UpdateCoA(ctx, a)
}

func (s *Service) DeleteCoA(ctx context.Context, id, tenantID uint) error {
	existing, err := s.repo.GetCoA(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("account not found")
	}
	return s.repo.DeleteCoA(ctx, id, tenantID)
}

// ── Cost Centers ──────────────────────────────────────────────────────────────

func (s *Service) ListCostCenters(ctx context.Context, tenantID uint) ([]*CostCenter, error) {
	return s.repo.ListCostCenters(ctx, tenantID)
}

func (s *Service) GetCostCenter(ctx context.Context, id, tenantID uint) (*CostCenter, error) {
	return s.repo.GetCostCenter(ctx, id, tenantID)
}

func (s *Service) CreateCostCenter(ctx context.Context, cc *CostCenter) error {
	if err := s.repo.CreateCostCenter(ctx, cc); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("cost center code already exists in this tenant")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateCostCenter(ctx context.Context, cc *CostCenter) error {
	return s.repo.UpdateCostCenter(ctx, cc)
}

// ── Payment Terms ─────────────────────────────────────────────────────────────

func (s *Service) ListPaymentTerms(ctx context.Context, tenantID uint) ([]*PaymentTerm, error) {
	return s.repo.ListPaymentTerms(ctx, tenantID)
}

func (s *Service) GetPaymentTerm(ctx context.Context, id, tenantID uint) (*PaymentTerm, error) {
	return s.repo.GetPaymentTerm(ctx, id, tenantID)
}

func (s *Service) CreatePaymentTerm(ctx context.Context, pt *PaymentTerm) error {
	if err := s.repo.CreatePaymentTerm(ctx, pt); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("payment term code already exists in this tenant")
		}
		return err
	}
	return nil
}

func (s *Service) UpdatePaymentTerm(ctx context.Context, pt *PaymentTerm) error {
	return s.repo.UpdatePaymentTerm(ctx, pt)
}

// ── Banks ─────────────────────────────────────────────────────────────────────

func (s *Service) ListBanks(ctx context.Context) ([]*Bank, error) {
	return s.repo.ListBanks(ctx)
}

func (s *Service) CreateBank(ctx context.Context, b *Bank) error {
	return s.repo.CreateBank(ctx, b)
}

// ── Company Bank Accounts ─────────────────────────────────────────────────────

func (s *Service) ListBankAccounts(ctx context.Context, tenantID uint) ([]*CompanyBankAccount, error) {
	return s.repo.ListBankAccounts(ctx, tenantID)
}

func (s *Service) GetBankAccount(ctx context.Context, id, tenantID uint) (*CompanyBankAccount, error) {
	return s.repo.GetBankAccount(ctx, id, tenantID)
}

func (s *Service) CreateBankAccount(ctx context.Context, ba *CompanyBankAccount) error {
	return s.repo.CreateBankAccount(ctx, ba)
}

func (s *Service) UpdateBankAccount(ctx context.Context, ba *CompanyBankAccount) error {
	return s.repo.UpdateBankAccount(ctx, ba)
}

// ── Tax Codes ─────────────────────────────────────────────────────────────────

var validTaxTypes = map[string]bool{
	"VAT": true, "WHT": true, "SVAT": true, "EXEMPT": true,
}

func (s *Service) ListTaxCodes(ctx context.Context, tenantID uint) ([]*TaxCode, error) {
	return s.repo.ListTaxCodes(ctx, tenantID)
}

func (s *Service) GetTaxCode(ctx context.Context, id, tenantID uint) (*TaxCode, error) {
	return s.repo.GetTaxCode(ctx, id, tenantID)
}

func (s *Service) CreateTaxCode(ctx context.Context, t *TaxCode) error {
	if !validTaxTypes[strings.ToUpper(t.TaxType)] {
		return errors.New("invalid tax_type: must be VAT, WHT, SVAT, or EXEMPT")
	}
	t.TaxType = strings.ToUpper(t.TaxType)
	if err := s.repo.CreateTaxCode(ctx, t); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("tax code already exists in this tenant")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateTaxCode(ctx context.Context, t *TaxCode) error {
	return s.repo.UpdateTaxCode(ctx, t)
}

// ── Tax Groups ────────────────────────────────────────────────────────────────

func (s *Service) ListTaxGroups(ctx context.Context, tenantID uint) ([]*TaxGroup, error) {
	return s.repo.ListTaxGroups(ctx, tenantID)
}

func (s *Service) GetTaxGroup(ctx context.Context, id, tenantID uint) (*TaxGroup, error) {
	return s.repo.GetTaxGroup(ctx, id, tenantID)
}

func (s *Service) CreateTaxGroup(ctx context.Context, tg *TaxGroup) error {
	if err := s.repo.CreateTaxGroup(ctx, tg); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("tax group name already exists in this tenant")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateTaxGroup(ctx context.Context, tg *TaxGroup) error {
	return s.repo.UpdateTaxGroup(ctx, tg)
}
