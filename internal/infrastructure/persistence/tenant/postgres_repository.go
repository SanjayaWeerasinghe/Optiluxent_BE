package tenant

import (
	"context"
	"errors"

	domain "erp-system/internal/domain/tenant"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, t *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.WithContext(ctx).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func (r *PostgresRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.WithContext(ctx).
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, int64, error) {
	var tenants []*domain.Tenant
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Tenant{}).Where("deleted_at IS NULL")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("name ASC").Limit(limit).Offset(offset).Find(&tenants).Error
	return tenants, total, err
}

func (r *PostgresRepository) Update(ctx context.Context, t *domain.Tenant) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
