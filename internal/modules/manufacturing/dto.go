package manufacturing

// ── Cost Estimate DTOs ────────────────────────────────────────────────────────

type CreateEstimateRequest struct {
	ProductID        uint    `json:"product_id"         validate:"required"`
	UOMID            uint    `json:"uom_id"             validate:"required"`
	PlannedQty       float64 `json:"planned_qty"        validate:"omitempty,min=0"`
	BOMID            *uint   `json:"bom_id"`
	RoutingID        *uint   `json:"routing_id"`
	PlannedStartDate *string `json:"planned_start_date"`
	PlannedEndDate   *string `json:"planned_end_date"`
	Notes            string  `json:"notes"`
}

type UpdateEstimateRequest struct {
	ProductID        uint    `json:"product_id"         validate:"required"`
	UOMID            uint    `json:"uom_id"             validate:"required"`
	PlannedQty       float64 `json:"planned_qty"        validate:"omitempty,min=0"`
	BOMID            *uint   `json:"bom_id"`
	RoutingID        *uint   `json:"routing_id"`
	PlannedStartDate *string `json:"planned_start_date"`
	PlannedEndDate   *string `json:"planned_end_date"`
	Notes            string  `json:"notes"`
}

type AddEstimateLineRequest struct {
	LineType     string  `json:"line_type"     validate:"required"`
	ProductID    *uint   `json:"product_id"`
	WorkCenterID *uint   `json:"work_center_id"`
	Description  string  `json:"description"   validate:"max=500"`
	Quantity     float64 `json:"quantity"      validate:"omitempty,min=0"`
	UOMID        *uint   `json:"uom_id"`
	UnitCost     float64 `json:"unit_cost"     validate:"omitempty,min=0"`
	Notes        string  `json:"notes"`
}

// ── Production Plan DTOs ──────────────────────────────────────────────────────

type CreatePlanRequest struct {
	SOID             *uint   `json:"so_id"`
	DocumentTypeID   *uint   `json:"document_type_id"`
	ProductID        uint    `json:"product_id"         validate:"required"`
	UOMID            uint    `json:"uom_id"             validate:"required"`
	PlannedQty       float64 `json:"planned_qty"        validate:"omitempty,min=0"`
	BOMID            *uint   `json:"bom_id"`
	RoutingID        *uint   `json:"routing_id"`
	EstimateID       *uint   `json:"estimate_id"`
	WarehouseID      uint    `json:"warehouse_id"       validate:"required"`
	PlannedStartDate *string `json:"planned_start_date"`
	PlannedEndDate   *string `json:"planned_end_date"`
	Notes            string  `json:"notes"`
}

type UpdatePlanRequest struct {
	SOID             *uint   `json:"so_id"`
	DocumentTypeID   *uint   `json:"document_type_id"`
	ProductID        uint    `json:"product_id"         validate:"required"`
	UOMID            uint    `json:"uom_id"             validate:"required"`
	PlannedQty       float64 `json:"planned_qty"        validate:"omitempty,min=0"`
	BOMID            *uint   `json:"bom_id"`
	RoutingID        *uint   `json:"routing_id"`
	EstimateID       *uint   `json:"estimate_id"`
	WarehouseID      uint    `json:"warehouse_id"       validate:"required"`
	PlannedStartDate *string `json:"planned_start_date"`
	PlannedEndDate   *string `json:"planned_end_date"`
	Notes            string  `json:"notes"`
}

// ── Production Plan Inputs ────────────────────────────────────────────────────

type AddPlanInputRequest struct {
	ProductID uint    `json:"product_id" validate:"required"`
	Quantity  float64 `json:"quantity"   validate:"required,min=0"`
	UOMID     uint    `json:"uom_id"     validate:"required"`
	Notes     string  `json:"notes"      validate:"omitempty,max=500"`
}

type UpdatePlanInputRequest struct {
	ProductID uint    `json:"product_id" validate:"required"`
	Quantity  float64 `json:"quantity"   validate:"required,min=0"`
	UOMID     uint    `json:"uom_id"     validate:"required"`
	Notes     string  `json:"notes"      validate:"omitempty,max=500"`
}

// ── Production Order DTOs ─────────────────────────────────────────────────────

type CreateOrderRequest struct {
	PlanID      *uint   `json:"plan_id"`
	SOID        *uint   `json:"so_id"`
	ProductID   uint    `json:"product_id"   validate:"required"`
	UOMID       uint    `json:"uom_id"       validate:"required"`
	PlannedQty  float64 `json:"planned_qty"  validate:"omitempty,min=0"`
	WarehouseID uint    `json:"warehouse_id" validate:"required"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Notes       string  `json:"notes"`
}

type UpdateOrderRequest struct {
	PlanID      *uint   `json:"plan_id"`
	SOID        *uint   `json:"so_id"`
	ProductID   uint    `json:"product_id"   validate:"required"`
	UOMID       uint    `json:"uom_id"       validate:"required"`
	PlannedQty  float64 `json:"planned_qty"  validate:"omitempty,min=0"`
	WarehouseID uint    `json:"warehouse_id" validate:"required"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Notes       string  `json:"notes"`
}

type AddOutputRequest struct {
	ProductID   uint    `json:"product_id"   validate:"required"`
	UOMID       uint    `json:"uom_id"       validate:"required"`
	Quantity    float64 `json:"quantity"     validate:"omitempty,min=0"`
	UnitCost    float64 `json:"unit_cost"    validate:"omitempty,min=0"`
	WarehouseID *uint   `json:"warehouse_id"`
	LocationID  *uint   `json:"location_id"`
	Notes       string  `json:"notes"`
}

type AddResourceRequest struct {
	ResourceType string  `json:"resource_type"  validate:"required"`
	ProductID    *uint   `json:"product_id"`
	WorkCenterID *uint   `json:"work_center_id"`
	Description  string  `json:"description"    validate:"max=500"`
	Quantity     float64 `json:"quantity"       validate:"omitempty,min=0"`
	UOMID        *uint   `json:"uom_id"`
	UnitCost     float64 `json:"unit_cost"      validate:"omitempty,min=0"`
	Notes        string  `json:"notes"`
}

// ── Post-Cost DTOs ────────────────────────────────────────────────────────────

type CreatePostCostRequest struct {
	OrderID    uint   `json:"order_id"    validate:"required"`
	EstimateID *uint  `json:"estimate_id"`
	Notes      string `json:"notes"`
}
