package financial

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository defines persistence operations for the Financial submodule.
type Repository interface {
	// Currencies
	CountCurrencies(ctx context.Context) (int64, error)
	ListCurrencies(ctx context.Context, limit, offset int) ([]*Currency, error)
	GetCurrency(ctx context.Context, id uint) (*Currency, error)
	CreateCurrency(ctx context.Context, c *Currency) error
	UpdateCurrency(ctx context.Context, c *Currency) error
	SetBaseCurrency(ctx context.Context, id uint) error

	// Exchange Rates
	CountExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time) (int64, error)
	ListExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time, limit, offset int) ([]*ExchangeRate, error)
	GetLatestRates(ctx context.Context, tenantID uint) ([]*ExchangeRate, error)
	CreateExchangeRate(ctx context.Context, r *ExchangeRate) error

	// Chart of Accounts
	CountCoA(ctx context.Context, tenantID uint) (int64, error)
	ListCoA(ctx context.Context, tenantID uint, limit, offset int) ([]*ChartOfAccount, error)
	GetCoA(ctx context.Context, id, tenantID uint) (*ChartOfAccount, error)
	CreateCoA(ctx context.Context, a *ChartOfAccount) error
	UpdateCoA(ctx context.Context, a *ChartOfAccount) error
	DeleteCoA(ctx context.Context, id, tenantID uint) error

	// Cost Centers
	CountCostCenters(ctx context.Context, tenantID uint) (int64, error)
	ListCostCenters(ctx context.Context, tenantID uint, limit, offset int) ([]*CostCenter, error)
	GetCostCenter(ctx context.Context, id, tenantID uint) (*CostCenter, error)
	CreateCostCenter(ctx context.Context, cc *CostCenter) error
	UpdateCostCenter(ctx context.Context, cc *CostCenter) error

	// Payment Terms
	CountPaymentTerms(ctx context.Context, tenantID uint) (int64, error)
	ListPaymentTerms(ctx context.Context, tenantID uint, limit, offset int) ([]*PaymentTerm, error)
	GetPaymentTerm(ctx context.Context, id, tenantID uint) (*PaymentTerm, error)
	CreatePaymentTerm(ctx context.Context, pt *PaymentTerm) error
	UpdatePaymentTerm(ctx context.Context, pt *PaymentTerm) error

	// Banks
	CountBanks(ctx context.Context) (int64, error)
	ListBanks(ctx context.Context, limit, offset int) ([]*Bank, error)
	CreateBank(ctx context.Context, b *Bank) error

	// Company Bank Accounts
	CountBankAccounts(ctx context.Context, tenantID uint) (int64, error)
	ListBankAccounts(ctx context.Context, tenantID uint, limit, offset int) ([]*CompanyBankAccount, error)
	GetBankAccount(ctx context.Context, id, tenantID uint) (*CompanyBankAccount, error)
	CreateBankAccount(ctx context.Context, ba *CompanyBankAccount) error
	UpdateBankAccount(ctx context.Context, ba *CompanyBankAccount) error

	// Tax Codes
	CountTaxCodes(ctx context.Context, tenantID uint) (int64, error)
	ListTaxCodes(ctx context.Context, tenantID uint, limit, offset int) ([]*TaxCode, error)
	GetTaxCode(ctx context.Context, id, tenantID uint) (*TaxCode, error)
	CreateTaxCode(ctx context.Context, t *TaxCode) error
	UpdateTaxCode(ctx context.Context, t *TaxCode) error

	// Tax Groups
	CountTaxGroups(ctx context.Context, tenantID uint) (int64, error)
	ListTaxGroups(ctx context.Context, tenantID uint, limit, offset int) ([]*TaxGroup, error)
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

func (r *dbRepository) CountCurrencies(ctx context.Context) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&Currency{}).Count(&count).Error
}

func (r *dbRepository) ListCurrencies(ctx context.Context, limit, offset int) ([]*Currency, error) {
	var list []*Currency
	q := r.db.WithContext(ctx)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("code").Find(&list).Error
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

func (r *dbRepository) CountExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time) (int64, error) {
	q := r.db.WithContext(ctx).Model(&ExchangeRate{}).Where("tenant_id = ?", tenantID)
	if fromCurrencyID != nil {
		q = q.Where("from_currency_id = ?", *fromCurrencyID)
	}
	if from != nil {
		q = q.Where("effective_date >= ?", *from)
	}
	if to != nil {
		q = q.Where("effective_date <= ?", *to)
	}
	var count int64
	return count, q.Count(&count).Error
}

func (r *dbRepository) ListExchangeRates(ctx context.Context, tenantID uint, fromCurrencyID *uint, from, to *time.Time, limit, offset int) ([]*ExchangeRate, error) {
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
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
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

func (r *dbRepository) CountCoA(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&ChartOfAccount{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListCoA(ctx context.Context, tenantID uint, limit, offset int) ([]*ChartOfAccount, error) {
	var list []*ChartOfAccount
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("code").Find(&list).Error
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

func (r *dbRepository) CountCostCenters(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&CostCenter{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListCostCenters(ctx context.Context, tenantID uint, limit, offset int) ([]*CostCenter, error) {
	var list []*CostCenter
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("code").Find(&list).Error
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

func (r *dbRepository) CountPaymentTerms(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&PaymentTerm{}).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListPaymentTerms(ctx context.Context, tenantID uint, limit, offset int) ([]*PaymentTerm, error) {
	var list []*PaymentTerm
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND is_active = true", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("code").Find(&list).Error
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

func (r *dbRepository) CountBanks(ctx context.Context) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&Bank{}).Where("is_active = true").Count(&count).Error
}

func (r *dbRepository) ListBanks(ctx context.Context, limit, offset int) ([]*Bank, error) {
	var list []*Bank
	q := r.db.WithContext(ctx).Where("is_active = true")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("name").Find(&list).Error
	return list, err
}

func (r *dbRepository) CreateBank(ctx context.Context, b *Bank) error {
	return r.db.WithContext(ctx).Create(b).Error
}

// ── Company Bank Accounts ─────────────────────────────────────────────────────

func (r *dbRepository) CountBankAccounts(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&CompanyBankAccount{}).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListBankAccounts(ctx context.Context, tenantID uint, limit, offset int) ([]*CompanyBankAccount, error) {
	var list []*CompanyBankAccount
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND is_active = true", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Find(&list).Error
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

func (r *dbRepository) CountTaxCodes(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&TaxCode{}).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListTaxCodes(ctx context.Context, tenantID uint, limit, offset int) ([]*TaxCode, error) {
	var list []*TaxCode
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND is_active = true", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("code").Find(&list).Error
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

func (r *dbRepository) CountTaxGroups(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&TaxGroup{}).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListTaxGroups(ctx context.Context, tenantID uint, limit, offset int) ([]*TaxGroup, error) {
	var list []*TaxGroup
	q := r.db.WithContext(ctx).Preload("TaxCodes").Where("tenant_id = ? AND is_active = true", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("name").Find(&list).Error
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
