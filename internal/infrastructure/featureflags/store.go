package featureflags

import "context"

// Store defines the interface for reading and writing feature flags.
type Store interface {
	IsEnabled(ctx context.Context, tenantID *uint, name string) (bool, error)
	IsEnabledForUser(ctx context.Context, tenantID *uint, userID uint, name string) (bool, error)
	Get(ctx context.Context, tenantID *uint, name string) (*FeatureFlag, error)
	List(ctx context.Context, tenantID *uint) ([]*FeatureFlag, error)
	Set(ctx context.Context, tenantID *uint, name string, enabled bool) error
	Create(ctx context.Context, flag *FeatureFlag) error
	Delete(ctx context.Context, tenantID *uint, name string) error
	Invalidate(ctx context.Context, tenantID *uint, name string) error
}
