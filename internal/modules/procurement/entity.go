package procurement

import (
	"time"

	"gorm.io/gorm"
)

// ── Purchase Request ──────────────────────────────────────────────────────────

const (
	PRStatusDraft    = "DRAFT"
	PRStatusPending  = "PENDING_APPROVAL"
	PRStatusApproved = "APPROVED"
	PRStatusRejected = "REJECTED"
	PRStatusCancelled = "CANCELLED"
)

type PurchaseRequest struct {
	ID             uint           `json:"id"            gorm:"primaryKey"`
	TenantID       uint           `json:"tenant_id"     gorm:"not null;index"`
	Code           string         `json:"code"          gorm:"not null;size:50"`
	DocumentTypeID *uint          `json:"document_type_id"`
	RequestDate  string         `json:"request_date"  gorm:"type:date;not null"`
	RequiredDate *string        `json:"required_date" gorm:"type:date"`
	RequestedBy  *uint          `json:"requested_by"`
	DepartmentID *uint          `json:"department_id"`
	Status       string         `json:"status"        gorm:"not null;size:20;default:DRAFT"`
	Notes        string         `json:"notes"`
	ApprovedBy   *uint          `json:"approved_by"`
	ApprovedAt   *time.Time     `json:"approved_at"`
	CreatedBy    *uint          `json:"created_by"`
	Lines        []PRLine       `json:"lines,omitempty" gorm:"foreignKey:PRID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (PurchaseRequest) TableName() string { return "purchase_requests" }

type PRLine struct {
	ID             uint      `json:"id"              gorm:"primaryKey"`
	PRID           uint      `json:"pr_id"           gorm:"not null;index"`
	TenantID       uint      `json:"tenant_id"       gorm:"not null"`
	LineNumber     int       `json:"line_number"     gorm:"not null;default:1"`
	ProductID      uint      `json:"product_id"      gorm:"not null"`
	VariantID      *uint     `json:"variant_id"`
	Description    string    `json:"description"     gorm:"size:500"`
	Quantity       float64   `json:"quantity"        gorm:"not null"`
	UOMID          uint      `json:"uom_id"          gorm:"not null"`
	EstimatedPrice float64   `json:"estimated_price" gorm:"not null;default:0"`
	TransferRatio  float64   `json:"transfer_ratio"  gorm:"not null;default:1"`
	CurrencyID     *uint     `json:"currency_id"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (PRLine) TableName() string { return "purchase_request_lines" }

// ── Purchase Order ────────────────────────────────────────────────────────────

const (
	POStatusDraft     = "DRAFT"
	POStatusConfirmed = "CONFIRMED"
	POStatusPartial   = "PARTIAL"
	POStatusReceived  = "RECEIVED"
	POStatusCancelled = "CANCELLED"
)

type PurchaseOrder struct {
	ID             uint           `json:"id"              gorm:"primaryKey"`
	TenantID       uint           `json:"tenant_id"       gorm:"not null;index"`
	Code           string         `json:"code"            gorm:"not null;size:50"`
	DocumentTypeID *uint          `json:"document_type_id"`
	SupplierID     uint           `json:"supplier_id"     gorm:"not null;index"`
	PRID           *uint          `json:"pr_id"`
	OrderDate      string         `json:"order_date"      gorm:"type:date;not null"`
	ExpectedDate   *string        `json:"expected_date"   gorm:"type:date"`
	CurrencyID     uint           `json:"currency_id"     gorm:"not null"`
	ExchangeRate   float64        `json:"exchange_rate"   gorm:"not null;default:1"`
	PaymentTermID  *uint          `json:"payment_term_id"`
	WarehouseID    uint           `json:"warehouse_id"    gorm:"not null"`
	Status         string         `json:"status"          gorm:"not null;size:20;default:DRAFT"`
	Subtotal       float64        `json:"subtotal"        gorm:"not null;default:0"`
	TaxAmount      float64        `json:"tax_amount"      gorm:"not null;default:0"`
	DiscountAmount float64        `json:"discount_amount" gorm:"not null;default:0"`
	TotalAmount    float64        `json:"total_amount"    gorm:"not null;default:0"`
	Notes          string         `json:"notes"`
	ConfirmedBy    *uint          `json:"confirmed_by"`
	ConfirmedAt    *time.Time     `json:"confirmed_at"`
	CreatedBy      *uint          `json:"created_by"`
	Lines          []POLine       `json:"lines,omitempty" gorm:"foreignKey:POID"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

type POLine struct {
	ID            uint      `json:"id"             gorm:"primaryKey"`
	POID          uint      `json:"po_id"          gorm:"not null;index"`
	TenantID      uint      `json:"tenant_id"      gorm:"not null"`
	LineNumber    int       `json:"line_number"    gorm:"not null;default:1"`
	ProductID     uint      `json:"product_id"     gorm:"not null"`
	VariantID     *uint     `json:"variant_id"`
	Description   string    `json:"description"    gorm:"size:500"`
	Quantity      float64   `json:"quantity"       gorm:"not null"`
	UOMID         uint      `json:"uom_id"         gorm:"not null"`
	UnitPrice     float64   `json:"unit_price"     gorm:"not null;default:0"`
	DiscountPct   float64   `json:"discount_pct"   gorm:"not null;default:0"`
	TaxCodeID     *uint     `json:"tax_code_id"`
	TaxAmount     float64   `json:"tax_amount"     gorm:"not null;default:0"`
	LineTotal     float64   `json:"line_total"     gorm:"not null;default:0"`
	TransferRatio float64   `json:"transfer_ratio" gorm:"not null;default:1"`
	ReceivedQty   float64   `json:"received_qty"   gorm:"not null;default:0"`
	BilledQty     float64   `json:"billed_qty"     gorm:"not null;default:0"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (POLine) TableName() string { return "purchase_order_lines" }

// ── Goods Receipt ─────────────────────────────────────────────────────────────

const (
	GRNStatusDraft     = "DRAFT"
	GRNStatusConfirmed = "CONFIRMED"
	GRNStatusCancelled = "CANCELLED"
)

// GRN behavioural system_keys — these live on rows in `document_types`
// with system_key set. The BE branches on the resolved key.
// See migration 000033 for the seed.
const (
	GRNTypeWithPO            = "WITH_PO"
	GRNTypeWithoutPO         = "WITHOUT_PO"
	GRNTypeCustomerReturn    = "CUSTOMER_RETURN"
	GRNTypeProductionReturn  = "PRODUCTION_RETURN"
	GRNTypeProductionOutput  = "PRODUCTION_OUTPUT"
	// GRNTypeRefiningIntake — customer brings crude oil in for a Refinery
	// Service Production. Stock posts to a segregated pool keyed by the
	// linked Production ID and never enters our inventory valuation.
	GRNTypeRefiningIntake    = "REFINING_INTAKE"
)

// SOTypeRefineryService — seeded document_types.system_key for the sales-
// side "Refinery Service" classification. Present on the SO's DocumentTypeID
// when this is a customer refining engagement.
const SOTypeRefineryService = "REFINERY_SERVICE"

type GoodsReceipt struct {
	ID             uint           `json:"id"                gorm:"primaryKey"`
	TenantID       uint           `json:"tenant_id"         gorm:"not null;index"`
	Code           string         `json:"code"              gorm:"not null;size:50"`
	DocumentTypeID *uint          `json:"document_type_id"`
	POID           *uint          `json:"po_id"`
	MOID           *uint          `json:"mo_id"`
	SupplierID     uint           `json:"supplier_id"       gorm:"not null"`
	ReceiptDate string         `json:"receipt_date" gorm:"type:date;not null"`
	WarehouseID uint           `json:"warehouse_id" gorm:"not null"`
	Status      string         `json:"status"       gorm:"not null;size:20;default:DRAFT"`
	Notes       string         `json:"notes"`
	ConfirmedBy *uint          `json:"confirmed_by"`
	ConfirmedAt *time.Time     `json:"confirmed_at"`
	CreatedBy   *uint          `json:"created_by"`
	Lines       []GRNLine      `json:"lines,omitempty" gorm:"foreignKey:GRNID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (GoodsReceipt) TableName() string { return "goods_receipts" }

type GRNLine struct {
	ID            uint      `json:"id"             gorm:"primaryKey"`
	GRNID         uint      `json:"grn_id"         gorm:"not null;index"`
	TenantID      uint      `json:"tenant_id"      gorm:"not null"`
	POLineID      *uint     `json:"po_line_id"`
	LineNumber    int       `json:"line_number"    gorm:"not null;default:1"`
	ProductID     uint      `json:"product_id"     gorm:"not null"`
	VariantID     *uint     `json:"variant_id"`
	Quantity      float64   `json:"quantity"       gorm:"not null"`
	UOMID         uint      `json:"uom_id"         gorm:"not null"`
	LocationID    *uint     `json:"location_id"`
	UnitCost      float64   `json:"unit_cost"      gorm:"not null;default:0"`
	TotalCost     float64   `json:"total_cost"     gorm:"not null;default:0"`
	TransferRatio float64   `json:"transfer_ratio" gorm:"not null;default:1"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (GRNLine) TableName() string { return "goods_receipt_lines" }

// ── Purchase Invoice ──────────────────────────────────────────────────────────

const (
	InvStatusDraft     = "DRAFT"
	InvStatusPosted    = "POSTED"
	InvStatusPartial   = "PARTIAL"
	InvStatusPaid      = "PAID"
	InvStatusCancelled = "CANCELLED"
)

type PurchaseInvoice struct {
	ID                  uint           `json:"id"                    gorm:"primaryKey"`
	TenantID            uint           `json:"tenant_id"             gorm:"not null;index"`
	Code                string         `json:"code"                  gorm:"not null;size:50"`
	DocumentTypeID      *uint          `json:"document_type_id"`
	SupplierID          uint           `json:"supplier_id"           gorm:"not null;index"`
	POID                uint           `json:"po_id"                 gorm:"not null"`
	InvoiceDate         string         `json:"invoice_date"          gorm:"type:date;not null"`
	DueDate             *string        `json:"due_date"              gorm:"type:date"`
	SupplierInvoiceNo   string         `json:"supplier_invoice_no"   gorm:"size:100"`
	SupplierInvoiceDate *string        `json:"supplier_invoice_date" gorm:"type:date"`
	CurrencyID          uint           `json:"currency_id"           gorm:"not null"`
	ExchangeRate        float64        `json:"exchange_rate"         gorm:"not null;default:1"`
	PaymentTermID       *uint          `json:"payment_term_id"`
	Status              string         `json:"status"                gorm:"not null;size:20;default:DRAFT"`
	Subtotal            float64        `json:"subtotal"              gorm:"not null;default:0"`
	TaxAmount           float64        `json:"tax_amount"            gorm:"not null;default:0"`
	DiscountAmount      float64        `json:"discount_amount"       gorm:"not null;default:0"`
	TotalAmount         float64        `json:"total_amount"          gorm:"not null;default:0"`
	PaidAmount          float64        `json:"paid_amount"           gorm:"not null;default:0"`
	Notes               string         `json:"notes"`
	PostedBy            *uint          `json:"posted_by"`
	PostedAt            *time.Time     `json:"posted_at"`
	CreatedBy           *uint          `json:"created_by"`
	Lines               []InvoiceLine  `json:"lines,omitempty"       gorm:"foreignKey:InvoiceID"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at,omitempty"  gorm:"index"`
}

func (PurchaseInvoice) TableName() string { return "purchase_invoices" }

type InvoiceLine struct {
	ID                uint      `json:"id"                  gorm:"primaryKey"`
	InvoiceID         uint      `json:"invoice_id"          gorm:"not null;index"`
	TenantID          uint      `json:"tenant_id"           gorm:"not null"`
	GRNLineID         *uint     `json:"grn_line_id"`
	LineNumber        int       `json:"line_number"         gorm:"not null;default:1"`
	ProductID         uint      `json:"product_id"          gorm:"not null"`
	VariantID         *uint     `json:"variant_id"`
	Description       string    `json:"description"         gorm:"size:500"`
	Quantity          float64   `json:"quantity"            gorm:"not null"`
	UOMID             uint      `json:"uom_id"              gorm:"not null"`
	UnitPrice         float64   `json:"unit_price"          gorm:"not null;default:0"`
	DiscountPct       float64   `json:"discount_pct"        gorm:"not null;default:0"`
	TaxCodeID         *uint     `json:"tax_code_id"`
	TaxAmount         float64   `json:"tax_amount"          gorm:"not null;default:0"`
	LineTotal         float64   `json:"line_total"          gorm:"not null;default:0"`
	SupplierInvoiceNo string    `json:"supplier_invoice_no" gorm:"size:100"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (InvoiceLine) TableName() string { return "purchase_invoice_lines" }
