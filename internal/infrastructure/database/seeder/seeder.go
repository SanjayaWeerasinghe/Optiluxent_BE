package seeder

import (
	"context"
	"fmt"

	"erp-system/pkg/logger"

	"gorm.io/gorm"
)

// Seeder is a function that seeds data idempotently.
type Seeder func(ctx context.Context, db *gorm.DB) error

// Run executes a list of seeders in order.
func Run(ctx context.Context, db *gorm.DB, seeders ...Seeder) error {
	for _, s := range seeders {
		name := fmt.Sprintf("%T", s)
		logger.Info("seeder: running", logger.String("seeder", name))
		if err := s(ctx, db); err != nil {
			return fmt.Errorf("seeder %s: %w", name, err)
		}
	}
	logger.Info("seeder: all seeders complete")
	return nil
}

// Production runs seeders safe for production (tenant, permissions, roles, admin user).
func Production(ctx context.Context, db *gorm.DB, adminEmail, adminPassword string) error {
	return Run(ctx, db,
		SeedDefaultTenant,
		SeedPermissions,
		SeedRoles,
		func(ctx context.Context, db *gorm.DB) error {
			return SeedAdminUser(ctx, db, adminEmail, adminPassword)
		},
	)
}

// Development runs production seeders + dev-only test data.
func Development(ctx context.Context, db *gorm.DB, adminEmail, adminPassword string) error {
	if err := Production(ctx, db, adminEmail, adminPassword); err != nil {
		return err
	}
	return Run(ctx, db, SeedDevelopmentData)
}
