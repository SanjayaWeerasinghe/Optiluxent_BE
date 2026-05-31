package materials

// ── Core Material ─────────────────────────────────────────────────────────────

type CreateMaterialRequest struct {
	Code         string `json:"code"          validate:"required,max=50"`
	Name         string `json:"name"          validate:"required,max=200"`
	MaterialType string `json:"material_type" validate:"omitempty,oneof=RAW_MATERIAL SEMI_FINISHED SERVICE"`
	Color        string `json:"color"         validate:"omitempty,max=100"`
	CategoryID   *uint  `json:"category_id"`
	ArticleCode  string `json:"article_code"  validate:"omitempty,max=100"`
	Ref2         string `json:"ref2"          validate:"omitempty,max=100"`
	Barcode      string `json:"barcode"       validate:"omitempty,max=100"`
	Description  string `json:"description"`
	Notes        string `json:"notes"`
}

type UpdateMaterialRequest struct {
	Name         string `json:"name"          validate:"omitempty,max=200"`
	MaterialType string `json:"material_type" validate:"omitempty,oneof=RAW_MATERIAL SEMI_FINISHED SERVICE"`
	Color        string `json:"color"         validate:"omitempty,max=100"`
	CategoryID   *uint  `json:"category_id"`
	ArticleCode  string `json:"article_code"  validate:"omitempty,max=100"`
	Ref2         string `json:"ref2"          validate:"omitempty,max=100"`
	Barcode      string `json:"barcode"       validate:"omitempty,max=100"`
	Description  string `json:"description"`
	IsActive     *bool  `json:"is_active"`
	Notes        string `json:"notes"`
}

// ── Purchasing ────────────────────────────────────────────────────────────────

type UpsertPurchasingRequest struct {
	PurchasingUOMID  *uint   `json:"purchasing_uom_id"`
	UnderDeliveryPct float64 `json:"under_delivery_pct"`
	OverDeliveryPct  float64 `json:"over_delivery_pct"`
}

// ── Manufacturing ─────────────────────────────────────────────────────────────

type UpsertManufacturingRequest struct {
	ProductionUOMID       *uint   `json:"production_uom_id"`
	ReorderQtyLevel       float64 `json:"reorder_qty_level"`
	SafetyLevel           float64 `json:"safety_level"`
	ProductionDays        int     `json:"production_days"`
	DeliveryDays          int     `json:"delivery_days"`
	GRNDays               int     `json:"grn_days"`
	ProcurementRepeatDays int     `json:"procurement_repeat_days"`
}

// ── Warehouse ─────────────────────────────────────────────────────────────────

type UpsertWarehouseRequest struct {
	StockingUOMID       *uint  `json:"stocking_uom_id"`
	StockRemoval        string `json:"stock_removal"         validate:"omitempty,oneof=FIFO FEFO LIFO"`
	StorageMain         bool   `json:"storage_main"`
	StorageDamaged      bool   `json:"storage_damaged"`
	StorageHold         bool   `json:"storage_hold"`
	BatchProcess        bool   `json:"batch_process"`
	ProductionDateCheck bool   `json:"production_date_check"`
	ExpiryDateCheck     bool   `json:"expiry_date_check"`
	QCCheck             bool   `json:"qc_check"`
	GRNWithPOUOMImage   bool   `json:"grn_with_po_uom_image"`
}

// ── Vendors ───────────────────────────────────────────────────────────────────

type VendorRequest struct {
	VendorID       *uint    `json:"vendor_id"`
	ArticleNo      string   `json:"article_no"`
	DeliveryDays   int      `json:"delivery_days"`
	Cost           *float64 `json:"cost"`
	CurrencyID     *uint    `json:"currency_id"`
	ProjectedPrice *float64 `json:"projected_price"`
	MOQ            *float64 `json:"moq"`
	IsDefault      bool     `json:"is_default"`
}

// ── Measurements ──────────────────────────────────────────────────────────────

type MeasurementRequest struct {
	BaseUOMID       uint    `json:"base_uom_id"      validate:"required"`
	TargetUOMID     uint    `json:"target_uom_id"    validate:"required"`
	ConversionRatio float64 `json:"conversion_ratio" validate:"required,gt=0"`
}
