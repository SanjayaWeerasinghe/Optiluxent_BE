package main

import (
	"context"
	"fmt"
	"os"
	"time"

	auditinfra "erp-system/internal/infrastructure/audit"
	rediscache "erp-system/internal/infrastructure/cache/redis"
	"erp-system/internal/infrastructure/config"
	chdb "erp-system/internal/infrastructure/database/clickhouse"
	"erp-system/internal/infrastructure/database/postgres"
	"erp-system/internal/infrastructure/events"
	featureflags "erp-system/internal/infrastructure/featureflags"
	"erp-system/internal/infrastructure/http/handlers"
	"erp-system/internal/infrastructure/http/middleware"
	"erp-system/internal/infrastructure/http/server"
	auditrepo "erp-system/internal/infrastructure/persistence/audit"
	permrepo "erp-system/internal/infrastructure/persistence/permission"
	rolerepo "erp-system/internal/infrastructure/persistence/role"
	tenantrepo "erp-system/internal/infrastructure/persistence/tenant"
	userrepo "erp-system/internal/infrastructure/persistence/user"
	"erp-system/internal/infrastructure/security"
	audithandler "erp-system/internal/interfaces/http/handlers/audit"
	authhandler "erp-system/internal/interfaces/http/handlers/auth"
	ffhandler "erp-system/internal/interfaces/http/handlers/featureflags"
	rolehandler "erp-system/internal/interfaces/http/handlers/role"
	tenanthandler "erp-system/internal/interfaces/http/handlers/tenant"
	userhandler "erp-system/internal/interfaces/http/handlers/user"
	"erp-system/internal/modules"
	inventory "erp-system/internal/modules/inventory"
	manufacturing "erp-system/internal/modules/manufacturing"
	masterdata "erp-system/internal/modules/masterdata"
	procurement "erp-system/internal/modules/procurement"
	sales "erp-system/internal/modules/sales"
	jwtpkg "erp-system/pkg/jwt"
	"erp-system/pkg/logger"

	casbininfra "erp-system/internal/infrastructure/authorization/casbin"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logConfig := logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		OutputPath: cfg.Logging.OutputPath,
		EnableFile: cfg.Logging.EnableFile,
	}
	if err := logger.Initialize(logConfig); err != nil {
		fmt.Printf("Failed to initialise logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting ERP System",
		logger.String("version", cfg.App.Version),
		logger.String("environment", cfg.App.Environment),
	)

	// ── PostgreSQL ──────────────────────────────────────────────────────────────
	logger.Info("Connecting to PostgreSQL...")
	db, err := postgres.Connect(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Err(err))
	}
	defer postgres.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := postgres.HealthCheck(ctx); err != nil {
		logger.Fatal("Database health check failed", logger.Err(err))
	}
	logger.Info("Database health check passed")

	// ── ClickHouse ──────────────────────────────────────────────────────────────
	logger.Info("Connecting to ClickHouse...")
	ledgerDB, err := chdb.Connect(cfg)
	if err != nil {
		if cfg.App.Environment == "development" {
			logger.Warn("ClickHouse unavailable — ledger writes disabled (dev mode)", logger.Err(err))
		} else {
			logger.Fatal("Failed to connect to ClickHouse", logger.Err(err))
		}
	}
	if ledgerDB != nil {
		defer chdb.Close()
		chCtx, chCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer chCancel()
		if err := chdb.Migrate(chCtx); err != nil {
			logger.Fatal("ClickHouse schema migration failed", logger.Err(err))
		}
		logger.Info("ClickHouse ledger schema ready")
	}

	// ── Redis ───────────────────────────────────────────────────────────────────
	logger.Info("Connecting to Redis...")
	redisClient, err := rediscache.Connect(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.Err(err))
	}
	defer rediscache.Close()

	if err := rediscache.HealthCheck(ctx); err != nil {
		logger.Fatal("Redis health check failed", logger.Err(err))
	}
	logger.Info("Redis health check passed")

	// ── Auth components ─────────────────────────────────────────────────────────
	jwtManager := jwtpkg.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.Issuer,
	)
	blacklist := rediscache.NewTokenBlacklist(redisClient)
	userRepository := userrepo.NewPostgresRepository(db)
	lockout := security.NewAccountLockout(redisClient)
	authHandler := authhandler.NewHandler(
		userRepository,
		jwtManager,
		blacklist,
		lockout,
		cfg.JWT.AccessTokenExpiry,
	)
	logger.Info("Auth components initialised")

	// ── RBAC / Casbin enforcer ──────────────────────────────────────────────────
	enforcer, err := casbininfra.New(db)
	if err != nil {
		logger.Fatal("Failed to initialise Casbin enforcer", logger.Err(err))
	}
	logger.Info("Casbin RBAC enforcer initialised")

	// ── Repositories ────────────────────────────────────────────────────────────
	roleRepository := rolerepo.NewPostgresRepository(db)
	permRepository := permrepo.NewPostgresRepository(db)
	tenantRepository := tenantrepo.NewPostgresRepository(db)
	auditRepository := auditrepo.NewClickHouseRepository(ledgerDB)

	// ── Audit logger ─────────────────────────────────────────────────────────────
	auditLogger := auditinfra.NewLogger(auditRepository)

	// ── Event bus ───────────────────────────────────────────────────────────────
	eventBus := events.NewRedisStreamBus(redisClient)
	eventCtx, eventCancel := context.WithCancel(context.Background())
	defer eventCancel()
	if err := eventBus.Start(eventCtx); err != nil {
		logger.Fatal("Failed to start event bus", logger.Err(err))
	}
	logger.Info("Redis Streams event bus started")

	// ── Feature flags ────────────────────────────────────────────────────────────
	ffStore := featureflags.NewDBStore(db, redisClient)

	// ── HTTP handlers ────────────────────────────────────────────────────────────
	uHandler := userhandler.NewHandler(userRepository)
	rHandler := rolehandler.NewHandler(roleRepository, permRepository)
	tHandler := tenanthandler.NewHandler(tenantRepository)
	aHandler := audithandler.NewHandler(auditRepository)
	fHandler := ffhandler.NewHandler(ffStore)

	// ── Module registry ──────────────────────────────────────────────────────────
	registry := modules.NewRegistry()

	mdModule := masterdata.New(enforcer, auditLogger)
	registry.Register(mdModule)

	procModule := procurement.New(enforcer, auditLogger)
	registry.Register(procModule)

	invModule := inventory.New(enforcer, auditLogger)
	registry.Register(invModule)

	salesModule := sales.New(enforcer, auditLogger)
	registry.Register(salesModule)

	mfgModule := manufacturing.New(enforcer, auditLogger)
	registry.Register(mfgModule)

	if err := registry.Initialize(modules.Dependencies{DB: db, LedgerDB: ledgerDB, EventBus: eventBus}); err != nil {
		logger.Fatal("Failed to initialise modules", logger.Err(err))
	}
	logger.Info("Module registry initialised")

	logger.Info("All components initialised")

	// ── HTTP server ─────────────────────────────────────────────────────────────
	srv := server.New(cfg, logger.Get())
	app := srv.App()

	app.Use(middleware.RequestID())
	if cfg.Server.EnableCORS {
		app.Use(middleware.CORS(cfg))
	}
	app.Use(middleware.Logger(logger.Get()))

	healthHandler := handlers.NewHealthHandler(logger.Get())
	app.Group("/health").Use(func(c *fiber.Ctx) error { return c.Next() })
	healthHandler.RegisterRoutes(app.Group("/health"))

	metricsHandler := handlers.NewMetricsHandler()
	metricsHandler.RegisterRoutes(app.Group("/metrics"))

	// Prometheus HTTP request metrics
	app.Use(middleware.PrometheusMetrics(metricsHandler.GetRegistry()))

	isProd := cfg.App.Environment == "production"
	srv.SetupRoutes(
		jwtManager, blacklist, authHandler,
		enforcer, auditLogger,
		uHandler, rHandler, tHandler, aHandler, fHandler,
		mdModule, procModule, invModule, salesModule, mfgModule,
		redisClient, isProd,
	)

	logger.Info("HTTP server configured",
		logger.String("address", srv.GetAddress()),
	)

	if err := srv.StartWithGracefulShutdown(); err != nil {
		logger.Fatal("Server error", logger.Err(err))
	}

	logger.Info("ERP System shutdown complete")
}
