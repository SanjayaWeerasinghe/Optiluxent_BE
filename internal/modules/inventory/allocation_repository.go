package inventory

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// AllocationRepository — data access for stock_allocations. Kept in a
// separate interface from the main inventory Repository so callers (both
// inside this module and via the cross-module AllocationReserver interface)
// don't need to depend on the whole surface area.
type AllocationRepository interface {
	Create(ctx context.Context, a *Allocation) error
	Get(ctx context.Context, tenantID, id uint) (*Allocation, error)
	GetBySource(ctx context.Context, tenantID uint, sourceType string, sourceID uint) (*Allocation, error)
	SetStatus(ctx context.Context, tenantID, id uint, status string) error
	SetStatusBySource(ctx context.Context, tenantID uint, sourceType string, sourceID uint, status string) error
	ReleaseByDoc(ctx context.Context, tenantID uint, sourceType string, docID uint) error

	// SumActiveAllocations returns the running total of ACTIVE reservations
	// against a single stock scope key. Callers use this to compute
	// available = on_hand - reserved. NULL-handling on optional dimensions
	// (variant, location, production) mirrors the stock_balances uniqueness
	// semantics: a NULL in the caller's key matches only rows with NULL in
	// the same column.
	SumActiveAllocations(ctx context.Context, tenantID uint, key ScopeKey) (float64, error)

	// SumActiveTotalForProduct returns the SUM(quantity) of ACTIVE allocations
	// across every scope for a (tenant, product, warehouse) pair. Used to
	// derive the Stock Overview "Reserved" column.
	SumReservedByProductWarehouse(ctx context.Context, tenantID, productID, warehouseID uint) (float64, error)

	// ListActive — for the FE Allocations section. `limit=0` = unbounded.
	ListActive(ctx context.Context, tenantID uint, filters AllocationFilters, limit, offset int) ([]Allocation, error)
	CountActive(ctx context.Context, tenantID uint, filters AllocationFilters) (int64, error)

	// Access to the underlying gorm.DB — allocation_service reaches for a
	// transaction with SELECT FOR UPDATE on stock_balances, which is a
	// cross-table read-modify-write.
	DB() *gorm.DB
}

// ScopeKey — the (product, warehouse, variant, location, production) tuple
// that identifies a stock row. Uses pointers for optional dimensions so a nil
// means "the row where this column IS NULL," matching stock_balances'
// uniqueness treatment (NULLS NOT DISTINCT).
type ScopeKey struct {
	ProductID    uint
	VariantID    *uint
	WarehouseID  uint
	LocationID   *uint
	ProductionID *uint
}

// AllocationFilters — narrow the ListActive query for the FE.
type AllocationFilters struct {
	ProductID    *uint
	WarehouseID  *uint
	SourceType   string
	SourceDocID  *uint
}

type dbAllocationRepository struct{ db *gorm.DB }

func NewAllocationRepository(db *gorm.DB) AllocationRepository {
	return &dbAllocationRepository{db: db}
}

func (r *dbAllocationRepository) DB() *gorm.DB { return r.db }

func (r *dbAllocationRepository) Create(ctx context.Context, a *Allocation) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *dbAllocationRepository) Get(ctx context.Context, tenantID, id uint) (*Allocation, error) {
	var a Allocation
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *dbAllocationRepository) GetBySource(ctx context.Context, tenantID uint, sourceType string, sourceID uint) (*Allocation, error) {
	var a Allocation
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND source_type = ? AND source_id = ?", tenantID, sourceType, sourceID).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *dbAllocationRepository) SetStatus(ctx context.Context, tenantID, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&Allocation{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{"status": status, "updated_at": gorm.Expr("NOW()")}).Error
}

func (r *dbAllocationRepository) SetStatusBySource(ctx context.Context, tenantID uint, sourceType string, sourceID uint, status string) error {
	// Idempotent — only touches ACTIVE rows. A double-consume/double-release
	// no-ops silently instead of raising a spurious error.
	return r.db.WithContext(ctx).Model(&Allocation{}).
		Where("tenant_id = ? AND source_type = ? AND source_id = ? AND status = ?",
			tenantID, sourceType, sourceID, AllocStatusActive).
		Updates(map[string]interface{}{"status": status, "updated_at": gorm.Expr("NOW()")}).Error
}

func (r *dbAllocationRepository) ReleaseByDoc(ctx context.Context, tenantID uint, sourceType string, docID uint) error {
	return r.db.WithContext(ctx).Model(&Allocation{}).
		Where("tenant_id = ? AND source_type = ? AND source_doc_id = ? AND status = ?",
			tenantID, sourceType, docID, AllocStatusActive).
		Updates(map[string]interface{}{"status": AllocStatusCancelled, "updated_at": gorm.Expr("NOW()")}).Error
}

// scopeWhere builds the "same scope key" WHERE clause. Nullable fields use
// `IS NULL` when the caller passes nil so a general-stock scope doesn't
// collide with a refining-scope row.
func scopeWhere(q *gorm.DB, tenantID uint, key ScopeKey) *gorm.DB {
	q = q.Where("tenant_id = ? AND product_id = ? AND warehouse_id = ?",
		tenantID, key.ProductID, key.WarehouseID)
	if key.VariantID != nil {
		q = q.Where("variant_id = ?", *key.VariantID)
	} else {
		q = q.Where("variant_id IS NULL")
	}
	if key.LocationID != nil {
		q = q.Where("location_id = ?", *key.LocationID)
	} else {
		q = q.Where("location_id IS NULL")
	}
	if key.ProductionID != nil {
		q = q.Where("production_id = ?", *key.ProductionID)
	} else {
		q = q.Where("production_id IS NULL")
	}
	return q
}

func (r *dbAllocationRepository) SumActiveAllocations(ctx context.Context, tenantID uint, key ScopeKey) (float64, error) {
	var total float64
	q := scopeWhere(r.db.WithContext(ctx).Model(&Allocation{}), tenantID, key).
		Where("status = ?", AllocStatusActive).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total)
	if q.Error != nil {
		return 0, fmt.Errorf("sum active allocations: %w", q.Error)
	}
	return total, nil
}

func (r *dbAllocationRepository) SumReservedByProductWarehouse(ctx context.Context, tenantID, productID, warehouseID uint) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&Allocation{}).
		Where("tenant_id = ? AND product_id = ? AND warehouse_id = ? AND status = ?",
			tenantID, productID, warehouseID, AllocStatusActive).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	return total, err
}

// allocFilterWhere applies the AllocationFilters conditions to a *gorm.DB —
// factored so ListActive + CountActive stay in sync.
func allocFilterWhere(q *gorm.DB, tenantID uint, filters AllocationFilters) *gorm.DB {
	q = q.Where("tenant_id = ? AND status = ?", tenantID, AllocStatusActive)
	if filters.ProductID != nil {
		q = q.Where("product_id = ?", *filters.ProductID)
	}
	if filters.WarehouseID != nil {
		q = q.Where("warehouse_id = ?", *filters.WarehouseID)
	}
	if filters.SourceType != "" {
		q = q.Where("source_type = ?", filters.SourceType)
	}
	if filters.SourceDocID != nil {
		q = q.Where("source_doc_id = ?", *filters.SourceDocID)
	}
	return q
}

func (r *dbAllocationRepository) ListActive(ctx context.Context, tenantID uint, filters AllocationFilters, limit, offset int) ([]Allocation, error) {
	q := allocFilterWhere(r.db.WithContext(ctx), tenantID, filters).Order("id DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []Allocation
	return rows, q.Find(&rows).Error
}

func (r *dbAllocationRepository) CountActive(ctx context.Context, tenantID uint, filters AllocationFilters) (int64, error) {
	var n int64
	return n, allocFilterWhere(r.db.WithContext(ctx).Model(&Allocation{}), tenantID, filters).Count(&n).Error
}
