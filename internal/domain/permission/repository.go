package permission

import "context"

type Repository interface {
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id uint) (*Permission, error)
	GetByKey(ctx context.Context, resource, action string) (*Permission, error)
	List(ctx context.Context) ([]*Permission, error)
	Delete(ctx context.Context, id uint) error
}
