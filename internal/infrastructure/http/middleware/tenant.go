package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const LocalTenantID = "tenant_id"

// TenantContext extracts tenant_id from JWT claims (set by Authenticate middleware).
// Super admins may override the tenant via the X-Tenant-ID header.
func TenantContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// JWT claims already stored tenant_id via Authenticate middleware.
		// If the user is super_admin and provides X-Tenant-ID, use that instead.
		role, _ := c.Locals("user_role").(string)
		if role == "super_admin" {
			if headerTenant := c.Get("X-Tenant-ID"); headerTenant != "" {
				if tid, err := strconv.ParseUint(headerTenant, 10, 64); err == nil {
					c.Locals(LocalTenantID, uint(tid))
					return c.Next()
				}
			}
		}

		// Fall through: tenant_id already set from JWT claims by Authenticate middleware.
		return c.Next()
	}
}
