package seeder

import (
	"context"

	"gorm.io/gorm"
)

func SeedDefaultTenant(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		INSERT INTO tenants (name, slug, plan, status, config)
		VALUES ('Optiluxent', 'optiluxent', 'enterprise', 'active', '{}')
		ON CONFLICT DO NOTHING
	`).Error
}
