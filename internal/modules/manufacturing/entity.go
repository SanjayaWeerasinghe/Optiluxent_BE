package manufacturing

import (
	"time"

	"gorm.io/gorm"
)

// ── Cost Estimate (Pre-Costing) ───────────────────────────────────────────────

const (
	EstimateStatusDraft     = "DRAFT"
	EstimateStatusApproved  = "APPROVED"
	EstimateStatusCancelled = "CANCELLED"
)

type CostEstimate struct {
	ID                     uint               `json:"id"                       gorm:"primaryKey"`
	TenantID               uint               `json:"tenant_id"                gorm:"not null;index"`
	Code                   string             `json:"code"                     gorm:"not null;size:50"`
	ProductID              uint               `json:"product_id"               gorm:"not null"`
	UOMID                  uint               `json:"uom_id"                   gorm:"not null"`
	PlannedQty             float64            `json:"planned_qty"`
	BOMID                  *uint              `json:"bom_id"`
	RoutingID              *uint              `json:"routing_id"`
	PlannedStartDate       *string            `json:"planned_start_date"       gorm:"type:date"`
	PlannedEndDate         *string            `json:"planned_end_date"         gorm:"type:date"`
	EstimatedMaterialCost  float64            `json:"estimated_material_cost"  gorm:"not null;default:0"`
	EstimatedResourceCost  float64            `json:"estimated_resource_cost"  gorm:"not null;default:0"`
	EstimatedOverheadCost  float64            `json:"estimated_overhead_cost"  gorm:"not null;default:0"`
	TotalEstimatedCost     float64            `json:"total_estimated_cost"     gorm:"not null;default:0"`
	Status                 string             `json:"status"                   gorm:"not null;size:20;default:DRAFT"`
	Notes                  string             `json:"notes"`
	ApprovedBy             *uint              `json:"approved_by"`
	ApprovedAt             *time.Time         `json:"approved_at"`
	CreatedBy              *uint              `json:"created_by"`
	Lines                  []CostEstimateLine `json:"lines,omitempty"          gorm:"foreignKey:EstimateID"`
	CreatedAt              time.Time          `json:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at"`
	DeletedAt              gorm.DeletedAt     `json:"deleted_at,omitempty"     gorm:"index"`
}

func (CostEstimate) TableName() string { return "cost_estimates" }

const (
	EstimateLineTypeMaterial = "MATERIAL"
	EstimateLineTypeResource = "RESOURCE"
	EstimateLineTypeLabor    = "LABOR"
	EstimateLineTypeOverhead = "OVERHEAD"
)

type CostEstimateLine struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	EstimateID   uint      `json:"estimate_id"   gorm:"not null;index"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null"`
	LineNumber   int       `json:"line_number"   gorm:"not null;default:1"`
	LineType     string    `json:"line_type"     gorm:"not null;size:20"`
	ProductID    *uint     `json:"product_id"`
	WorkCenterID *uint     `json:"work_center_id"`
	Description  string    `json:"description"   gorm:"size:500"`
	Quantity     float64   `json:"quantity"`
	UOMID        *uint     `json:"uom_id"`
	UnitCost     float64   `json:"unit_cost"     gorm:"not null;default:0"`
	TotalCost    float64   `json:"total_cost"    gorm:"not null;default:0"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (CostEstimateLine) TableName() string { return "cost_estimate_lines" }

// ── Production Plan ───────────────────────────────────────────────────────────

const (
	PlanStatusDraft      = "DRAFT"
	PlanStatusReleased   = "RELEASED"
	PlanStatusInProgress = "IN_PROGRESS"
	PlanStatusCompleted  = "COMPLETED"
	PlanStatusCancelled  = "CANCELLED"
)

type ProductionPlan struct {
	ID               uint           `json:"id"                 gorm:"primaryKey"`
	TenantID         uint           `json:"tenant_id"          gorm:"not null;index"`
	Code             string         `json:"code"               gorm:"not null;size:50"`
	// SOID links this Plan to a Sales Order — populated for Refinery Service
	// Plans so the resulting Production carries the customer/SO context.
	SOID             *uint          `json:"so_id"`
	// DocumentTypeID classifies the Plan (e.g. REFINING_PLAN system_key).
	DocumentTypeID   *uint          `json:"document_type_id"`
	ProductID        uint           `json:"product_id"         gorm:"not null"`
	UOMID            uint           `json:"uom_id"             gorm:"not null"`
	PlannedQty       float64        `json:"planned_qty"`
	BOMID            *uint          `json:"bom_id"`
	RoutingID        *uint          `json:"routing_id"`
	EstimateID       *uint          `json:"estimate_id"`
	WarehouseID      uint           `json:"warehouse_id"       gorm:"not null"`
	PlannedStartDate *string        `json:"planned_start_date" gorm:"type:date"`
	PlannedEndDate   *string        `json:"planned_end_date"   gorm:"type:date"`
	Status           string         `json:"status"             gorm:"not null;size:20;default:DRAFT"`
	Notes            string         `json:"notes"`
	ReleasedBy       *uint          `json:"released_by"`
	ReleasedAt       *time.Time     `json:"released_at"`
	CreatedBy        *uint          `json:"created_by"`
	Inputs           []ProductionPlanInput `json:"inputs,omitempty" gorm:"foreignKey:PlanID"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (ProductionPlan) TableName() string { return "production_plans" }

// ProductionPlanInput lists the chemicals / materials / other resources the
// Plan anticipates consuming. Copied into MO.ProductionResource rows when a
// Production is manually created from the released Plan.
type ProductionPlanInput struct {
	ID         uint      `json:"id"          gorm:"primaryKey"`
	PlanID     uint      `json:"plan_id"     gorm:"not null;index"`
	TenantID   uint      `json:"tenant_id"   gorm:"not null;index"`
	LineNumber int       `json:"line_number" gorm:"not null;default:1"`
	ProductID  uint      `json:"product_id"  gorm:"not null"`
	Quantity   float64   `json:"quantity"    gorm:"not null;default:0"`
	UOMID      uint      `json:"uom_id"      gorm:"not null"`
	Notes      string    `json:"notes"       gorm:"size:500"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (ProductionPlanInput) TableName() string { return "production_plan_inputs" }

// ── Production Order ──────────────────────────────────────────────────────────

const (
	OrderStatusDraft      = "DRAFT"
	OrderStatusInProgress = "IN_PROGRESS"
	OrderStatusCompleted  = "COMPLETED"
	OrderStatusCancelled  = "CANCELLED"
)

type ProductionOrder struct {
	ID          uint               `json:"id"           gorm:"primaryKey"`
	TenantID    uint               `json:"tenant_id"    gorm:"not null;index"`
	Code        string             `json:"code"         gorm:"not null;size:50"`
	PlanID      *uint              `json:"plan_id"`
	SOID        *uint              `json:"so_id"` // Sales-order-based MOs; mutually exclusive with plan_id in the UI.
	ProductID   uint               `json:"product_id"   gorm:"not null"`
	UOMID       uint               `json:"uom_id"       gorm:"not null"`
	PlannedQty  float64            `json:"planned_qty"`
	ProducedQty float64            `json:"produced_qty" gorm:"not null;default:0"`
	WarehouseID uint               `json:"warehouse_id" gorm:"not null"`
	StartDate   *string            `json:"start_date"   gorm:"type:date"`
	EndDate     *string            `json:"end_date"     gorm:"type:date"`
	Status      string             `json:"status"       gorm:"not null;size:20;default:DRAFT"`
	Notes       string             `json:"notes"`
	CompletedBy *uint              `json:"completed_by"`
	CompletedAt *time.Time         `json:"completed_at"`
	CreatedBy   *uint              `json:"created_by"`
	Outputs     []ProductionOutput   `json:"outputs,omitempty"   gorm:"foreignKey:OrderID"`
	Resources   []ProductionResource `json:"resources,omitempty" gorm:"foreignKey:OrderID"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`
}

func (ProductionOrder) TableName() string { return "production_orders" }

type ProductionOutput struct {
	ID          uint      `json:"id"          gorm:"primaryKey"`
	OrderID     uint      `json:"order_id"    gorm:"not null;index"`
	TenantID    uint      `json:"tenant_id"   gorm:"not null"`
	LineNumber  int       `json:"line_number" gorm:"not null;default:1"`
	ProductID   uint      `json:"product_id"  gorm:"not null"`
	UOMID       uint      `json:"uom_id"      gorm:"not null"`
	Quantity    float64   `json:"quantity"`
	UnitCost    float64   `json:"unit_cost"   gorm:"not null;default:0"`
	TotalCost   float64   `json:"total_cost"  gorm:"not null;default:0"`
	WarehouseID *uint     `json:"warehouse_id"`
	LocationID  *uint     `json:"location_id"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ProductionOutput) TableName() string { return "production_outputs" }

const (
	ResourceTypeMaterial = "MATERIAL"
	ResourceTypeLabor    = "LABOR"
	ResourceTypeOverhead = "OVERHEAD"
)

type ProductionResource struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	OrderID      uint      `json:"order_id"      gorm:"not null;index"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null"`
	LineNumber   int       `json:"line_number"   gorm:"not null;default:1"`
	ResourceType string    `json:"resource_type" gorm:"not null;size:20"`
	ProductID    *uint     `json:"product_id"`
	WorkCenterID *uint     `json:"work_center_id"`
	Description  string    `json:"description"   gorm:"size:500"`
	Quantity     float64   `json:"quantity"`
	UOMID        *uint     `json:"uom_id"`
	UnitCost     float64   `json:"unit_cost"     gorm:"not null;default:0"`
	TotalCost    float64   `json:"total_cost"    gorm:"not null;default:0"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (ProductionResource) TableName() string { return "production_resources" }

// ── Post-Costing ──────────────────────────────────────────────────────────────

const (
	PostCostStatusDraft     = "DRAFT"
	PostCostStatusFinalized = "FINALIZED"
)

type PostCost struct {
	ID                     uint           `json:"id"                       gorm:"primaryKey"`
	TenantID               uint           `json:"tenant_id"                gorm:"not null;index"`
	Code                   string         `json:"code"                     gorm:"not null;size:50"`
	OrderID                uint           `json:"order_id"                 gorm:"not null"`
	EstimateID             *uint          `json:"estimate_id"`
	ActualMaterialCost     float64        `json:"actual_material_cost"     gorm:"not null;default:0"`
	ActualResourceCost     float64        `json:"actual_resource_cost"     gorm:"not null;default:0"`
	ActualOverheadCost     float64        `json:"actual_overhead_cost"     gorm:"not null;default:0"`
	TotalActualCost        float64        `json:"total_actual_cost"        gorm:"not null;default:0"`
	EstimatedMaterialCost  float64        `json:"estimated_material_cost"  gorm:"not null;default:0"`
	EstimatedResourceCost  float64        `json:"estimated_resource_cost"  gorm:"not null;default:0"`
	EstimatedOverheadCost  float64        `json:"estimated_overhead_cost"  gorm:"not null;default:0"`
	TotalEstimatedCost     float64        `json:"total_estimated_cost"     gorm:"not null;default:0"`
	VarianceAmount         float64        `json:"variance_amount"          gorm:"not null;default:0"`
	VariancePct            float64        `json:"variance_pct"             gorm:"not null;default:0"`
	Status                 string         `json:"status"                   gorm:"not null;size:20;default:DRAFT"`
	Notes                  string         `json:"notes"`
	FinalizedBy            *uint          `json:"finalized_by"`
	FinalizedAt            *time.Time     `json:"finalized_at"`
	CreatedBy              *uint          `json:"created_by"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (PostCost) TableName() string { return "post_costs" }
