package financial

import (
	auditinfra "erp-system/internal/infrastructure/audit"
	"erp-system/internal/infrastructure/http/middleware"
	"erp-system/internal/domain/rbac"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes mounts all Financial routes onto the given router.
// The router must already have auth + tenant middleware applied.
func RegisterRoutes(router fiber.Router, h *Handler, enforcer rbac.Enforcer, auditLogger *auditinfra.Logger) {
	auditMW := middleware.AuditLog(auditLogger)
	canRead   := middleware.RequirePermission(enforcer, "masterdata", "read")
	canWrite  := middleware.RequirePermission(enforcer, "masterdata", "write")

	// Currencies (global reference data)
	curr := router.Group("/currencies")
	curr.Get("",              canRead,  h.ListCurrencies)
	curr.Post("",             canWrite, auditMW, h.CreateCurrency)
	curr.Put("/:id",          canWrite, auditMW, h.UpdateCurrency)
	curr.Put("/:id/set-base", canWrite, auditMW, h.SetBaseCurrency)

	// Exchange rates (tenant-scoped)
	er := router.Group("/exchange-rates")
	er.Get("",        canRead,  h.ListExchangeRates)
	er.Get("/latest", canRead,  h.GetLatestRates)
	er.Post("",       canWrite, auditMW, h.CreateExchangeRate)

	// Chart of accounts (tenant-scoped)
	coa := router.Group("/chart-of-accounts")
	coa.Get("",      canRead,  h.ListCoA)
	coa.Post("",     canWrite, auditMW, h.CreateCoA)
	coa.Get("/:id",  canRead,  h.GetCoA)
	coa.Put("/:id",  canWrite, auditMW, h.UpdateCoA)
	coa.Delete("/:id", middleware.RequirePermission(enforcer, "masterdata", "delete"), auditMW, h.DeleteCoA)

	// Cost centers (tenant-scoped)
	cc := router.Group("/cost-centers")
	cc.Get("",     canRead,  h.ListCostCenters)
	cc.Post("",    canWrite, auditMW, h.CreateCostCenter)
	cc.Put("/:id", canWrite, auditMW, h.UpdateCostCenter)

	// Payment terms (tenant-scoped)
	pt := router.Group("/payment-terms")
	pt.Get("",     canRead,  h.ListPaymentTerms)
	pt.Post("",    canWrite, auditMW, h.CreatePaymentTerm)
	pt.Put("/:id", canWrite, auditMW, h.UpdatePaymentTerm)

	// Banks (global reference data)
	banks := router.Group("/banks")
	banks.Get("",  canRead,  h.ListBanks)
	banks.Post("", canWrite, auditMW, h.CreateBank)

	// Company bank accounts (tenant-scoped)
	ba := router.Group("/bank-accounts")
	ba.Get("",     canRead,  h.ListBankAccounts)
	ba.Post("",    canWrite, auditMW, h.CreateBankAccount)
	ba.Put("/:id", canWrite, auditMW, h.UpdateBankAccount)

	// Tax codes (tenant-scoped)
	tc := router.Group("/tax-codes")
	tc.Get("",     canRead,  h.ListTaxCodes)
	tc.Post("",    canWrite, auditMW, h.CreateTaxCode)
	tc.Put("/:id", canWrite, auditMW, h.UpdateTaxCode)

	// Tax groups (tenant-scoped)
	tg := router.Group("/tax-groups")
	tg.Get("",     canRead,  h.ListTaxGroups)
	tg.Post("",    canWrite, auditMW, h.CreateTaxGroup)
	tg.Put("/:id", canWrite, auditMW, h.UpdateTaxGroup)
}
