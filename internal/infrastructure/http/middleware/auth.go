package middleware

import (
	"strings"

	rediscache "erp-system/internal/infrastructure/cache/redis"
	apperrors "erp-system/pkg/errors"
	httputil "erp-system/pkg/http"
	jwtpkg "erp-system/pkg/jwt"
	"erp-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// Authenticate returns a Fiber handler that validates the JWT from the
// Authorization header, checks the token blacklist, and stores user claims
// in the request context locals for downstream handlers.
func Authenticate(jwtManager *jwtpkg.Manager, blacklist *rediscache.TokenBlacklist) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return httputil.Error(c, apperrors.TokenMissing())
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return httputil.Unauthorized(c, "Authorization header must be 'Bearer <token>'")
		}

		token := parts[1]

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			return httputil.Error(c, apperrors.TokenInvalid())
		}

		if claims.Type != string(jwtpkg.AccessToken) {
			return httputil.Unauthorized(c, "Invalid token type")
		}

		ctx := c.UserContext()
		blacklisted, err := blacklist.IsBlacklisted(ctx, token)
		if err != nil {
			logger.Error("Auth middleware: blacklist check failed", logger.Err(err))
			return httputil.InternalServerError(c, "Token validation failed")
		}
		if blacklisted {
			return httputil.Error(c, apperrors.TokenInvalid())
		}

		// Store claims in locals so handlers can read them without re-parsing the token.
		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("token", token)

		return c.Next()
	}
}

// RequireRole returns a middleware that allows only users with one of the given roles.
// Must be placed after Authenticate.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("user_role").(string)
		if _, ok := allowed[role]; !ok {
			return httputil.Error(c, apperrors.InsufficientPermission("access this resource"))
		}
		return c.Next()
	}
}
