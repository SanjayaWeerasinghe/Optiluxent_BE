package financial

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository defines persistence operations for the Financial submodule.
type Repository interface {
	// Currencies
	ListCurrencies(ctx context.Context) ([]*Currency, error)
	GetCurrency(ctx context.Context, id uint) (*Currency, error)
	CreateCurrency(ctx context.Context, c *Currency) error
	UpdateCurrency(ctx context.Context, c *Currency) error
	SetBaseCurrency(ctx context.Context, id uint) error

	// Exchange Rates
	ListExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time) ([]*ExchangeRate, error)
	GetLatestRates(ctx context.Context, tenantID uint) ([]*ExchangeRate, error)
	CreateExchangeRate(ctx context.Context, r *ExchangeRate) error

	// Chart of Accounts
	ListCoA(ctx context.Context, tenantID uint) ([]*ChartOfAccount, error)
	GetCoA(ctx context.Context, id, tenantID uint) (*ChartOfAccount, error)
	CreateCoA(ctx context.Context, a *ChartOfAccount) error
	UpdateCoA(ctx context.Context, a *ChartOfAccount) error
	DeleteCoA(ctx context.Context, id, tenantID uint) error

	// Cost Centers
	ListCostCenters(ctx context.Context, tenantID uint) ([]*CostCenter, error)
	GetCostCenter(ctx context.Context, id, tenantID uint) (*CostCenter, error)
	CreateCostCenter(ctx context.Context, cc *CostCenter) error
	UpdateCostCenter(ctx context.Context, cc *CostCenter) error

	// Payment Terms
	ListPaymentTerms(ctx context.Context, tenantID uint) ([]*PaymentTerm, error)
	GetPaymentTerm(ctx context.Context, id, tenantID uint) (*PaymentTerm, error)
	CreatePaymentTerm(ctx context.Context, pt *PaymentTerm) error
	UpdatePaymentTerm(ctx context.Context, pt *PaymentTerm) error

	// Banks
	ListBanks(ctx context.Context) ([]*Bank, error)
	CreateBank(ctx context.Context, b *Bank) error

	// Company Bank Accounts
	ListBankAccounts(ctx context.Context, tenantID uint) ([]*CompanyBankAccount, error)
	GetBankAccount(ctx context.Context, id, tenantID uint) (*CompanyBankAccount, error)
	CreateBankAccount(ctx context.Context, ba *CompanyBankAccount) error
	UpdateBankAccount(ctx context.Context, ba *CompanyBankAccount) error

	// Tax Codes
	ListTaxCodes(ctx context.Context, tenantID uint) ([]*TaxCode, error)
	GetTaxCode(ctx context.Context, id, tenantID uint) (*TaxCode, error)
	CreateTaxCode(ctx context.Context, t *TaxCode) error
	UpdateTaxCode(ctx context.Context, t *TaxCode) error

	// Tax Groups
	ListTaxGroups(ctx context.Context, tenantID uint) ([]*TaxGroup, error)
	GetTaxGroup(ctx context.Context, id, tenantID uint) (*TaxGroup, error)
	CreateTaxGroup(ctx context.Context, tg *TaxGroup) error
	UpdateTaxGroup(ctx context.Context, tg *TaxGroup) error
}

type dbRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &dbRepository{db: db}
}

// ── Currencies ────────────────────────────────────────────────────────────────

func (r *dbRepository) ListCurrencies(ctx context.Context) ([]*Currency, error) {
	var list []*Currency
	err := r.db.WithContext(ctx).Order("code").Find(&list).Error
	return list, err
}

func (r *dbRepository) GetCurrency(ctx context.Context, id uint) (*Currency, error) {
	var c Currency
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &c, err
}

func (r *dbRepository) CreateCurrency(ctx context.Context, c *Currency) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *dbRepository) UpdateCurrency(ctx context.Context, c *Currency) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *dbRepository) SetBaseCurrency(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Currency{}).Where("is_base = true").Update("is_base", false).Error; err != nil {
			return err
		}
		return tx.Model(&Currency{}).Where("id = ?", id).Update("is_base", true).Error
	})
}

// ── Exchange Rates ────────────────────────────────────────────────────────────

func (r *dbRepository) ListExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time) ([]*ExchangeRate, error) {
	var list []*ExchangeRate
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if fromCurrencyID != nil {
		q = q.Where("from_currency_id = ?", *fromCurrencyID)
	}
	if from != nil {
		q = q.Where("effective_date >= ?", *from)
	}
	if to != nil {
		q = q.Where("effective_date <= ?", *to)
	}
	err := q.Order("effective_date DESC").Find(&list).Error
	return list, err
}

func (r *dbRepository) GetLatestRates(ctx context.Context, tenantID uint) ([]*ExchangeRate, error) {
	var list []*ExchangeRate
	// Subquery: latest effective_date per from_currency
	sub := r.db.WithContext(ctx).
		Model(&ExchangeRate{}).
		Select("from_currency_id, MAX(effective_date) as max_date").
		Where("tenant_id = ?", tenantID).
		Group("from_currency_id")

	err := r.db.WithContext(ctx).
		Joins("JOIN (?) latest ON exchange_rates.from_currency_id = latest.from_currency_id AND exchange_rates.effective_date = latest.max_date", sub).
		Where("exchange_rates.tenant_id = ?", tenantID).
		Find(&list).Error
	return list, err
}

func (r *dbRepository) CreateExchangeRate(ctx context.Context, rate *ExchangeRate) error {
	return r.db.WithContext(ctx).Create(rate).Error
}

// ── Chart of Accounts ─────────────────────────────────────────────────────────

func (r *dbRepository) ListCoA(ctx context.Context, tenantID uint) ([]*ChartOfAccount, error) {
	var list []*ChartOfAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("code").
		Find(&list).Error
	return list, err
}

func (r *dbRepository) GetCoA(ctx context.Context, id, tenantID uint) (*ChartOfAccount, error) {
	var a ChartOfAccount
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &a, err
}

func (r *dbRepository) CreateCoA(ctx context.Context, a *ChartOfAccount) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *dbRepository) UpdateCoA(ctx context.Context, a *ChartOfAccount) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *dbRepository) DeleteCoA(ctx context.Context, id, tenantID uint) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&ChartOfAccount{}).Error
}

// ── Cost Centers ──────────────────────────────────────────────────────────────

func (r *dbRepository) ListCostCenters(ctx context.Context, tenantID uint) ([]*CostCenter, error) {
	var list []*CostCenter
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("code").
		Find(&list).Error
	return list, err
}

func (r *dbRepository) GetCostCenter(ctx context.Context, id, tenantID uint) (*CostCenter, error) {
	var cc CostCenter
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&cc).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &cc, err
}

func (r *dbRepository) CreateCostCenter(ctx context.Context, cc *CostCenter) error {
	return r.db.WithContext(ctx).Create(cc).Error
}

func (r *dbRepository) UpdateCostCenter(ctx context.Context, cc *CostCenter) error {
	return r.db.WithContext(ctx).Save(cc).Error
}

// ── Payment Terms ─────────────────────────────────────────────────────────────

func (r *dbRepository) ListPaymentTerms(ctx context.Context, tenantID uint) ([]*PaymentTerm, error) {
	var list []*PaymentTerm
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Order("code").
		Find(&list).Error
	return list, err
}

func (r *dbRepository) GetPaymentTerm(ctx context.Context, id, tenantID uint) (*PaymentTerm, error) {
	var pt PaymentTerm
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&pt).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &pt, err
}

func (r *dbRepository) CreatePaymentTerm(ctx context.Context, pt *PaymentTerm) error {
	return r.db.WithContext(ctx).Create(pt).Error
}

func (r *dbRepository) UpdatePaymentTerm(ctx context.Context, pt *PaymentTerm) error {
	return r.db.WithContext(ctx).Save(pt).Error
}

// ── Banks ─────────────────────────────────────────────────────────────────────

func (r *dbRepository) ListBanks(ctx context.Context) ([]*Bank, error) {
	var list []*Bank
	err := r.db.WithContext(ctx).
		Where("is_active = true").
		Order("name").
		Find(&list).Error
	return list, err
}

func (r *dbRepository) CreateBank(ctx context.Context, b *Bank) error {
	return r.db.WithContext(ctx).Create(b).Error
}

// ── Company Bank Accounts ─────────────────────────────────────────────────────

func (r *dbRepository) ListBankAccounts(ctx context.Context, tenantID uint) ([]*CompanyBankAccount, error) {
	var list []*CompanyBankAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Find(&list).Error
	return list, err
}

func (r *dbRepository) GetBankAccount(ctx context.Context, id, tenantID uint) (*CompanyBankAccount, error) {
	var ba CompanyBankAccount
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&ba).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &ba, err
}

func (r *dbRepository) CreateBankAccount(ctx context.Context, ba *CompanyBankAccount) error {
	return r.db.WithContext(ctx).Create(ba).Error
}

func (r *dbRepository) UpdateBankAccount(ctx context.Context, ba *CompanyBankAccount) error {
	return r.db.WithContext(ctx).Save(ba).Error
}

// ── Tax Codes ─────────────────────────────────────────────────────────────────

func (r *dbRepository) ListTaxCodes(ctx context.Context, tenantID uint) ([]*TaxCode, error) {
	var list []*TaxCode
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Order("code").
		Find(&list).Error
	return list, err
}

func (r *dbRepository) GetTaxCode(ctx context.Context, id, tenantID uint) (*TaxCode, error) {
	var t TaxCode
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&t).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &t, err
}

func (r *dbRepository) CreateTaxCode(ctx context.Context, t *TaxCode) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *dbRepository) UpdateTaxCode(ctx context.Context, t *TaxCode) error {
	return r.db.WithContext(ctx).Save(t).Error
}

// ── Tax Groups ────────────────────────────────────────────────────────────────

func (r *dbRepository) ListTaxGroups(ctx context.Context, tenantID uint) ([]*TaxGroup, error) {
	var list []*TaxGroup
	err := r.db.WithContext(ctx).
		Preload("TaxCodes").
		Where("tenant_id = ? AND is_active = true", tenantID).
		Order("name").
		Find(&list).Error
	return list, err
}

func (r *dbRepository) GetTaxGroup(ctx context.Context, id, tenantID uint) (*TaxGroup, error) {
	var tg TaxGroup
	err := r.db.WithContext(ctx).
		Preload("TaxCodes").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&tg).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &tg, err
}

func (r *dbRepository) CreateTaxGroup(ctx context.Context, tg *TaxGroup) error {
	return r.db.WithContext(ctx).Create(tg).Error
}

func (r *dbRepository) UpdateTaxGroup(ctx context.Context, tg *TaxGroup) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(tg).Error
}
