package inventory

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	httputil "erp-system/pkg/http"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func tenantFromCtx(c *fiber.Ctx) uint {
	v, _ := c.Locals("tenant_id").(uint)
	return v
}

func validateStruct(v *validator.Validate, s interface{}) map[string]interface{} {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	errs := map[string]interface{}{}
	for _, e := range err.(validator.ValidationErrors) {
		errs[e.Field()] = e.Tag()
	}
	return errs
}

// ── Warehouses ────────────────────────────────────────────────────────────────

func (h *Handler) ListWarehouses(c *fiber.Ctx) error {
	rows, err := h.svc.ListWarehouses(c.Context(), tenantFromCtx(c))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list warehouses")
	}
	return httputil.Success(c, "warehouses retrieved", rows)
}

func (h *Handler) GetWarehouse(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	w, err := h.svc.GetWarehouse(c.Context(), tenantFromCtx(c), uint(id))
	if err != nil {
		return httputil.NotFound(c, "warehouse not found")
	}
	return httputil.Success(c, "warehouse retrieved", w)
}

func (h *Handler) CreateWarehouse(c *fiber.Ctx) error {
	var req CreateWarehouseRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	w, err := h.svc.CreateWarehouse(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "warehouse created", w)
}

func (h *Handler) UpdateWarehouse(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateWarehouseRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	w, err := h.svc.UpdateWarehouse(c.Context(), tenantFromCtx(c), uint(id), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "warehouse updated", w)
}

func (h *Handler) DeleteWarehouse(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	if err := h.svc.DeleteWarehouse(c.Context(), tenantFromCtx(c), uint(id)); err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Storage Locations ─────────────────────────────────────────────────────────

func (h *Handler) ListLocations(c *fiber.Ctx) error {
	warehouseID, err := strconv.ParseUint(c.Params("warehouseId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid warehouse id")
	}
	rows, err := h.svc.ListLocations(c.Context(), tenantFromCtx(c), uint(warehouseID))
	if err != nil {
		return httputil.InternalServerError(c, "failed to list locations")
	}
	return httputil.Success(c, "locations retrieved", rows)
}

func (h *Handler) CreateLocation(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	warehouseID, err := strconv.ParseUint(c.Params("warehouseId"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid warehouse id")
	}
	var req CreateLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	l, err := h.svc.CreateLocation(c.Context(), tenantID, uint(warehouseID), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "location created", l)
}

func (h *Handler) UpdateLocation(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "invalid id")
	}
	var req UpdateLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	existing := &StorageLocation{ID: uint(id), TenantID: tenantID}
	l, err := h.svc.UpdateLocation(c.Context(), tenantID, &req, existing)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "location updated", l)
}

// ── Stock Ledger ──────────────────────────────────────────────────────────────

func (h *Handler) CreateStockEntry(c *fiber.Ctx) error {
	var req CreateStockEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	e, err := h.svc.CreateStockEntry(c.Context(), tenantFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "stock entry created", e)
}

func (h *Handler) GetStockBalance(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	warehouseID, _ := strconv.ParseUint(c.Query("warehouse_id"), 10, 64)
	productIDStr := c.Query("product_id")
	var productID *uint
	if productIDStr != "" {
		pid, err := strconv.ParseUint(productIDStr, 10, 64)
		if err == nil {
			p := uint(pid)
			productID = &p
		}
	}
	rows, err := h.svc.GetStockBalance(c.Context(), tenantID, uint(warehouseID), productID)
	if err != nil {
		return httputil.InternalServerError(c, "failed to get stock balance")
	}
	return httputil.Success(c, "stock balance retrieved", rows)
}

func (h *Handler) ListStockLedger(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	warehouseID, _ := strconv.ParseUint(c.Query("warehouse_id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit", "100"))

	rows, err := h.svc.ListStockLedger(c.Context(), tenantID, uint(productID), uint(warehouseID), limit)
	if err != nil {
		return httputil.InternalServerError(c, "failed to list stock ledger")
	}
	return httputil.Success(c, "stock ledger retrieved", rows)
}
