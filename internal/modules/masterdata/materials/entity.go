package materials

import "time"

type Material struct {
	ID           uint   `json:"id"            gorm:"primaryKey"`
	TenantID     uint   `json:"tenant_id"     gorm:"not null;index"`
	Code         string `json:"code"          gorm:"not null;size:50"`
	Name         string `json:"name"          gorm:"not null;size:200"`
	MaterialType string `json:"material_type" gorm:"not null;size:20;default:RAW_MATERIAL"`
	Color        string `json:"color"         gorm:"size:100"`
	CategoryID   *uint  `json:"category_id"`
	ArticleCode  string `json:"article_code"  gorm:"size:100"`
	Ref2         string `json:"ref2"          gorm:"size:100"`
	Barcode      string `json:"barcode"       gorm:"size:100"`
	Description  string `json:"description"`
	IsActive     bool   `json:"is_active"     gorm:"not null;default:true"`
	Notes        string `json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Material) TableName() string { return "mm_materials" }

type MaterialPurchasing struct {
	ID               uint    `json:"id"                 gorm:"primaryKey"`
	MaterialID       uint    `json:"material_id"        gorm:"not null"`
	TenantID         uint    `json:"tenant_id"          gorm:"not null"`
	PurchasingUOMID  *uint   `json:"purchasing_uom_id"`
	UnderDeliveryPct float64 `json:"under_delivery_pct" gorm:"not null;default:0"`
	OverDeliveryPct  float64 `json:"over_delivery_pct"  gorm:"not null;default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PurchasingUOMCode string `json:"purchasing_uom_code" gorm:"->"`
	PurchasingUOMName string `json:"purchasing_uom_name" gorm:"->"`
}

func (MaterialPurchasing) TableName() string { return "mm_purchasing" }

type MaterialManufacturing struct {
	ID                    uint    `json:"id"                      gorm:"primaryKey"`
	MaterialID            uint    `json:"material_id"             gorm:"not null"`
	TenantID              uint    `json:"tenant_id"               gorm:"not null"`
	ProductionUOMID       *uint   `json:"production_uom_id"`
	ReorderQtyLevel       float64 `json:"reorder_qty_level"       gorm:"not null;default:0"`
	SafetyLevel           float64 `json:"safety_level"            gorm:"not null;default:0"`
	ProductionDays        int     `json:"production_days"         gorm:"not null;default:0"`
	DeliveryDays          int     `json:"delivery_days"           gorm:"not null;default:0"`
	GRNDays               int     `json:"grn_days"                gorm:"not null;default:0"`
	ProcurementRepeatDays int     `json:"procurement_repeat_days" gorm:"not null;default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProductionUOMCode string `json:"production_uom_code" gorm:"->"`
	ProductionUOMName string `json:"production_uom_name" gorm:"->"`
}

func (MaterialManufacturing) TableName() string { return "mm_manufacturing" }

type MaterialWarehouse struct {
	ID                   uint   `json:"id"                     gorm:"primaryKey"`
	MaterialID           uint   `json:"material_id"            gorm:"not null"`
	TenantID             uint   `json:"tenant_id"              gorm:"not null"`
	StockingUOMID        *uint  `json:"stocking_uom_id"`
	StockRemoval         string `json:"stock_removal"          gorm:"not null;size:10;default:FIFO"`
	StorageMain          bool   `json:"storage_main"           gorm:"not null;default:false"`
	StorageDamaged       bool   `json:"storage_damaged"        gorm:"not null;default:false"`
	StorageHold          bool   `json:"storage_hold"           gorm:"not null;default:false"`
	BatchProcess         bool   `json:"batch_process"          gorm:"not null;default:false"`
	ProductionDateCheck  bool   `json:"production_date_check"  gorm:"not null;default:false"`
	ExpiryDateCheck      bool   `json:"expiry_date_check"      gorm:"not null;default:false"`
	QCCheck              bool   `json:"qc_check"               gorm:"not null;default:false"`
	GRNWithPOUOMImage    bool   `json:"grn_with_po_uom_image"  gorm:"column:grn_with_po_uom_image;not null;default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StockingUOMCode string `json:"stocking_uom_code" gorm:"->"`
	StockingUOMName string `json:"stocking_uom_name" gorm:"->"`
}

func (MaterialWarehouse) TableName() string { return "mm_warehouse" }

type MaterialVendor struct {
	ID             uint     `json:"id"              gorm:"primaryKey"`
	MaterialID     uint     `json:"material_id"     gorm:"not null;index"`
	TenantID       uint     `json:"tenant_id"       gorm:"not null"`
	VendorID       *uint    `json:"vendor_id"`
	ArticleNo      string   `json:"article_no"      gorm:"size:100"`
	DeliveryDays   int      `json:"delivery_days"   gorm:"not null;default:0"`
	Cost           *float64 `json:"cost"`
	CurrencyID     *uint    `json:"currency_id"`
	ProjectedPrice *float64 `json:"projected_price"`
	MOQ            *float64 `json:"moq"`
	IsDefault      bool     `json:"is_default"      gorm:"not null;default:false"`

	CreatedAt time.Time `json:"created_at"`

	VendorName   string `json:"vendor_name"   gorm:"->"`
	CurrencyCode string `json:"currency_code" gorm:"->"`
}

func (MaterialVendor) TableName() string { return "mm_vendors" }

type MaterialMeasurement struct {
	ID              uint    `json:"id"               gorm:"primaryKey"`
	MaterialID      uint    `json:"material_id"      gorm:"not null;index"`
	TenantID        uint    `json:"tenant_id"        gorm:"not null"`
	BaseUOMID       uint    `json:"base_uom_id"      gorm:"not null"`
	TargetUOMID     uint    `json:"target_uom_id"    gorm:"not null"`
	ConversionRatio float64 `json:"conversion_ratio" gorm:"not null;default:1"`

	CreatedAt time.Time `json:"created_at"`

	BaseUOMCode   string `json:"base_uom_code"   gorm:"->"`
	BaseUOMName   string `json:"base_uom_name"   gorm:"->"`
	TargetUOMCode string `json:"target_uom_code" gorm:"->"`
	TargetUOMName string `json:"target_uom_name" gorm:"->"`
}

func (MaterialMeasurement) TableName() string { return "mm_measurements" }

// MaterialDetail is returned by GET /:id
type MaterialDetail struct {
	Material
	CategoryCode string `json:"category_code" gorm:"->"`
	CategoryName string `json:"category_name" gorm:"->"`
}
