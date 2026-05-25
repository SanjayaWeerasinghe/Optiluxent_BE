package seeder

import (
	"context"
	"fmt"

	domainuser "erp-system/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAdminUser creates the default super_admin user if it doesn't exist.
// adminEmail and adminPassword are required; no hardcoded defaults.
func SeedAdminUser(ctx context.Context, db *gorm.DB, adminEmail, adminPassword string) error {
	if adminEmail == "" || adminPassword == "" {
		return fmt.Errorf("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD must be set")
	}

	// Get the default tenant ID (Optiluxent)
	var tenantID uint
	if err := db.WithContext(ctx).
		Raw("SELECT id FROM tenants WHERE slug = 'optiluxent' LIMIT 1").
		Scan(&tenantID).Error; err != nil {
		return fmt.Errorf("lookup default tenant: %w", err)
	}

	var existing domainuser.User
	err := db.WithContext(ctx).Where("email = ?", adminEmail).First(&existing).Error
	if err == nil {
		// User already exists — ensure tenant_id is set
		if existing.TenantID == nil && tenantID > 0 {
			if err := db.WithContext(ctx).Model(&existing).
				Update("tenant_id", tenantID).Error; err != nil {
				return fmt.Errorf("update admin tenant_id: %w", err)
			}
		}
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	user := &domainuser.User{
		TenantID:     &tenantID,
		Email:        adminEmail,
		PasswordHash: string(hashed),
		FirstName:    "System",
		LastName:     "Administrator",
		Role:         "super_admin",
		Status:       domainuser.StatusActive,
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	// Assign super_admin role via user_roles
	return db.WithContext(ctx).Exec(`
		INSERT INTO user_roles (user_id, role_id)
		SELECT ?, r.id FROM roles r WHERE r.name = 'super_admin'
		ON CONFLICT DO NOTHING
	`, user.ID).Error
}
