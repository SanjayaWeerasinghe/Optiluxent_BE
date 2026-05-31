package mrp

type CreateMaterialMasterRequest struct {
	ProductID           uint     `json:"product_id"           validate:"required"`
	MaterialType        string   `json:"material_type"        validate:"omitempty,oneof=RAW_MATERIAL SEMI_FINISHED SERVICE"`
	MRPType             string   `json:"mrp_type"             validate:"omitempty,oneof=MRP REORDER_POINT MANUAL NO_PLANNING"`
	ProcurementType     string   `json:"procurement_type"     validate:"omitempty,oneof=EXTERNAL IN_HOUSE BOTH"`
	PurchaseLeadTime    int      `json:"purchase_lead_time"   validate:"min=0"`
	ProductionLeadTime  int      `json:"production_lead_time" validate:"min=0"`
	SafetyStock         float64  `json:"safety_stock"         validate:"min=0"`
	ReorderPoint        float64  `json:"reorder_point"        validate:"min=0"`
	MinOrderQty         float64  `json:"min_order_qty"        validate:"min=0"`
	MaxOrderQty         *float64 `json:"max_order_qty"        validate:"omitempty,min=0"`
	LotSizeType         string   `json:"lot_size_type"        validate:"omitempty,oneof=LOT_FOR_LOT FIXED_LOT ECONOMIC_ORDER"`
	FixedLotSize        *float64 `json:"fixed_lot_size"       validate:"omitempty,min=0"`
	ABCClass            string   `json:"abc_class"            validate:"omitempty,oneof=A B C"`
	ShelfLifeDays       *int     `json:"shelf_life_days"      validate:"omitempty,min=0"`
	IsActive            *bool    `json:"is_active"`
	Notes               string   `json:"notes"`
}

type UpdateMaterialMasterRequest struct {
	MaterialType        string   `json:"material_type"        validate:"omitempty,oneof=RAW_MATERIAL SEMI_FINISHED SERVICE"`
	MRPType             string   `json:"mrp_type"             validate:"omitempty,oneof=MRP REORDER_POINT MANUAL NO_PLANNING"`
	ProcurementType     string   `json:"procurement_type"     validate:"omitempty,oneof=EXTERNAL IN_HOUSE BOTH"`
	PurchaseLeadTime    *int     `json:"purchase_lead_time"   validate:"omitempty,min=0"`
	ProductionLeadTime  *int     `json:"production_lead_time" validate:"omitempty,min=0"`
	SafetyStock         *float64 `json:"safety_stock"         validate:"omitempty,min=0"`
	ReorderPoint        *float64 `json:"reorder_point"        validate:"omitempty,min=0"`
	MinOrderQty         *float64 `json:"min_order_qty"        validate:"omitempty,min=0"`
	MaxOrderQty         *float64 `json:"max_order_qty"        validate:"omitempty,min=0"`
	LotSizeType         string   `json:"lot_size_type"        validate:"omitempty,oneof=LOT_FOR_LOT FIXED_LOT ECONOMIC_ORDER"`
	FixedLotSize        *float64 `json:"fixed_lot_size"       validate:"omitempty,min=0"`
	ABCClass            string   `json:"abc_class"            validate:"omitempty,oneof=A B C"`
	ShelfLifeDays       *int     `json:"shelf_life_days"      validate:"omitempty,min=0"`
	IsActive            *bool    `json:"is_active"`
	Notes               string   `json:"notes"`
}
