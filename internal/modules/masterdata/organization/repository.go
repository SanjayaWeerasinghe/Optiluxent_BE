package organization

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines persistence operations for the Organization submodule.
type Repository interface {
	GetCompany(ctx context.Context) (*Company, error)
	SaveCompany(ctx context.Context, c *Company) error

	CountDepartments(ctx context.Context, tenantID uint) (int64, error)
	ListDepartments(ctx context.Context, tenantID uint, limit, offset int) ([]*Department, error)
	GetDepartment(ctx context.Context, id, tenantID uint) (*Department, error)
	CreateDepartment(ctx context.Context, d *Department) error
	UpdateDepartment(ctx context.Context, d *Department) error
	DeleteDepartment(ctx context.Context, id, tenantID uint) error

	CountFiscalYears(ctx context.Context, tenantID uint) (int64, error)
	ListFiscalYears(ctx context.Context, tenantID uint, limit, offset int) ([]*FiscalYear, error)
	GetFiscalYear(ctx context.Context, id, tenantID uint) (*FiscalYear, error)
	CreateFiscalYear(ctx context.Context, fy *FiscalYear) error
	CloseFiscalYear(ctx context.Context, id, tenantID uint) error

	ListAccountingPeriods(ctx context.Context, tenantID uint, fiscalYearID *uint) ([]*AccountingPeriod, error)
	GetAccountingPeriod(ctx context.Context, id, tenantID uint) (*AccountingPeriod, error)
	CloseAccountingPeriod(ctx context.Context, id, tenantID uint) error

	CountDocumentSequences(ctx context.Context, tenantID uint) (int64, error)
	ListDocumentSequences(ctx context.Context, tenantID uint, limit, offset int) ([]*DocumentSequence, error)
	GetDocumentSequence(ctx context.Context, id, tenantID uint) (*DocumentSequence, error)
	CreateDocumentSequence(ctx context.Context, ds *DocumentSequence) error
	UpdateDocumentSequence(ctx context.Context, ds *DocumentSequence) error

	ListCountries(ctx context.Context) ([]*Country, error)
	ListStates(ctx context.Context, countryID uint) ([]*State, error)
}

type dbRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &dbRepository{db: db}
}

func (r *dbRepository) GetCompany(ctx context.Context) (*Company, error) {
	var c Company
	err := r.db.WithContext(ctx).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &c, err
}

func (r *dbRepository) SaveCompany(ctx context.Context, c *Company) error {
	if c.ID == 0 {
		return r.db.WithContext(ctx).Create(c).Error
	}
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *dbRepository) CountDepartments(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&Department{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error
}

func (r *dbRepository) ListDepartments(ctx context.Context, tenantID uint, limit, offset int) ([]*Department, error) {
	var depts []*Department
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("name").Find(&depts).Error
	return depts, err
}

func (r *dbRepository) GetDepartment(ctx context.Context, id, tenantID uint) (*Department, error) {
	var d Department
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &d, err
}

func (r *dbRepository) CreateDepartment(ctx context.Context, d *Department) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *dbRepository) UpdateDepartment(ctx context.Context, d *Department) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *dbRepository) DeleteDepartment(ctx context.Context, id, tenantID uint) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&Department{}).Error
}

func (r *dbRepository) CountFiscalYears(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&FiscalYear{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *dbRepository) ListFiscalYears(ctx context.Context, tenantID uint, limit, offset int) ([]*FiscalYear, error) {
	var fys []*FiscalYear
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("start_date DESC").Find(&fys).Error
	return fys, err
}

func (r *dbRepository) GetFiscalYear(ctx context.Context, id, tenantID uint) (*FiscalYear, error) {
	var fy FiscalYear
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&fy).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &fy, err
}

func (r *dbRepository) CreateFiscalYear(ctx context.Context, fy *FiscalYear) error {
	return r.db.WithContext(ctx).Create(fy).Error
}

func (r *dbRepository) CloseFiscalYear(ctx context.Context, id, tenantID uint) error {
	result := r.db.WithContext(ctx).
		Model(&FiscalYear{}).
		Where("id = ? AND tenant_id = ? AND is_closed = false", id, tenantID).
		Update("is_closed", true)
	return result.Error
}

func (r *dbRepository) ListAccountingPeriods(ctx context.Context, tenantID uint, fiscalYearID *uint) ([]*AccountingPeriod, error) {
	var periods []*AccountingPeriod
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if fiscalYearID != nil {
		q = q.Where("fiscal_year_id = ?", *fiscalYearID)
	}
	err := q.Order("period_number").Find(&periods).Error
	return periods, err
}

func (r *dbRepository) GetAccountingPeriod(ctx context.Context, id, tenantID uint) (*AccountingPeriod, error) {
	var p AccountingPeriod
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (r *dbRepository) CloseAccountingPeriod(ctx context.Context, id, tenantID uint) error {
	result := r.db.WithContext(ctx).
		Model(&AccountingPeriod{}).
		Where("id = ? AND tenant_id = ? AND is_closed = false", id, tenantID).
		Update("is_closed", true)
	return result.Error
}

func (r *dbRepository) CountDocumentSequences(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&DocumentSequence{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *dbRepository) ListDocumentSequences(ctx context.Context, tenantID uint, limit, offset int) ([]*DocumentSequence, error) {
	var seqs []*DocumentSequence
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("document_type").Find(&seqs).Error
	return seqs, err
}

func (r *dbRepository) GetDocumentSequence(ctx context.Context, id, tenantID uint) (*DocumentSequence, error) {
	var s DocumentSequence
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *dbRepository) CreateDocumentSequence(ctx context.Context, ds *DocumentSequence) error {
	return r.db.WithContext(ctx).Create(ds).Error
}

func (r *dbRepository) UpdateDocumentSequence(ctx context.Context, ds *DocumentSequence) error {
	return r.db.WithContext(ctx).Save(ds).Error
}

func (r *dbRepository) ListCountries(ctx context.Context) ([]*Country, error) {
	var countries []*Country
	err := r.db.WithContext(ctx).
		Where("is_active = true").
		Order("name").
		Find(&countries).Error
	return countries, err
}

func (r *dbRepository) ListStates(ctx context.Context, countryID uint) ([]*State, error) {
	var states []*State
	err := r.db.WithContext(ctx).
		Where("country_id = ?", countryID).
		Order("name").
		Find(&states).Error
	return states, err
}
