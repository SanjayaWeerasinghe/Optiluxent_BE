package inventory

// ── Material Request DTOs ─────────────────────────────────────────────────────

type CreateMRRequest struct {
	RequestedBy  uint   `json:"requested_by"  validate:"required"`
	DepartmentID *uint  `json:"department_id"`
	WarehouseID  uint   `json:"warehouse_id"  validate:"required"`
	NeededDate   string `json:"needed_date"   validate:"omitempty"`
	Notes        string `json:"notes"`
}

type UpdateMRRequest struct {
	DepartmentID *uint  `json:"department_id"`
	WarehouseID  uint   `json:"warehouse_id"  validate:"required"`
	NeededDate   string `json:"needed_date"   validate:"omitempty"`
	Notes        string `json:"notes"`
}

type RejectMRRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type AddMRLineRequest struct {
	ProductID    uint    `json:"product_id"    validate:"required"`
	VariantID    *uint   `json:"variant_id"`
	UOMId        uint    `json:"uom_id"        validate:"required"`
	RequestedQty float64 `json:"requested_qty" validate:"required,min=0"`
	Notes        string  `json:"notes"`
}

type UpdateMRLineRequest struct {
	ProductID    uint    `json:"product_id"    validate:"required"`
	VariantID    *uint   `json:"variant_id"`
	UOMId        uint    `json:"uom_id"        validate:"required"`
	RequestedQty float64 `json:"requested_qty" validate:"required,min=0"`
	Notes        string  `json:"notes"`
}

// ── Goods Transfer DTOs ───────────────────────────────────────────────────────

type CreateTransferRequest struct {
	FromWarehouseID uint   `json:"from_warehouse_id" validate:"required"`
	ToWarehouseID   uint   `json:"to_warehouse_id"   validate:"required"`
	TransferDate    string `json:"transfer_date"     validate:"omitempty"`
	Notes           string `json:"notes"`
}

type UpdateTransferRequest struct {
	FromWarehouseID uint   `json:"from_warehouse_id" validate:"required"`
	ToWarehouseID   uint   `json:"to_warehouse_id"   validate:"required"`
	TransferDate    string `json:"transfer_date"     validate:"omitempty"`
	Notes           string `json:"notes"`
}

type AddTransferLineRequest struct {
	ProductID      uint    `json:"product_id"       validate:"required"`
	VariantID      *uint   `json:"variant_id"`
	FromLocationID *uint   `json:"from_location_id"`
	ToLocationID   *uint   `json:"to_location_id"`
	UOMId          uint    `json:"uom_id"           validate:"required"`
	Quantity       float64 `json:"quantity"         validate:"required,min=0"`
	Notes          string  `json:"notes"`
}

type UpdateTransferLineRequest struct {
	ProductID      uint    `json:"product_id"       validate:"required"`
	VariantID      *uint   `json:"variant_id"`
	FromLocationID *uint   `json:"from_location_id"`
	ToLocationID   *uint   `json:"to_location_id"`
	UOMId          uint    `json:"uom_id"           validate:"required"`
	Quantity       float64 `json:"quantity"         validate:"required,min=0"`
	Notes          string  `json:"notes"`
}

// ── Goods Issue DTOs ──────────────────────────────────────────────────────────

type CreateIssueRequest struct {
	IssueDate     string `json:"issue_date"     validate:"omitempty"`
	WarehouseID   uint   `json:"warehouse_id"   validate:"required"`
	IssueReason   string `json:"issue_reason"   validate:"required"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   *uint  `json:"reference_id"`
	Notes         string `json:"notes"`
}

type UpdateIssueRequest struct {
	IssueDate     string `json:"issue_date"     validate:"omitempty"`
	WarehouseID   uint   `json:"warehouse_id"   validate:"required"`
	IssueReason   string `json:"issue_reason"   validate:"required"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   *uint  `json:"reference_id"`
	Notes         string `json:"notes"`
}

type AddIssueLineRequest struct {
	ProductID  uint    `json:"product_id"  validate:"required"`
	VariantID  *uint   `json:"variant_id"`
	LocationID *uint   `json:"location_id"`
	UOMId      uint    `json:"uom_id"      validate:"required"`
	Quantity   float64 `json:"quantity"    validate:"required,min=0"`
	UnitCost   float64 `json:"unit_cost"   validate:"omitempty,min=0"`
	Notes      string  `json:"notes"`
}

type UpdateIssueLineRequest struct {
	ProductID  uint    `json:"product_id"  validate:"required"`
	VariantID  *uint   `json:"variant_id"`
	LocationID *uint   `json:"location_id"`
	UOMId      uint    `json:"uom_id"      validate:"required"`
	Quantity   float64 `json:"quantity"    validate:"required,min=0"`
	UnitCost   float64 `json:"unit_cost"   validate:"omitempty,min=0"`
	Notes      string  `json:"notes"`
}

// ── Stock Adjustment DTOs ─────────────────────────────────────────────────────

type CreateAdjustmentRequest struct {
	AdjustmentDate string `json:"adjustment_date" validate:"omitempty"`
	WarehouseID    uint   `json:"warehouse_id"    validate:"required"`
	AdjustReason   string `json:"adjust_reason"   validate:"required"`
	Notes          string `json:"notes"`
}

type UpdateAdjustmentRequest struct {
	AdjustmentDate string `json:"adjustment_date" validate:"omitempty"`
	WarehouseID    uint   `json:"warehouse_id"    validate:"required"`
	AdjustReason   string `json:"adjust_reason"   validate:"required"`
	Notes          string `json:"notes"`
}

type AddAdjustmentLineRequest struct {
	ProductID  uint    `json:"product_id"  validate:"required"`
	VariantID  *uint   `json:"variant_id"`
	LocationID *uint   `json:"location_id"`
	UOMId      uint    `json:"uom_id"      validate:"required"`
	Quantity   float64 `json:"quantity"    validate:"required"`
	UnitCost   float64 `json:"unit_cost"   validate:"omitempty,min=0"`
	Notes      string  `json:"notes"`
}

type UpdateAdjustmentLineRequest struct {
	ProductID  uint    `json:"product_id"  validate:"required"`
	VariantID  *uint   `json:"variant_id"`
	LocationID *uint   `json:"location_id"`
	UOMId      uint    `json:"uom_id"      validate:"required"`
	Quantity   float64 `json:"quantity"    validate:"required"`
	UnitCost   float64 `json:"unit_cost"   validate:"omitempty,min=0"`
	Notes      string  `json:"notes"`
}

// ── Quality Check DTOs ────────────────────────────────────────────────────────

type CreateQualityCheckRequest struct {
	ReferenceType string `json:"reference_type"`
	ReferenceID   *uint  `json:"reference_id"`
	WarehouseID   uint   `json:"warehouse_id"  validate:"required"`
	CheckDate     string `json:"check_date"    validate:"omitempty"`
	InspectorID   *uint  `json:"inspector_id"`
	Notes         string `json:"notes"`
}

type AddQCLineRequest struct {
	ProductID       uint    `json:"product_id"       validate:"required"`
	VariantID       *uint   `json:"variant_id"`
	QtyChecked      float64 `json:"qty_checked"      validate:"required,min=0"`
	QtyPassed       float64 `json:"qty_passed"       validate:"omitempty,min=0"`
	QtyFailed       float64 `json:"qty_failed"       validate:"omitempty,min=0"`
	Result          string  `json:"result"           validate:"omitempty"`
	RejectionReason string  `json:"rejection_reason"`
	Notes           string  `json:"notes"`
}

type UpdateQCLineRequest struct {
	ProductID       uint    `json:"product_id"       validate:"required"`
	VariantID       *uint   `json:"variant_id"`
	QtyChecked      float64 `json:"qty_checked"      validate:"required,min=0"`
	QtyPassed       float64 `json:"qty_passed"       validate:"omitempty,min=0"`
	QtyFailed       float64 `json:"qty_failed"       validate:"omitempty,min=0"`
	Result          string  `json:"result"           validate:"omitempty"`
	RejectionReason string  `json:"rejection_reason"`
	Notes           string  `json:"notes"`
}
