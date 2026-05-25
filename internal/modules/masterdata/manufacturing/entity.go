package manufacturing

import (
	"time"

	"gorm.io/gorm"
)

type BillOfMaterials struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	TenantID  uint           `json:"tenant_id"  gorm:"not null;index"`
	Code      string         `json:"code"       gorm:"not null;size:50"`
	ProductID uint           `json:"product_id" gorm:"not null;index"`
	VariantID *uint          `json:"variant_id"`
	Quantity  float64        `json:"quantity"   gorm:"not null;default:1"`
	UOMID     uint           `json:"uom_id"     gorm:"not null"`
	BOMType   string         `json:"bom_type"   gorm:"not null;size:20;default:MANUFACTURE"`
	IsActive  bool           `json:"is_active"  gorm:"not null;default:true"`
	Notes     string         `json:"notes"`
	Lines     []BOMLine      `json:"lines,omitempty" gorm:"foreignKey:BOMID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (BillOfMaterials) TableName() string { return "bill_of_materials" }

type BOMLine struct {
	ID          uint      `json:"id"           gorm:"primaryKey"`
	BOMID       uint      `json:"bom_id"       gorm:"not null;index"`
	TenantID    uint      `json:"tenant_id"    gorm:"not null"`
	ComponentID uint      `json:"component_id" gorm:"not null;index"`
	VariantID   *uint     `json:"variant_id"`
	Quantity    float64   `json:"quantity"     gorm:"not null"`
	UOMID       uint      `json:"uom_id"       gorm:"not null"`
	ScrapPct    float64   `json:"scrap_percent" gorm:"column:scrap_percent;not null;default:0"`
	Sequence    int       `json:"sequence"     gorm:"not null;default:10"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (BOMLine) TableName() string { return "bom_lines" }

type WorkCenter struct {
	ID           uint           `json:"id"             gorm:"primaryKey"`
	TenantID     uint           `json:"tenant_id"      gorm:"not null;index"`
	Code         string         `json:"code"           gorm:"not null;size:20"`
	Name         string         `json:"name"           gorm:"not null;size:200"`
	Capacity     float64        `json:"capacity"       gorm:"not null;default:1"`
	CostPerHour  float64        `json:"cost_per_hour"  gorm:"not null;default:0"`
	CurrencyID   uint           `json:"currency_id"    gorm:"not null"`
	IsActive     bool           `json:"is_active"      gorm:"not null;default:true"`
	Notes        string         `json:"notes"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (WorkCenter) TableName() string { return "work_centers" }

type Routing struct {
	ID         uint               `json:"id"         gorm:"primaryKey"`
	TenantID   uint               `json:"tenant_id"  gorm:"not null;index"`
	Code       string             `json:"code"       gorm:"not null;size:50"`
	ProductID  uint               `json:"product_id" gorm:"not null;index"`
	IsActive   bool               `json:"is_active"  gorm:"not null;default:true"`
	Notes      string             `json:"notes"`
	Operations []RoutingOperation `json:"operations,omitempty" gorm:"foreignKey:RoutingID"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	DeletedAt  gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`
}

func (Routing) TableName() string { return "routings" }

type RoutingOperation struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	RoutingID    uint      `json:"routing_id"    gorm:"not null;index"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null"`
	Sequence     int       `json:"sequence"      gorm:"not null;default:10"`
	Name         string    `json:"name"          gorm:"not null;size:200"`
	WorkCenterID uint      `json:"work_center_id" gorm:"not null;index"`
	SetupTime    float64   `json:"setup_time"    gorm:"not null;default:0"`
	CycleTime    float64   `json:"cycle_time"    gorm:"not null;default:0"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RoutingOperation) TableName() string { return "routing_operations" }
