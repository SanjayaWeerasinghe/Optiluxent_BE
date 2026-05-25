package audit

import "context"

type ListFilter struct {
	TenantID   *uint
	UserID     *uint
	Resource   string
	ResourceID string
	Action     string
	DateFrom   string
	DateTo     string
	Limit      int
	Offset     int
}

type Repository interface {
	Create(ctx context.Context, log *Log) error
	List(ctx context.Context, filter ListFilter) ([]*Log, int64, error)
	GetByID(ctx context.Context, id uint) (*Log, error)
}
