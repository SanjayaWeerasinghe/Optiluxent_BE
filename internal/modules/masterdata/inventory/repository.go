package inventory

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CountWarehouses(ctx context.Context, tenantID uint) (int64, error)
	ListWarehouses(ctx context.Context, tenantID uint, limit, offset int) ([]Warehouse, error)
	GetWarehouse(ctx context.Context, tenantID, id uint) (*Warehouse, error)
	CreateWarehouse(ctx context.Context, w *Warehouse) error
	UpdateWarehouse(ctx context.Context, w *Warehouse) error
	DeleteWarehouse(ctx context.Context, tenantID, id uint) error

	ListLocations(ctx context.Context, tenantID, warehouseID uint) ([]StorageLocation, error)
	CreateLocation(ctx context.Context, l *StorageLocation) error
	UpdateLocation(ctx context.Context, l *StorageLocation) error

	CreateStockEntry(ctx context.Context, e *StockLedger) error
	GetStockBalance(ctx context.Context, tenantID, warehouseID uint, productID *uint) ([]StockBalance, error)
	ListStockLedger(ctx context.Context, tenantID, productID, warehouseID uint, limit int) ([]StockLedger, error)
}

type dbRepository struct {
	db       *gorm.DB
	ledgerDB *sql.DB
}

func NewRepository(db *gorm.DB, ledgerDB *sql.DB) Repository {
	return &dbRepository{db: db, ledgerDB: ledgerDB}
}

func (r *dbRepository) CountWarehouses(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&Warehouse{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *dbRepository) ListWarehouses(ctx context.Context, tenantID uint, limit, offset int) ([]Warehouse, error) {
	var rows []Warehouse
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	return rows, q.Order("name").Find(&rows).Error
}

func (r *dbRepository) GetWarehouse(ctx context.Context, tenantID, id uint) (*Warehouse, error) {
	var w Warehouse
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&w).Error
	return &w, err
}

func (r *dbRepository) CreateWarehouse(ctx context.Context, w *Warehouse) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *dbRepository) UpdateWarehouse(ctx context.Context, w *Warehouse) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *dbRepository) DeleteWarehouse(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Warehouse{}).Error
}

func (r *dbRepository) ListLocations(ctx context.Context, tenantID, warehouseID uint) ([]StorageLocation, error) {
	var rows []StorageLocation
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND warehouse_id = ?", tenantID, warehouseID).
		Order("name").Find(&rows).Error
}

func (r *dbRepository) CreateLocation(ctx context.Context, l *StorageLocation) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *dbRepository) UpdateLocation(ctx context.Context, l *StorageLocation) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *dbRepository) CreateStockEntry(ctx context.Context, e *StockLedger) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *dbRepository) GetStockBalance(ctx context.Context, tenantID, warehouseID uint, productID *uint) ([]StockBalance, error) {
	q := r.db.WithContext(ctx).Table("stock_balances").
		Select("product_id, variant_id, warehouse_id, location_id, quantity, 0 as total_cost").
		Where("tenant_id = ? AND quantity != 0", tenantID)

	if warehouseID != 0 {
		q = q.Where("warehouse_id = ?", warehouseID)
	}
	if productID != nil {
		q = q.Where("product_id = ?", *productID)
	}

	var rows []StockBalance
	return rows, q.Find(&rows).Error
}

// ListStockLedger reads from the ClickHouse analytics store — that's where
// every stock movement is written (via inventory/procurement writeToLedger
// paths). The Postgres stock_ledger table exists as a schema mirror for
// legacy tooling but nothing writes to it in the current flow.
func (r *dbRepository) ListStockLedger(ctx context.Context, tenantID, productID, warehouseID uint, limit int) ([]StockLedger, error) {
	if r.ledgerDB == nil {
		return nil, nil
	}
	q := `SELECT tenant_id, product_id, variant_id, warehouse_id, location_id,
	             transaction_type, reference_type, reference_id,
	             quantity, unit_cost, total_cost, transaction_date, notes, created_by
	      FROM stock_ledger WHERE tenant_id = ?`
	args := []interface{}{tenantID}
	if productID != 0 {
		q += ` AND product_id = ?`
		args = append(args, productID)
	}
	if warehouseID != 0 {
		q += ` AND warehouse_id = ?`
		args = append(args, warehouseID)
	}
	q += ` ORDER BY transaction_date DESC, created_at DESC`
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := r.ledgerDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []StockLedger
	for rows.Next() {
		var (
			e               StockLedger
			tenantIDRaw     uint64
			productIDRaw    uint64
			variantIDRaw    uint64
			warehouseIDRaw  uint64
			locationIDRaw   uint64
			referenceIDRaw  uint64
			createdByRaw    uint64
			transactionDate time.Time
		)
		if err := rows.Scan(
			&tenantIDRaw, &productIDRaw, &variantIDRaw, &warehouseIDRaw, &locationIDRaw,
			&e.TransactionType, &e.ReferenceType, &referenceIDRaw,
			&e.Quantity, &e.UnitCost, &e.TotalCost, &transactionDate, &e.Notes, &createdByRaw,
		); err != nil {
			return nil, err
		}
		e.TenantID = uint(tenantIDRaw)
		e.ProductID = uint(productIDRaw)
		e.WarehouseID = uint(warehouseIDRaw)
		if variantIDRaw > 0 {
			v := uint(variantIDRaw)
			e.VariantID = &v
		}
		if locationIDRaw > 0 {
			l := uint(locationIDRaw)
			e.LocationID = &l
		}
		if referenceIDRaw > 0 {
			r := uint(referenceIDRaw)
			e.ReferenceID = &r
		}
		if createdByRaw > 0 {
			c := uint(createdByRaw)
			e.CreatedBy = &c
		}
		e.TransactionDate = transactionDate.Format("2006-01-02")
		out = append(out, e)
	}
	return out, rows.Err()
}
