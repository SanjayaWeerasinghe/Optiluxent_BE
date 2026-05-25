package permission

import (
	"context"
	"errors"

	domain "erp-system/internal/domain/permission"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, p *domain.Permission) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var p domain.Permission
	err := r.db.WithContext(ctx).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *PostgresRepository) GetByKey(ctx context.Context, resource, action string) (*domain.Permission, error) {
	var p domain.Permission
	err := r.db.WithContext(ctx).
		Where("resource = ? AND action = ?", resource, action).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *PostgresRepository) List(ctx context.Context) ([]*domain.Permission, error) {
	var perms []*domain.Permission
	err := r.db.WithContext(ctx).Order("resource ASC, action ASC").Find(&perms).Error
	return perms, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Permission{}, id).Error
}
