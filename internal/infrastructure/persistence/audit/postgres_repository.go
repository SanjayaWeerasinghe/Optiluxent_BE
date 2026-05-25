package audit

import (
	"context"
	"errors"

	domain "erp-system/internal/domain/audit"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, log *domain.Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*domain.Log, error) {
	var log domain.Log
	err := r.db.WithContext(ctx).First(&log, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &log, err
}

func (r *PostgresRepository) List(ctx context.Context, f domain.ListFilter) ([]*domain.Log, int64, error) {
	var logs []*domain.Log
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Log{})

	if f.TenantID != nil {
		q = q.Where("tenant_id = ?", *f.TenantID)
	}
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.Resource != "" {
		q = q.Where("resource = ?", f.Resource)
	}
	if f.ResourceID != "" {
		q = q.Where("resource_id = ?", f.ResourceID)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.DateFrom != "" {
		q = q.Where("created_at >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		q = q.Where("created_at <= ?", f.DateTo)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	err := q.Order("created_at DESC").Limit(limit).Offset(f.Offset).Find(&logs).Error
	return logs, total, err
}
