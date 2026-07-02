package mmcategories

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Count(ctx context.Context, tenantID uint) (int64, error)
	List(ctx context.Context, tenantID uint, limit, offset int) ([]MaterialCategory, error)
	Get(ctx context.Context, tenantID, id uint) (*MaterialCategory, error)
	Create(ctx context.Context, c *MaterialCategory) error
	Update(ctx context.Context, c *MaterialCategory) error
	Delete(ctx context.Context, tenantID, id uint) error
}

type gormRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

func (r *gormRepo) Count(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&MaterialCategory{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *gormRepo) List(ctx context.Context, tenantID uint, limit, offset int) ([]MaterialCategory, error) {
	var rows []MaterialCategory
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.Order("code").Find(&rows).Error
	return rows, err
}

func (r *gormRepo) Get(ctx context.Context, tenantID, id uint) (*MaterialCategory, error) {
	var c MaterialCategory
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&c).Error
	return &c, err
}

func (r *gormRepo) Create(ctx context.Context, c *MaterialCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *gormRepo) Update(ctx context.Context, c *MaterialCategory) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *gormRepo) Delete(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&MaterialCategory{}).Error
}
