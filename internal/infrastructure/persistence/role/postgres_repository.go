package role

import (
	"context"
	"errors"

	domain "erp-system/internal/domain/role"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (r *PostgresRepository) GetByName(ctx context.Context, tenantID *uint, name string) (*domain.Role, error) {
	var role domain.Role
	q := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name)
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	} else {
		q = q.Where("tenant_id IS NULL")
	}
	err := q.First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (r *PostgresRepository) List(ctx context.Context, tenantID *uint) ([]*domain.Role, error) {
	var roles []*domain.Role
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if tenantID != nil {
		q = q.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	}
	err := q.Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *PostgresRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.Role{}).
		Where("id = ? AND is_system = FALSE", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *PostgresRepository) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove existing
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}
		// Bulk insert
		type rolePermission struct {
			RoleID       uint
			PermissionID uint
		}
		rows := make([]rolePermission, len(permissionIDs))
		for i, pid := range permissionIDs {
			rows[i] = rolePermission{RoleID: roleID, PermissionID: pid}
		}
		return tx.Table("role_permissions").CreateInBatches(rows, 100).Error
	})
}

func (r *PostgresRepository) RemovePermission(ctx context.Context, roleID, permissionID uint) error {
	return r.db.WithContext(ctx).
		Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

func (r *PostgresRepository) GetPermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error) {
	var perms []*domain.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}
