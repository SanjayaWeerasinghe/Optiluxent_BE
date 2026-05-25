package seeder

import (
	"context"

	"gorm.io/gorm"
)

// SeedDevelopmentData seeds extra test data for local development only.
// Never run in production.
func SeedDevelopmentData(ctx context.Context, db *gorm.DB) error {
	// Additional test tenant
	if err := db.WithContext(ctx).Exec(`
		INSERT INTO tenants (name, slug, plan, status, config)
		VALUES ('Test Corp', 'test-corp', 'standard', 'active', '{}')
		ON CONFLICT DO NOTHING
	`).Error; err != nil {
		return err
	}

	// Sample feature flags
	flags := []struct{ name, description string }{
		{"new_dashboard", "Enable the new dashboard UI"},
		{"bulk_import", "Enable bulk product import feature"},
		{"advanced_reports", "Enable advanced reporting module"},
	}
	for _, f := range flags {
		if err := db.WithContext(ctx).Exec(`
			INSERT INTO feature_flags (name, description, enabled)
			VALUES (?, ?, false)
			ON CONFLICT DO NOTHING
		`, f.name, f.description).Error; err != nil {
			return err
		}
	}

	return nil
}
