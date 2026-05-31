package mrp

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context, tenantID uint, materialType string) ([]MaterialMasterView, error)
	Get(ctx context.Context, tenantID, id uint) (*MaterialMasterView, error)
	Create(ctx context.Context, m *MaterialMaster) error
	Update(ctx context.Context, m *MaterialMaster) error
	Delete(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

const viewSelect = `mrp_material_master.*,
	p.code AS product_code,
	p.name AS product_name`

func (r *dbRepository) List(ctx context.Context, tenantID uint, materialType string) ([]MaterialMasterView, error) {
	var rows []MaterialMasterView
	q := r.db.WithContext(ctx).
		Table("mrp_material_master").
		Select(viewSelect).
		Joins("JOIN products p ON p.id = mrp_material_master.product_id").
		Where("mrp_material_master.tenant_id = ?", tenantID)
	if materialType != "" {
		q = q.Where("mrp_material_master.material_type = ?", materialType)
	}
	err := q.Order("p.code").Find(&rows).Error
	return rows, err
}

func (r *dbRepository) Get(ctx context.Context, tenantID, id uint) (*MaterialMasterView, error) {
	var m MaterialMasterView
	err := r.db.WithContext(ctx).
		Table("mrp_material_master").
		Select(viewSelect).
		Joins("JOIN products p ON p.id = mrp_material_master.product_id").
		Where("mrp_material_master.tenant_id = ? AND mrp_material_master.id = ?", tenantID, id).
		First(&m).Error
	return &m, err
}

func (r *dbRepository) Create(ctx context.Context, m *MaterialMaster) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dbRepository) Update(ctx context.Context, m *MaterialMaster) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *dbRepository) Delete(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&MaterialMaster{}).Error
}
