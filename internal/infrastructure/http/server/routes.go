package server

import (
	"time"

	"erp-system/internal/domain/rbac"
	auditinfra "erp-system/internal/infrastructure/audit"
	rediscache "erp-system/internal/infrastructure/cache/redis"
	"erp-system/internal/infrastructure/http/middleware"
	inventory "erp-system/internal/modules/inventory"
	manufacturing "erp-system/internal/modules/manufacturing"
	masterdata "erp-system/internal/modules/masterdata"
	procurement "erp-system/internal/modules/procurement"
	sales "erp-system/internal/modules/sales"
	audithandler "erp-system/internal/interfaces/http/handlers/audit"
	authhandler "erp-system/internal/interfaces/http/handlers/auth"
	ffhandler "erp-system/internal/interfaces/http/handlers/featureflags"
	rolehandler "erp-system/internal/interfaces/http/handlers/role"
	tenanthandler "erp-system/internal/interfaces/http/handlers/tenant"
	userhandler "erp-system/internal/interfaces/http/handlers/user"
	jwtpkg "erp-system/pkg/jwt"
	"erp-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RouteRegistrar is an interface for registering routes.
type RouteRegistrar interface {
	RegisterRoutes(router fiber.Router)
}

// SetupRoutes sets up all application routes.
func (s *Server) SetupRoutes(
	jwtManager *jwtpkg.Manager,
	blacklist *rediscache.TokenBlacklist,
	authHandler *authhandler.Handler,
	enforcer rbac.Enforcer,
	auditLogger *auditinfra.Logger,
	userHandler *userhandler.Handler,
	roleHandler *rolehandler.Handler,
	tenantHandler *tenanthandler.Handler,
	auditHandler *audithandler.Handler,
	ffHandler *ffhandler.Handler,
	masterdataModule *masterdata.Module,
	procurementModule *procurement.Module,
	inventoryModule *inventory.Module,
	salesModule *sales.Module,
	manufacturingModule *manufacturing.Module,
	redisClient *redis.Client,
	isProd bool,
) {
	// Security headers on every response
	s.app.Use(middleware.SecurityHeaders(isProd))

	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "ERP System API",
			"version": s.config.App.Version,
		})
	})

	v1 := s.app.Group("/api/v1")
	s.setupV1Routes(v1, jwtManager, blacklist, authHandler, enforcer, auditLogger,
		userHandler, roleHandler, tenantHandler, auditHandler, ffHandler,
		masterdataModule, procurementModule, inventoryModule, salesModule, manufacturingModule, redisClient)
}

func (s *Server) setupV1Routes(
	router fiber.Router,
	jwtManager *jwtpkg.Manager,
	blacklist *rediscache.TokenBlacklist,
	authHandler *authhandler.Handler,
	enforcer rbac.Enforcer,
	auditLogger *auditinfra.Logger,
	userHandler *userhandler.Handler,
	roleHandler *rolehandler.Handler,
	tenantHandler *tenanthandler.Handler,
	auditHandler *audithandler.Handler,
	ffHandler *ffhandler.Handler,
	masterdataModule *masterdata.Module,
	procurementModule *procurement.Module,
	inventoryModule *inventory.Module,
	salesModule *sales.Module,
	manufacturingModule *manufacturing.Module,
	redisClient *redis.Client,
) {
	rateLimiter := middleware.NewRateLimiter(redisClient)
	authMW := middleware.Authenticate(jwtManager, blacklist)
	tenantMW := middleware.TenantContext()
	auditMW := middleware.AuditLog(auditLogger)

	// ── Auth routes (public + token-protected) ─────────────────────────────────
	auth := router.Group("/auth")
	// Rate limit: 10 login/register attempts per minute per IP
	authRateLimit := rateLimiter.PerIP(10, time.Minute)
	auth.Post("/register", authRateLimit, authHandler.Register)
	auth.Post("/login", authRateLimit, authHandler.Login)
	auth.Post("/refresh", rateLimiter.PerIP(20, time.Minute), authHandler.Refresh)
	auth.Post("/logout", authMW, authHandler.Logout)
	auth.Get("/me", authMW, authHandler.Me)

	// ── Authenticated + tenant-scoped base ─────────────────────────────────────
	protected := router.Group("", authMW, tenantMW)

	// ── User management ────────────────────────────────────────────────────────
	users := protected.Group("/users")
	users.Get("", middleware.RequirePermission(enforcer, "users", "read"), userHandler.List)
	users.Get("/:id", middleware.RequirePermission(enforcer, "users", "read"), userHandler.Get)
	users.Put("/:id", middleware.RequirePermission(enforcer, "users", "update"), auditMW, userHandler.Update)
	users.Delete("/:id", middleware.RequirePermission(enforcer, "users", "delete"), auditMW, userHandler.Delete)

	// Own profile / password (any authenticated user)
	protected.Put("/profile", auditMW, userHandler.UpdateProfile)
	protected.Put("/password", auditMW, userHandler.ChangePassword)

	// ── Role & permission management ───────────────────────────────────────────
	roles := protected.Group("/roles")
	roles.Get("", middleware.RequirePermission(enforcer, "roles", "read"), roleHandler.List)
	roles.Post("", middleware.RequirePermission(enforcer, "roles", "create"), auditMW, roleHandler.Create)
	roles.Get("/:id", middleware.RequirePermission(enforcer, "roles", "read"), roleHandler.Get)
	roles.Put("/:id", middleware.RequirePermission(enforcer, "roles", "update"), auditMW, roleHandler.Update)
	roles.Delete("/:id", middleware.RequirePermission(enforcer, "roles", "delete"), auditMW, roleHandler.Delete)
	roles.Post("/:id/permissions", middleware.RequirePermission(enforcer, "roles", "update"), auditMW, roleHandler.AssignPermissions)
	roles.Delete("/:id/permissions/:perm_id", middleware.RequirePermission(enforcer, "roles", "update"), auditMW, roleHandler.RemovePermission)

	perms := protected.Group("/permissions")
	perms.Get("", middleware.RequirePermission(enforcer, "permissions", "read"), roleHandler.ListPermissions)

	// ── Tenant management (super_admin only — enforced via Casbin policy) ───────
	tenants := protected.Group("/tenants")
	tenants.Post("", middleware.RequirePermission(enforcer, "tenants", "create"), auditMW, tenantHandler.Create)
	tenants.Get("", middleware.RequirePermission(enforcer, "tenants", "read"), tenantHandler.List)
	tenants.Get("/:id", middleware.RequirePermission(enforcer, "tenants", "read"), tenantHandler.Get)
	tenants.Put("/:id", middleware.RequirePermission(enforcer, "tenants", "update"), auditMW, tenantHandler.Update)

	// ── Feature flags ──────────────────────────────────────────────────────────
	flags := protected.Group("/feature-flags")
	flags.Get("", middleware.RequirePermission(enforcer, "feature_flags", "read"), ffHandler.List)
	flags.Post("", middleware.RequirePermission(enforcer, "feature_flags", "manage"), auditMW, ffHandler.Create)
	flags.Get("/:name", middleware.RequirePermission(enforcer, "feature_flags", "read"), ffHandler.Get)
	flags.Put("/:name/enabled", middleware.RequirePermission(enforcer, "feature_flags", "manage"), auditMW, ffHandler.SetEnabled)
	flags.Delete("/:name", middleware.RequirePermission(enforcer, "feature_flags", "manage"), auditMW, ffHandler.Delete)
	flags.Post("/:name/invalidate", middleware.RequirePermission(enforcer, "feature_flags", "manage"), ffHandler.Invalidate)

	// ── Audit logs ─────────────────────────────────────────────────────────────
	protected.Get("/audit-logs", middleware.RequirePermission(enforcer, "audit", "read"), auditHandler.List)

	// ── Master Data module ─────────────────────────────────────────────────────
	masterdataModule.RegisterRoutes(protected.Group("/masterdata"))

	// ── Procurement module ─────────────────────────────────────────────────────
	procurementModule.RegisterRoutes(protected.Group("/procurement"))

	// ── Inventory module ───────────────────────────────────────────────────────
	inventoryModule.RegisterRoutes(protected.Group("/inventory"))

	// ── Sales module ───────────────────────────────────────────────────────────
	salesModule.RegisterRoutes(protected.Group("/sales"))

	// ── Manufacturing module ───────────────────────────────────────────────────
	manufacturingModule.RegisterRoutes(protected.Group("/manufacturing"))
}

// RegisterModuleRoutes registers routes for a dynamically loaded module.
func (s *Server) RegisterModuleRoutes(moduleName string, registrar RouteRegistrar) {
	moduleGroup := s.app.Group("/api/v1/" + moduleName)
	registrar.RegisterRoutes(moduleGroup)
	s.logger.Info("Module routes registered", logger.String("module", moduleName))
}

// GetRouter returns the main router for testing or advanced configuration.
func (s *Server) GetRouter() fiber.Router {
	return s.app
}

// PrintRoutes prints all registered routes.
func (s *Server) PrintRoutes() {
	for _, route := range s.app.GetRoutes() {
		s.logger.Info("Route",
			logger.String("method", route.Method),
			logger.String("path", route.Path))
	}
}
