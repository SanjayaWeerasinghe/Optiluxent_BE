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
	finance "erp-system/internal/modules/finance"
	hr "erp-system/internal/modules/hr"
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
	uHandler := userhandler.NewHandler(userRepository, roleRepository, enforcer)
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

	finModule := finance.New(enforcer, auditLogger)
	registry.Register(finModule)

	hrModule := hr.New(enforcer, auditLogger)
	registry.Register(hrModule)

	if err := registry.Initialize(modules.Dependencies{DB: db, LedgerDB: ledgerDB, EventBus: eventBus}); err != nil {
		logger.Fatal("Failed to initialise modules", logger.Err(err))
	}
	logger.Info("Module registry initialised")

	// ── Cross-module wiring ─────────────────────────────────────────────────────
	// Procurement (GRN confirm) and Manufacturing (output add) need to auto-create
	// QualityCheck entries owned by Inventory. Wire adapters that bridge the
	// per-package QCAutoLine types to inventory's CreateAutoQC.
	if invSvc := invModule.Service(); invSvc != nil {
		procModule.Service().SetQCAutoCreator(procQCAdapter{inv: invSvc})
		mfgModule.Service().SetQCAutoCreator(mfgQCAdapter{inv: invSvc})
		logger.Info("QC auto-creator wired into procurement and manufacturing")
	}
	// PRODUCTION_OUTPUT GRNs bump the source Manufacturing Order's produced_qty.
	if mfgSvc := mfgModule.Service(); mfgSvc != nil {
		procModule.Service().SetMOProducedBumper(mfgSvc)
		logger.Info("MO produced-qty bumper wired into procurement")
	}
	// DocumentType resolver — the masterdata documenttypes service satisfies
	// both procurement.DocumentTypeResolver and inventory.DocumentTypeResolver
	// structurally, no adapter needed.
	if dtSvc := mdModule.DocumentTypeService(); dtSvc != nil {
		procModule.Service().SetDocumentTypeResolver(dtSvc)
		invModule.Service().SetDocumentTypeResolver(dtSvc)
		logger.Info("Document-type resolver wired into procurement + inventory")
	}
	// Product-Kind provider — lets stock-hitting flows (GRN item add,
	// MR/GI/GT/SA line add, DO line add) reject SERVICE products.
	if prodSvc := mdModule.ProductService(); prodSvc != nil {
		procModule.Service().SetProductKindProvider(prodSvc)
		logger.Info("Product-kind provider wired into procurement")
	}
	// Damaged-bin resolver — the masterdata/inventory (storage-locations)
	// service satisfies inventory.DamagedBinResolver structurally. Used by
	// SubmitQualityCheck to route qty_failed onto the receiving
	// warehouse's DAMAGED bin (auto-creating one if missing).
	if mdInvSvc := mdModule.InventoryMasterdataService(); mdInvSvc != nil {
		invModule.Service().SetDamagedBinResolver(mdInvSvc)
		logger.Info("Damaged-bin resolver wired into inventory")
	}
	// Finance module wiring — bi-directional:
	//   • finance.RecordPayment needs to call sales.ApplyInvoicePayment /
	//     procurement.ApplyInvoicePayment to bump paid_amount + status.
	//   • sales/procurement PostInvoice + RecordPayment need to trigger
	//     finance.PostSIToGL / PostPIToGL + insert a payments row.
	// Each side is satisfied by thin structural adapters — no reflection.
	if finSvc := finModule.Service(); finSvc != nil {
		finSvc.SetSalesApplier(salesApplierAdapter{sales: salesModule.Service()})
		finSvc.SetProcurementApplier(procApplierAdapter{proc: procModule.Service()})
		finSvc.SetSalesInvoiceLookup(salesInvoiceLookupAdapter{sales: salesModule.Service()})
		finSvc.SetProcurementInvoiceLookup(procInvoiceLookupAdapter{proc: procModule.Service()})
		procModule.Service().SetFinancePoster(financePosterProc{fin: finSvc})
		salesModule.Service().SetFinancePoster(financePosterSales{fin: finSvc})
		salesModule.Service().SetFinanceChecker(finSvc)
		logger.Info("Finance module wired into procurement + sales")
	}
	// Stock Allocation wiring — SO line CRUD needs to reserve/consume via
	// inventory's allocation service. The adapter translates sales' local
	// SOAllocReserveRequest to inventory's ReserveRequest so the two modules
	// stay import-cycle-free.
	if invAlloc := invModule.Service().Allocation(); invAlloc != nil {
		salesModule.Service().SetAllocationReserver(salesAllocAdapter{alloc: invAlloc})
		logger.Info("Stock allocation service wired into sales")
	}
	// MO dashboard readers — one function per source list. Each adapter maps
	// the source module's rich types down to the lightweight dashboard row
	// types manufacturing works with.
	if mfgSvc := mfgModule.Service(); mfgSvc != nil {
		invSvc := invModule.Service()
		procSvc := procModule.Service()
		mfgSvc.SetDashboardReaders(manufacturing.DashboardReaders{
			ListMRsByMO: func(ctx context.Context, tenantID, moID uint) ([]manufacturing.LinkedMR, error) {
				rows, err := invSvc.ListMRsByMO(ctx, tenantID, moID)
				if err != nil {
					return nil, err
				}
				out := make([]manufacturing.LinkedMR, 0, len(rows))
				for _, r := range rows {
					lines := make([]manufacturing.MRLineRow, 0, len(r.Lines))
					for _, l := range r.Lines {
						lines = append(lines, manufacturing.MRLineRow{
							ProductID: l.ProductID, UOMID: l.UOMId,
							RequestedQty: l.RequestedQty, IssuedQty: l.IssuedQty,
						})
					}
					out = append(out, manufacturing.LinkedMR{ID: r.ID, Code: r.Code, Status: r.Status, Lines: lines})
				}
				return out, nil
			},
			ListGIsByMO: func(ctx context.Context, tenantID, moID uint) ([]manufacturing.LinkedGI, error) {
				rows, err := invSvc.ListIssuesByMO(ctx, tenantID, moID)
				if err != nil {
					return nil, err
				}
				out := make([]manufacturing.LinkedGI, 0, len(rows))
				for _, r := range rows {
					lines := make([]manufacturing.StockLineRow, 0, len(r.Lines))
					for _, l := range r.Lines {
						lines = append(lines, manufacturing.StockLineRow{ProductID: l.ProductID, UOMID: l.UOMId, Quantity: l.Quantity})
					}
					out = append(out, manufacturing.LinkedGI{ID: r.ID, Code: r.Code, Status: r.Status, Lines: lines})
				}
				return out, nil
			},
			ListGTsByMO: func(ctx context.Context, tenantID, moID uint) ([]manufacturing.LinkedGT, error) {
				rows, err := invSvc.ListTransfersByMO(ctx, tenantID, moID)
				if err != nil {
					return nil, err
				}
				out := make([]manufacturing.LinkedGT, 0, len(rows))
				for _, r := range rows {
					lines := make([]manufacturing.StockLineRow, 0, len(r.Lines))
					for _, l := range r.Lines {
						lines = append(lines, manufacturing.StockLineRow{ProductID: l.ProductID, UOMID: l.UOMId, Quantity: l.Quantity})
					}
					out = append(out, manufacturing.LinkedGT{ID: r.ID, Code: r.Code, Status: r.Status, Lines: lines})
				}
				return out, nil
			},
			ListGRNsByMO: func(ctx context.Context, tenantID, moID uint) ([]manufacturing.LinkedGRN, error) {
				rows, err := procSvc.ListGRNsByMO(ctx, tenantID, moID)
				if err != nil {
					return nil, err
				}
				out := make([]manufacturing.LinkedGRN, 0, len(rows))
				docTypeSvc := mdModule.DocumentTypeService()
				for _, r := range rows {
					lines := make([]manufacturing.StockLineRow, 0, len(r.Lines))
					for _, l := range r.Lines {
						lines = append(lines, manufacturing.StockLineRow{ProductID: l.ProductID, UOMID: l.UOMID, Quantity: l.Quantity})
					}
					// Resolve the seeded system_key so the dashboard's PRODUCTION_OUTPUT
					// filter keeps working after grn_type was dropped from the schema.
					var grnKey string
					if r.DocumentTypeID != nil && docTypeSvc != nil {
						grnKey, _ = docTypeSvc.ResolveSystemKey(ctx, tenantID, *r.DocumentTypeID)
					}
					out = append(out, manufacturing.LinkedGRN{ID: r.ID, Code: r.Code, Status: r.Status, GRNType: grnKey, Lines: lines})
				}
				return out, nil
			},
			ListQCsForGRNs: func(ctx context.Context, tenantID uint, grnIDs []uint) ([]manufacturing.LinkedQC, error) {
				rows, err := invSvc.ListQCsByGRNIDs(ctx, tenantID, grnIDs)
				if err != nil {
					return nil, err
				}
				docTypeSvc := mdModule.DocumentTypeService()
				out := make([]manufacturing.LinkedQC, 0, len(rows))
				for _, r := range rows {
					var checked, passed, failed float64
					for _, l := range r.Lines {
						checked += l.QtyChecked
						passed  += l.QtyPassed
						failed  += l.QtyFailed
					}
					// Resolve qc_type via the Type's system_key so the dashboard's
					// MATERIAL_QC vs PRODUCT_QC filters keep working.
					var qcKey string
					if r.DocumentTypeID != nil && docTypeSvc != nil {
						qcKey, _ = docTypeSvc.ResolveSystemKey(ctx, tenantID, *r.DocumentTypeID)
					}
					out = append(out, manufacturing.LinkedQC{
						ID: r.ID, Code: r.Code, QCType: qcKey, Status: r.Status,
						QtyChecked: checked, QtyPassed: passed, QtyFailed: failed,
						RefType: r.ReferenceType, RefID: r.ReferenceID,
					})
				}
				return out, nil
			},
		})
		logger.Info("MO dashboard readers wired")
	}

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
		mdModule, procModule, invModule, salesModule, mfgModule, finModule, hrModule,
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

// procQCAdapter satisfies procurement.QCAutoCreator by translating procurement's
// QCAutoLine slice into inventory's QCAutoLine slice and delegating to invSvc.
type procQCAdapter struct{ inv *inventory.Service }

func (a procQCAdapter) CreateAutoQC(
	ctx context.Context,
	tenantID, userID uint,
	qcType, refType string,
	refID, warehouseID uint,
	notes string,
	lines []procurement.QCAutoLine,
) error {
	out := make([]inventory.QCAutoLine, len(lines))
	for i, l := range lines {
		out[i] = inventory.QCAutoLine{ProductID: l.ProductID, VariantID: l.VariantID, Quantity: l.Quantity}
	}
	_, err := a.inv.CreateAutoQC(ctx, tenantID, userID, qcType, refType, refID, warehouseID, notes, out)
	return err
}

// mfgQCAdapter does the same for manufacturing.QCAutoCreator.
type mfgQCAdapter struct{ inv *inventory.Service }

func (a mfgQCAdapter) CreateAutoQC(
	ctx context.Context,
	tenantID, userID uint,
	qcType, refType string,
	refID, warehouseID uint,
	notes string,
	lines []manufacturing.QCAutoLine,
) error {
	out := make([]inventory.QCAutoLine, len(lines))
	for i, l := range lines {
		out[i] = inventory.QCAutoLine{ProductID: l.ProductID, VariantID: l.VariantID, Quantity: l.Quantity}
	}
	_, err := a.inv.CreateAutoQC(ctx, tenantID, userID, qcType, refType, refID, warehouseID, notes, out)
	return err
}

// ── Finance ↔ sales/procurement adapters ────────────────────────────────────
//
// Each side defines its own tiny interface; the adapters translate between
// module-neutral finance signatures and the concrete service methods.

type salesApplierAdapter struct{ sales *sales.Service }

func (a salesApplierAdapter) ApplyPayment(ctx context.Context, tenantID, invoiceID uint, amount float64) (uint, uint, error) {
	return a.sales.ApplyInvoicePayment(ctx, tenantID, invoiceID, amount)
}

type procApplierAdapter struct{ proc *procurement.Service }

func (a procApplierAdapter) ApplyPayment(ctx context.Context, tenantID, invoiceID uint, amount float64) (uint, uint, error) {
	return a.proc.ApplyInvoicePayment(ctx, tenantID, invoiceID, amount)
}

type salesInvoiceLookupAdapter struct{ sales *sales.Service }

func (a salesInvoiceLookupAdapter) LookupInvoice(ctx context.Context, tenantID, invoiceID uint) (uint, uint, float64, float64, error) {
	return a.sales.LookupInvoiceSI(ctx, tenantID, invoiceID)
}

type procInvoiceLookupAdapter struct{ proc *procurement.Service }

func (a procInvoiceLookupAdapter) LookupInvoice(ctx context.Context, tenantID, invoiceID uint) (uint, uint, float64, float64, error) {
	return a.proc.LookupInvoicePI(ctx, tenantID, invoiceID)
}

// financePosterProc bridges finance.Service to procurement.FinancePostGL —
// the return type on PostPIToGL is any so we don't leak finance types.
type financePosterProc struct{ fin *finance.Service }

func (f financePosterProc) PostPIToGL(ctx context.Context, tenantID, userID, piID uint) (interface{}, error) {
	return f.fin.PostPIToGL(ctx, tenantID, userID, piID)
}

func (f financePosterProc) RecordInvoicePaymentPI(ctx context.Context, tenantID, userID, piID uint, amount float64, method, referenceNo, notes string, bankAccountID *uint, paymentDate string) error {
	return f.fin.RecordInvoicePaymentPI(ctx, tenantID, userID, piID, amount, method, referenceNo, notes, bankAccountID, paymentDate)
}

type financePosterSales struct{ fin *finance.Service }

func (f financePosterSales) PostSIToGL(ctx context.Context, tenantID, userID, siID uint) (interface{}, error) {
	return f.fin.PostSIToGL(ctx, tenantID, userID, siID)
}

func (f financePosterSales) RecordInvoicePaymentSI(ctx context.Context, tenantID, userID, siID uint, amount float64, method, referenceNo, notes string, bankAccountID *uint, paymentDate string) error {
	return f.fin.RecordInvoicePaymentSI(ctx, tenantID, userID, siID, amount, method, referenceNo, notes, bankAccountID, paymentDate)
}

// salesAllocAdapter — bridges sales.AllocationReserver to inventory's
// AllocationService. The two interfaces are deliberately declared
// separately (sales can't import inventory without a cycle) so this
// adapter translates the local sales.SOAllocReserveRequest into
// inventory.ReserveRequest at the call site.
type salesAllocAdapter struct{ alloc *inventory.AllocationService }

func (a salesAllocAdapter) Reserve(ctx context.Context, tenantID uint, req sales.SOAllocReserveRequest) error {
	_, err := a.alloc.Reserve(ctx, inventory.ReserveRequest{
		TenantID: tenantID,
		ScopeKey: inventory.ScopeKey{
			ProductID:   req.ProductID,
			VariantID:   req.VariantID,
			WarehouseID: req.WarehouseID,
		},
		Quantity:    req.Quantity,
		SourceType:  inventory.AllocSourceSOLine,
		SourceID:    req.SourceID,
		SourceDocID: req.SourceDocID,
		Notes:       req.Notes,
		OnUpdate:    req.OnUpdate,
	})
	return err
}

func (a salesAllocAdapter) Release(ctx context.Context, tenantID uint, sourceType string, sourceID uint) error {
	return a.alloc.Release(ctx, tenantID, sourceType, sourceID)
}

func (a salesAllocAdapter) Consume(ctx context.Context, tenantID uint, sourceType string, sourceID uint) error {
	return a.alloc.Consume(ctx, tenantID, sourceType, sourceID)
}

func (a salesAllocAdapter) ReleaseByDoc(ctx context.Context, tenantID uint, sourceType string, docID uint) error {
	return a.alloc.ReleaseByDoc(ctx, tenantID, sourceType, docID)
}
