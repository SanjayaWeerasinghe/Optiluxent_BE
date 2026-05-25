package inventory

import (
	"time"

	"gorm.io/gorm"
)

type Warehouse struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	TenantID  uint           `json:"tenant_id"  gorm:"not null;index"`
	Code      string         `json:"code"       gorm:"not null;size:20"`
	Name      string         `json:"name"       gorm:"not null;size:200"`
	Address   string         `json:"address"`
	IsActive  bool           `json:"is_active"  gorm:"not null;default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Warehouse) TableName() string { return "warehouses" }

type StorageLocation struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null;index"`
	WarehouseID  uint      `json:"warehouse_id"  gorm:"not null;index"`
	Code         string    `json:"code"          gorm:"not null;size:20"`
	Name         string    `json:"name"          gorm:"not null;size:200"`
	LocationType string    `json:"location_type" gorm:"not null;size:20;default:STORAGE"`
	IsActive     bool      `json:"is_active"     gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (StorageLocation) TableName() string { return "storage_locations" }

type StockLedger struct {
	ID              uint      `json:"id"               gorm:"primaryKey"`
	TenantID        uint      `json:"tenant_id"        gorm:"not null;index"`
	ProductID       uint      `json:"product_id"       gorm:"not null;index"`
	VariantID       *uint     `json:"variant_id"`
	WarehouseID     uint      `json:"warehouse_id"     gorm:"not null;index"`
	LocationID      *uint     `json:"location_id"`
	TransactionType string    `json:"transaction_type" gorm:"not null;size:30"`
	ReferenceType   string    `json:"reference_type"   gorm:"size:50"`
	ReferenceID     *uint     `json:"reference_id"`
	Quantity        float64   `json:"quantity"         gorm:"not null"`
	UnitCost        float64   `json:"unit_cost"        gorm:"not null;default:0"`
	TotalCost       float64   `json:"total_cost"       gorm:"not null;default:0"`
	TransactionDate string    `json:"transaction_date" gorm:"type:date;not null"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       *uint     `json:"created_by"`
}

func (StockLedger) TableName() string { return "stock_ledger" }

type StockBalance struct {
	ProductID   uint    `json:"product_id"`
	VariantID   *uint   `json:"variant_id"`
	WarehouseID uint    `json:"warehouse_id"`
	LocationID  *uint   `json:"location_id"`
	Quantity    float64 `json:"quantity"`
	TotalCost   float64 `json:"total_cost"`
}
