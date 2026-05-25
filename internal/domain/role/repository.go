package role

import "context"

type Repository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByName(ctx context.Context, tenantID *uint, name string) (*Role, error)
	List(ctx context.Context, tenantID *uint) ([]*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uint) error
	AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	RemovePermission(ctx context.Context, roleID, permissionID uint) error
	GetPermissions(ctx context.Context, roleID uint) ([]*Permission, error)
}
