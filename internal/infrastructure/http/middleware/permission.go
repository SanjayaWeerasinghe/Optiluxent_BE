package middleware

import (
	"erp-system/internal/domain/rbac"
	httputil "erp-system/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission checks that the authenticated user's role allows resource:action.
// Must be used after the Authenticate middleware (which sets user_role in locals).
func RequirePermission(enforcer rbac.Enforcer, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || role == "" {
			return httputil.Unauthorized(c, "Authentication required")
		}

		allowed, err := enforcer.Enforce(c.Context(), role, resource, action)
		if err != nil {
			return httputil.InternalServerError(c, "Permission check failed")
		}
		if !allowed {
			return httputil.Forbidden(c, "Insufficient permissions")
		}

		return c.Next()
	}
}
