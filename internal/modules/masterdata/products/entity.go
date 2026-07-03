package products

import (
	"time"

	"gorm.io/gorm"
)

type ProductCategory struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	TenantID  uint           `json:"tenant_id"  gorm:"not null;index"`
	Code      string         `json:"code"       gorm:"not null;size:20"`
	Name      string         `json:"name"       gorm:"not null;size:200"`
	ParentID  *uint          `json:"parent_id"`
	IsActive  bool           `json:"is_active"  gorm:"not null;default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (ProductCategory) TableName() string { return "product_categories" }

type UnitOfMeasure struct {
	ID               uint      `json:"id"                gorm:"primaryKey"`
	TenantID         uint      `json:"tenant_id"         gorm:"not null;index"`
	Code             string    `json:"code"              gorm:"not null;size:20"`
	Name             string    `json:"name"              gorm:"not null;size:100"`
	UOMType          string    `json:"uom_type"          gorm:"not null;size:20;default:UNIT"`
	BaseUOMID        *uint     `json:"base_uom_id"`
	ConversionFactor float64   `json:"conversion_factor" gorm:"not null;default:1"`
	IsActive         bool      `json:"is_active"         gorm:"not null;default:true"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (UnitOfMeasure) TableName() string { return "units_of_measure" }

type Product struct {
	ID             uint           `json:"id"              gorm:"primaryKey"`
	TenantID       uint           `json:"tenant_id"       gorm:"not null;index"`
	Code           string         `json:"code"            gorm:"not null;size:50"`
	Name           string         `json:"name"            gorm:"not null;size:200"`
	Description    string         `json:"description"`
	ProductType    string         `json:"product_type"    gorm:"not null;size:20;default:FINISHED"`
	CategoryID     *uint          `json:"category_id"`
	BaseUOMID       uint           `json:"base_uom_id"       gorm:"not null"`
	PurchaseUOMID   *uint          `json:"purchase_uom_id"`
	SalesUOMID      *uint          `json:"sales_uom_id"`
	StockUOMID      *uint          `json:"stock_uom_id"`
	ProductionUOMID *uint          `json:"production_uom_id"`
	TaxCodeID       *uint          `json:"tax_code_id"`
	CostPrice      float64        `json:"cost_price"      gorm:"not null;default:0"`
	StandardPrice  float64        `json:"standard_price"  gorm:"not null;default:0"`
	MinStockQty    float64        `json:"min_stock_qty"   gorm:"not null;default:0"`
	ReorderQty     float64        `json:"reorder_qty"     gorm:"not null;default:0"`
	LeadTimeDays   int            `json:"lead_time_days"  gorm:"not null;default:0"`
	IsPurchased    bool           `json:"is_purchased"    gorm:"not null;default:true"`
	IsSold         bool           `json:"is_sold"         gorm:"not null;default:true"`
	IsManufactured bool           `json:"is_manufactured" gorm:"not null;default:false"`
	IsActive       bool           `json:"is_active"       gorm:"not null;default:true"`
	Notes          string         `json:"notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Product) TableName() string { return "products" }

type ProductVariant struct {
	ID         uint      `json:"id"          gorm:"primaryKey"`
	TenantID   uint      `json:"tenant_id"   gorm:"not null;index"`
	ProductID  uint      `json:"product_id"  gorm:"not null;index"`
	Code       string    `json:"code"        gorm:"not null;size:50"`
	Name       string    `json:"name"        gorm:"not null;size:200"`
	Attributes []byte    `json:"attributes"  gorm:"type:jsonb"`
	CostPrice  float64   `json:"cost_price"  gorm:"not null;default:0"`
	SalesPrice float64   `json:"sales_price" gorm:"not null;default:0"`
	IsActive   bool      `json:"is_active"   gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (ProductVariant) TableName() string { return "product_variants" }

type ProductPrice struct {
	ID            uint      `json:"id"             gorm:"primaryKey"`
	TenantID      uint      `json:"tenant_id"      gorm:"not null;index"`
	ProductID     uint      `json:"product_id"     gorm:"not null;index"`
	PartyID       *uint     `json:"party_id"`
	PriceType     string    `json:"price_type"     gorm:"not null;size:20;default:SALES"`
	CurrencyID    uint      `json:"currency_id"    gorm:"not null"`
	Price         float64   `json:"price"          gorm:"not null"`
	MinQty        float64   `json:"min_qty"        gorm:"not null;default:0"`
	EffectiveFrom string    `json:"effective_from" gorm:"type:date;not null"`
	EffectiveTo   *string   `json:"effective_to"   gorm:"type:date"`
	IsActive      bool      `json:"is_active"      gorm:"not null;default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ProductPrice) TableName() string { return "product_prices" }
