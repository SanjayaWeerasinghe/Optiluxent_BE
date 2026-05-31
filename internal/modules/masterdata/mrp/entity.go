package mrp

import "time"

// MaterialMaster holds MRP planning parameters for a product.
type MaterialMaster struct {
	ID                  uint     `json:"id"                   gorm:"primaryKey"`
	TenantID            uint     `json:"tenant_id"            gorm:"not null;index"`
	ProductID           uint     `json:"product_id"           gorm:"not null;index"`
	MaterialType        string   `json:"material_type"        gorm:"not null;size:20;default:RAW_MATERIAL"`
	MRPType             string   `json:"mrp_type"             gorm:"not null;size:20;default:MRP"`
	ProcurementType     string   `json:"procurement_type"     gorm:"not null;size:20;default:EXTERNAL"`
	PurchaseLeadTime    int      `json:"purchase_lead_time"   gorm:"not null;default:0"`
	ProductionLeadTime  int      `json:"production_lead_time" gorm:"not null;default:0"`
	SafetyStock         float64  `json:"safety_stock"         gorm:"not null;default:0"`
	ReorderPoint        float64  `json:"reorder_point"        gorm:"not null;default:0"`
	MinOrderQty         float64  `json:"min_order_qty"        gorm:"not null;default:0"`
	MaxOrderQty         *float64 `json:"max_order_qty"`
	LotSizeType         string   `json:"lot_size_type"        gorm:"not null;size:20;default:LOT_FOR_LOT"`
	FixedLotSize        *float64 `json:"fixed_lot_size"`
	ABCClass            string   `json:"abc_class"            gorm:"size:1"`
	ShelfLifeDays       *int     `json:"shelf_life_days"`
	IsActive            bool     `json:"is_active"            gorm:"not null;default:true"`
	Notes               string   `json:"notes"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (MaterialMaster) TableName() string { return "mrp_material_master" }

// MaterialMasterView joins product info for list/get responses.
type MaterialMasterView struct {
	MaterialMaster
	ProductCode string `json:"product_code" gorm:"->"`
	ProductName string `json:"product_name" gorm:"->"`
}

func (MaterialMasterView) TableName() string { return "mrp_material_master" }
