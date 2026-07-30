package inventory

import "time"

// Stock Allocation — reserves qty against a specific (product, warehouse,
// variant, location, production) scope for a pending downstream doc.
// See migration 000047_stock_allocations for the full column contract.

const (
	AllocSourceSOLine = "SO_LINE"
	AllocSourceMRLine = "MR_LINE"
	AllocSourceGILine = "GI_LINE"
	AllocSourceGTLine = "GT_LINE"

	AllocStatusActive    = "ACTIVE"
	AllocStatusConsumed  = "CONSUMED"
	AllocStatusCancelled = "CANCELLED"
)

// Allocation mirrors the stock_allocations row 1:1. Nullable optional fields
// use *uint so we can preserve NULL semantics on the DB side — refining
// segregation depends on `production_id IS NULL` behaving correctly under the
// scope-matching queries.
type Allocation struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null;index"`
	ProductID    uint      `json:"product_id"    gorm:"not null"`
	VariantID    *uint     `json:"variant_id"`
	WarehouseID  uint      `json:"warehouse_id"  gorm:"not null"`
	LocationID   *uint     `json:"location_id"`
	ProductionID *uint     `json:"production_id"`
	Quantity     float64   `json:"quantity"      gorm:"type:decimal(18,4);not null"`
	SourceType   string    `json:"source_type"   gorm:"not null;size:20"`
	SourceID     uint      `json:"source_id"     gorm:"not null"`
	SourceDocID  uint      `json:"source_doc_id" gorm:"not null"`
	Status       string    `json:"status"        gorm:"not null;size:15;default:ACTIVE"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Allocation) TableName() string { return "stock_allocations" }
