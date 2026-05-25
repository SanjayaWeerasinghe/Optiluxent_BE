package hr

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	ListJobPositions(ctx context.Context, tenantID uint) ([]JobPosition, error)
	GetJobPosition(ctx context.Context, tenantID, id uint) (*JobPosition, error)
	CreateJobPosition(ctx context.Context, jp *JobPosition) error
	UpdateJobPosition(ctx context.Context, jp *JobPosition) error

	ListEmployees(ctx context.Context, tenantID uint, activeOnly bool, departmentID *uint) ([]Employee, error)
	GetEmployee(ctx context.Context, tenantID, id uint) (*Employee, error)
	CreateEmployee(ctx context.Context, e *Employee) error
	UpdateEmployee(ctx context.Context, e *Employee) error
	DeleteEmployee(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

func (r *dbRepository) ListJobPositions(ctx context.Context, tenantID uint) ([]JobPosition, error) {
	var rows []JobPosition
	return rows, r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name").Find(&rows).Error
}

func (r *dbRepository) GetJobPosition(ctx context.Context, tenantID, id uint) (*JobPosition, error) {
	var jp JobPosition
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&jp).Error
	return &jp, err
}

func (r *dbRepository) CreateJobPosition(ctx context.Context, jp *JobPosition) error {
	return r.db.WithContext(ctx).Create(jp).Error
}

func (r *dbRepository) UpdateJobPosition(ctx context.Context, jp *JobPosition) error {
	return r.db.WithContext(ctx).Save(jp).Error
}

func (r *dbRepository) ListEmployees(ctx context.Context, tenantID uint, activeOnly bool, departmentID *uint) ([]Employee, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if activeOnly {
		q = q.Where("is_active = true")
	}
	if departmentID != nil {
		q = q.Where("department_id = ?", *departmentID)
	}
	var rows []Employee
	return rows, q.Order("last_name, first_name").Find(&rows).Error
}

func (r *dbRepository) GetEmployee(ctx context.Context, tenantID, id uint) (*Employee, error) {
	var e Employee
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&e).Error
	return &e, err
}

func (r *dbRepository) CreateEmployee(ctx context.Context, e *Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *dbRepository) UpdateEmployee(ctx context.Context, e *Employee) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *dbRepository) DeleteEmployee(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Employee{}).Error
}
