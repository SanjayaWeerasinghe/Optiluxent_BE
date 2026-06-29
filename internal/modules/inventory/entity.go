package inventory

import "time"

// ── Material Request ──────────────────────────────────────────────────────────

const (
	MRStatusDraft     = "DRAFT"
	MRStatusApproved  = "APPROVED"
	MRStatusRejected  = "REJECTED"
	MRStatusFulfilled = "FULFILLED"
	MRStatusPartial   = "PARTIAL"
)

type MaterialRequest struct {
	ID           uint       `json:"id"            gorm:"primaryKey"`
	TenantID     uint       `json:"tenant_id"     gorm:"not null;index"`
	Code         string     `json:"code"          gorm:"not null;size:50"`
	RequestedBy  uint       `json:"requested_by"  gorm:"not null"`
	DepartmentID *uint      `json:"department_id"`
	WarehouseID  uint       `json:"warehouse_id"  gorm:"not null"`
	NeededDate   string     `json:"needed_date"   gorm:"type:date;not null"`
	Status       string     `json:"status"        gorm:"not null;size:20;default:DRAFT"`
	Notes        string     `json:"notes"`
	ApprovedBy   *uint      `json:"approved_by"`
	ApprovedAt   *time.Time `json:"approved_at"`
	RejectedBy   *uint      `json:"rejected_by"`
	RejectedAt   *time.Time `json:"rejected_at"`
	RejectReason string     `json:"reject_reason"`
	CreatedBy    uint       `json:"created_by"    gorm:"not null"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Lines        []MRLine   `json:"lines,omitempty" gorm:"foreignKey:MRID"`
}

func (MaterialRequest) TableName() string { return "material_requests" }

type MRLine struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null"`
	MRID         uint      `json:"mr_id"         gorm:"not null;index"`
	LineNumber   int       `json:"line_number"   gorm:"not null;default:1"`
	ProductID    uint      `json:"product_id"    gorm:"not null"`
	VariantID    *uint     `json:"variant_id"`
	UOMId        uint      `json:"uom_id"        gorm:"not null"`
	RequestedQty float64   `json:"requested_qty" gorm:"type:decimal(18,4);not null"`
	IssuedQty    float64   `json:"issued_qty"    gorm:"type:decimal(18,4);not null;default:0"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (MRLine) TableName() string { return "material_request_lines" }

// ── Goods Transfer ────────────────────────────────────────────────────────────

const (
	GTStatusDraft     = "DRAFT"
	GTStatusInTransit = "IN_TRANSIT"
	GTStatusReceived  = "RECEIVED"
	GTStatusCancelled = "CANCELLED"
)

type GoodsTransfer struct {
	ID              uint       `json:"id"               gorm:"primaryKey"`
	TenantID        uint       `json:"tenant_id"        gorm:"not null;index"`
	Code            string     `json:"code"             gorm:"not null;size:50"`
	FromWarehouseID uint       `json:"from_warehouse_id" gorm:"not null"`
	ToWarehouseID   uint       `json:"to_warehouse_id"  gorm:"not null"`
	TransferDate    string     `json:"transfer_date"    gorm:"type:date;not null"`
	Status          string     `json:"status"           gorm:"not null;size:20;default:DRAFT"`
	Notes           string     `json:"notes"`
	SentBy          *uint      `json:"sent_by"`
	SentAt          *time.Time `json:"sent_at"`
	ReceivedBy      *uint      `json:"received_by"`
	ReceivedAt      *time.Time `json:"received_at"`
	CreatedBy       uint       `json:"created_by"       gorm:"not null"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Lines           []GTLine   `json:"lines,omitempty"  gorm:"foreignKey:TransferID"`
}

func (GoodsTransfer) TableName() string { return "goods_transfers" }

type GTLine struct {
	ID             uint      `json:"id"              gorm:"primaryKey"`
	TenantID       uint      `json:"tenant_id"       gorm:"not null"`
	TransferID     uint      `json:"transfer_id"     gorm:"not null;index"`
	LineNumber     int       `json:"line_number"     gorm:"not null;default:1"`
	ProductID      uint      `json:"product_id"      gorm:"not null"`
	VariantID      *uint     `json:"variant_id"`
	FromLocationID *uint     `json:"from_location_id"`
	ToLocationID   *uint     `json:"to_location_id"`
	UOMId          uint      `json:"uom_id"          gorm:"not null"`
	Quantity       float64   `json:"quantity"        gorm:"type:decimal(18,4);not null"`
	ReceivedQty    float64   `json:"received_qty"    gorm:"type:decimal(18,4);not null;default:0"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (GTLine) TableName() string { return "goods_transfer_lines" }

// ── Goods Issue ───────────────────────────────────────────────────────────────

const (
	GIStatusDraft     = "DRAFT"
	GIStatusConfirmed = "CONFIRMED"

	GIReasonProduction  = "PRODUCTION"
	GIReasonSale        = "SALE"
	GIReasonExpense     = "EXPENSE"
	GIReasonDamage      = "DAMAGE"
	GIReasonAdjustment  = "ADJUSTMENT"
	GIReasonQCRejection = "QC_REJECTION"
	GIReasonOther       = "OTHER"
)

type GoodsIssue struct {
	ID            uint       `json:"id"             gorm:"primaryKey"`
	TenantID      uint       `json:"tenant_id"      gorm:"not null;index"`
	Code          string     `json:"code"           gorm:"not null;size:50"`
	IssueDate     string     `json:"issue_date"     gorm:"type:date;not null"`
	WarehouseID   uint       `json:"warehouse_id"   gorm:"not null"`
	IssueReason   string     `json:"issue_reason"   gorm:"not null;size:50"`
	ReferenceType string     `json:"reference_type" gorm:"size:50"`
	ReferenceID   *uint      `json:"reference_id"`
	Status        string     `json:"status"         gorm:"not null;size:20;default:DRAFT"`
	Notes         string     `json:"notes"`
	CreatedBy     uint       `json:"created_by"     gorm:"not null"`
	ConfirmedBy   *uint      `json:"confirmed_by"`
	ConfirmedAt   *time.Time `json:"confirmed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Lines         []GILine   `json:"lines,omitempty" gorm:"foreignKey:IssueID"`
}

func (GoodsIssue) TableName() string { return "goods_issues" }

type GILine struct {
	ID         uint      `json:"id"          gorm:"primaryKey"`
	TenantID   uint      `json:"tenant_id"   gorm:"not null"`
	IssueID    uint      `json:"issue_id"    gorm:"not null;index"`
	LineNumber int       `json:"line_number" gorm:"not null;default:1"`
	ProductID  uint      `json:"product_id"  gorm:"not null"`
	VariantID  *uint     `json:"variant_id"`
	LocationID *uint     `json:"location_id"`
	UOMId      uint      `json:"uom_id"      gorm:"not null"`
	Quantity   float64   `json:"quantity"    gorm:"type:decimal(18,4);not null"`
	UnitCost   float64   `json:"unit_cost"   gorm:"type:decimal(18,4);not null;default:0"`
	TotalCost  float64   `json:"total_cost"  gorm:"type:decimal(18,4);not null;default:0"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (GILine) TableName() string { return "goods_issue_lines" }

// ── Stock Adjustment ──────────────────────────────────────────────────────────

const (
	SAStatusDraft     = "DRAFT"
	SAStatusConfirmed = "CONFIRMED"

	SAReasonStocktake  = "STOCKTAKE"
	SAReasonDamage     = "DAMAGE"
	SAReasonExpiry     = "EXPIRY"
	SAReasonWriteOff   = "WRITE_OFF"
	SAReasonCorrection = "CORRECTION"
	SAReasonOther      = "OTHER"
)

type StockAdjustment struct {
	ID             uint       `json:"id"              gorm:"primaryKey"`
	TenantID       uint       `json:"tenant_id"       gorm:"not null;index"`
	Code           string     `json:"code"            gorm:"not null;size:50"`
	AdjustmentDate string     `json:"adjustment_date" gorm:"type:date;not null"`
	WarehouseID    uint       `json:"warehouse_id"    gorm:"not null"`
	AdjustReason   string     `json:"adjust_reason"   gorm:"not null;size:50"`
	Status         string     `json:"status"          gorm:"not null;size:20;default:DRAFT"`
	Notes          string     `json:"notes"`
	CreatedBy      uint       `json:"created_by"      gorm:"not null"`
	ConfirmedBy    *uint      `json:"confirmed_by"`
	ConfirmedAt    *time.Time `json:"confirmed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Lines          []SALine   `json:"lines,omitempty" gorm:"foreignKey:AdjustmentID"`
}

func (StockAdjustment) TableName() string { return "stock_adjustments" }

type SALine struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null"`
	AdjustmentID uint      `json:"adjustment_id" gorm:"not null;index"`
	LineNumber   int       `json:"line_number"   gorm:"not null;default:1"`
	ProductID    uint      `json:"product_id"    gorm:"not null"`
	VariantID    *uint     `json:"variant_id"`
	LocationID   *uint     `json:"location_id"`
	UOMId        uint      `json:"uom_id"        gorm:"not null"`
	Quantity     float64   `json:"quantity"      gorm:"type:decimal(18,4);not null"`
	UnitCost     float64   `json:"unit_cost"     gorm:"type:decimal(18,4);not null;default:0"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SALine) TableName() string { return "stock_adjustment_lines" }

// ── Quality Check ─────────────────────────────────────────────────────────────

const (
	QCStatusPending    = "PENDING"
	QCStatusInProgress = "IN_PROGRESS"
	QCStatusPassed     = "PASSED"
	QCStatusFailed     = "FAILED"
	QCStatusPartial    = "PARTIAL"
)

// QC types — distinguish inspection of inbound goods vs production output.
const (
	QCTypeMaterial = "MATERIAL_QC" // inbound goods inspection (linked to GRN)
	QCTypeProduct  = "PRODUCT_QC"  // production output inspection (linked to Production Output)
)

type QualityCheck struct {
	ID            uint      `json:"id"             gorm:"primaryKey"`
	TenantID      uint      `json:"tenant_id"      gorm:"not null;index"`
	Code          string    `json:"code"           gorm:"not null;size:50"`
	QCType        string    `json:"qc_type"        gorm:"not null;size:30;default:MATERIAL_QC"`
	ReferenceType string    `json:"reference_type" gorm:"size:50"`
	ReferenceID   *uint     `json:"reference_id"`
	WarehouseID   uint      `json:"warehouse_id"   gorm:"not null"`
	CheckDate     string    `json:"check_date"     gorm:"type:date;not null"`
	Status        string    `json:"status"         gorm:"not null;size:20;default:PENDING"`
	InspectorID   *uint     `json:"inspector_id"`
	Notes         string    `json:"notes"`
	CreatedBy     uint      `json:"created_by"     gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Lines         []QCLine  `json:"lines,omitempty" gorm:"foreignKey:CheckID"`
}

func (QualityCheck) TableName() string { return "quality_checks" }

type QCLine struct {
	ID              uint      `json:"id"               gorm:"primaryKey"`
	TenantID        uint      `json:"tenant_id"        gorm:"not null"`
	CheckID         uint      `json:"check_id"         gorm:"not null;index"`
	LineNumber      int       `json:"line_number"      gorm:"not null;default:1"`
	ProductID       uint      `json:"product_id"       gorm:"not null"`
	VariantID       *uint     `json:"variant_id"`
	QtyChecked      float64   `json:"qty_checked"      gorm:"type:decimal(18,4);not null"`
	QtyPassed       float64   `json:"qty_passed"       gorm:"type:decimal(18,4);not null;default:0"`
	QtyFailed       float64   `json:"qty_failed"       gorm:"type:decimal(18,4);not null;default:0"`
	Result          string    `json:"result"           gorm:"not null;size:20;default:PENDING"`
	RejectionReason string    `json:"rejection_reason"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (QCLine) TableName() string { return "quality_check_lines" }

// ── Stock Balance (read-only, API response) ───────────────────────────────────

type StockBalanceRow struct {
	ProductID   uint    `json:"product_id"`
	VariantID   *uint   `json:"variant_id"`
	WarehouseID uint    `json:"warehouse_id"`
	LocationID  *uint   `json:"location_id"`
	Quantity    float64 `json:"quantity"`
	ReservedQty float64 `json:"reserved_qty"`
}
