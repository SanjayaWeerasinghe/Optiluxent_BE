package seeder

import (
	"context"
	"strings"

	"erp-system/internal/domain/rbac"

	"gorm.io/gorm"
)

func SeedPermissions(ctx context.Context, db *gorm.DB) error {
	for _, key := range rbac.AllPermissions {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			continue
		}
		resource, action := parts[0], parts[1]
		if err := db.WithContext(ctx).Exec(`
			INSERT INTO permissions (resource, action, description)
			VALUES (?, ?, ?)
			ON CONFLICT (resource, action) DO NOTHING
		`, resource, action, resource+":"+action).Error; err != nil {
			return err
		}
	}
	return nil
}
