package seeder

import (
	"context"
	"strings"

	"erp-system/internal/domain/rbac"

	"gorm.io/gorm"
)

func SeedRoles(ctx context.Context, db *gorm.DB) error {
	roles := []struct {
		name        string
		description string
		isSystem    bool
	}{
		{rbac.RoleSuperAdmin, "Full system access across all tenants", true},
		{rbac.RoleAdmin, "Full access within tenant", true},
		{rbac.RoleManager, "Module management access within tenant", true},
		{rbac.RoleUser, "Standard user access", true},
		{rbac.RoleGuest, "Read-only access", true},
	}

	for _, r := range roles {
		var roleID uint
		if err := db.WithContext(ctx).Raw(`
			INSERT INTO roles (name, description, is_system)
			VALUES (?, ?, ?)
			ON CONFLICT DO NOTHING
			RETURNING id
		`, r.name, r.description, r.isSystem).Scan(&roleID).Error; err != nil {
			return err
		}

		// Fetch role ID if insert was skipped (conflict)
		if roleID == 0 {
			if err := db.WithContext(ctx).Raw(
				`SELECT id FROM roles WHERE name = ? AND deleted_at IS NULL`, r.name,
			).Scan(&roleID).Error; err != nil {
				return err
			}
		}

		perms := rbac.DefaultRolePermissions[r.name]
		if len(perms) == 0 {
			continue
		}

		// If wildcard, assign all permissions
		if perms[0] == "*" {
			if err := db.WithContext(ctx).Exec(`
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT ?, id FROM permissions
				ON CONFLICT DO NOTHING
			`, roleID).Error; err != nil {
				return err
			}
			continue
		}

		for _, key := range perms {
			parts := strings.SplitN(key, ":", 2)
			if len(parts) != 2 {
				continue
			}
			if err := db.WithContext(ctx).Exec(`
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT ?, id FROM permissions WHERE resource = ? AND action = ?
				ON CONFLICT DO NOTHING
			`, roleID, parts[0], parts[1]).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
