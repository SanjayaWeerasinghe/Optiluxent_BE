package inventory

import (
	"context"

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

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

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

func (r *dbRepository) ListStockLedger(ctx context.Context, tenantID, productID, warehouseID uint, limit int) ([]StockLedger, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if productID != 0 {
		q = q.Where("product_id = ?", productID)
	}
	if warehouseID != 0 {
		q = q.Where("warehouse_id = ?", warehouseID)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []StockLedger
	return rows, q.Order("created_at DESC").Find(&rows).Error
}
