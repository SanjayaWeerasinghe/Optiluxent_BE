package sales

import (
	"time"

	"gorm.io/gorm"
)

// ── Sales Quotation ───────────────────────────────────────────────────────────

const (
	SQStatusDraft     = "DRAFT"
	SQStatusSent      = "SENT"
	SQStatusAccepted  = "ACCEPTED"
	SQStatusRejected  = "REJECTED"
	SQStatusExpired   = "EXPIRED"
	SQStatusCancelled = "CANCELLED"
)

type SalesQuotation struct {
	ID                 uint           `json:"id"                    gorm:"primaryKey"`
	TenantID           uint           `json:"tenant_id"             gorm:"not null;index"`
	Code               string         `json:"code"                  gorm:"not null;size:50"`
	CustomerID         uint           `json:"customer_id"           gorm:"not null;index"`
	QuotationDate      string         `json:"quotation_date"        gorm:"type:date;not null"`
	ValidUntil         *string        `json:"valid_until"           gorm:"type:date"`
	CurrencyID         uint           `json:"currency_id"           gorm:"not null"`
	ExchangeRate       float64        `json:"exchange_rate"         gorm:"not null;default:1"`
	PaymentTermID      *uint          `json:"payment_term_id"`
	WarehouseID        uint           `json:"warehouse_id"          gorm:"not null"`
	Status             string         `json:"status"                gorm:"not null;size:20;default:DRAFT"`
	Subtotal           float64        `json:"subtotal"              gorm:"not null;default:0"`
	TaxAmount          float64        `json:"tax_amount"            gorm:"not null;default:0"`
	DiscountAmount     float64        `json:"discount_amount"       gorm:"not null;default:0"`
	TotalAmount        float64        `json:"total_amount"          gorm:"not null;default:0"`
	CustomerReference  string         `json:"customer_reference"    gorm:"size:100"`
	TermsAndConditions string         `json:"terms_and_conditions"`
	Notes              string         `json:"notes"`
	ConvertedSOID      *uint          `json:"converted_so_id"`
	AcceptedBy         *uint          `json:"accepted_by"`
	AcceptedAt         *time.Time     `json:"accepted_at"`
	RejectedBy         *uint          `json:"rejected_by"`
	CreatedBy          *uint          `json:"created_by"`
	Lines              []SQLine       `json:"lines,omitempty"       gorm:"foreignKey:SQID"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty"  gorm:"index"`
}

func (SalesQuotation) TableName() string { return "sales_quotations" }

type SQLine struct {
	ID          uint      `json:"id"            gorm:"primaryKey"`
	SQID        uint      `json:"sq_id"         gorm:"not null;index"`
	TenantID    uint      `json:"tenant_id"     gorm:"not null"`
	LineNumber  int       `json:"line_number"   gorm:"not null;default:1"`
	ProductID   uint      `json:"product_id"    gorm:"not null"`
	VariantID   *uint     `json:"variant_id"`
	Description string    `json:"description"   gorm:"size:500"`
	Quantity    float64   `json:"quantity"      gorm:"not null"`
	UOMID       uint      `json:"uom_id"        gorm:"not null"`
	UnitPrice   float64   `json:"unit_price"    gorm:"not null;default:0"`
	DiscountPct float64   `json:"discount_pct"  gorm:"not null;default:0"`
	TaxCodeID   *uint     `json:"tax_code_id"`
	TaxAmount   float64   `json:"tax_amount"    gorm:"not null;default:0"`
	LineTotal   float64   `json:"line_total"    gorm:"not null;default:0"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SQLine) TableName() string { return "sales_quotation_lines" }

// ── Sales Order ───────────────────────────────────────────────────────────────

const (
	SOStatusDraft     = "DRAFT"
	SOStatusConfirmed = "CONFIRMED"
	SOStatusPartial   = "PARTIAL"
	SOStatusDelivered = "DELIVERED"
	SOStatusCancelled = "CANCELLED"
)

type SalesOrder struct {
	ID                   uint           `json:"id"                      gorm:"primaryKey"`
	TenantID             uint           `json:"tenant_id"               gorm:"not null;index"`
	Code                 string         `json:"code"                    gorm:"not null;size:50"`
	SQID                 *uint          `json:"sq_id"`
	CustomerID           uint           `json:"customer_id"             gorm:"not null;index"`
	OrderDate            string         `json:"order_date"              gorm:"type:date;not null"`
	ExpectedDeliveryDate *string        `json:"expected_delivery_date"  gorm:"type:date"`
	CurrencyID           uint           `json:"currency_id"             gorm:"not null"`
	ExchangeRate         float64        `json:"exchange_rate"           gorm:"not null;default:1"`
	PaymentTermID        *uint          `json:"payment_term_id"`
	WarehouseID          uint           `json:"warehouse_id"            gorm:"not null"`
	Status               string         `json:"status"                  gorm:"not null;size:20;default:DRAFT"`
	Subtotal             float64        `json:"subtotal"                gorm:"not null;default:0"`
	TaxAmount            float64        `json:"tax_amount"              gorm:"not null;default:0"`
	DiscountAmount       float64        `json:"discount_amount"         gorm:"not null;default:0"`
	TotalAmount          float64        `json:"total_amount"            gorm:"not null;default:0"`
	Notes                string         `json:"notes"`
	ConfirmedBy          *uint          `json:"confirmed_by"`
	ConfirmedAt          *time.Time     `json:"confirmed_at"`
	CreatedBy            *uint          `json:"created_by"`
	Lines                []SOLine       `json:"lines,omitempty"         gorm:"foreignKey:SOID"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"deleted_at,omitempty"    gorm:"index"`
}

func (SalesOrder) TableName() string { return "sales_orders" }

type SOLine struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	SOID         uint      `json:"so_id"         gorm:"not null;index"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null"`
	LineNumber   int       `json:"line_number"   gorm:"not null;default:1"`
	ProductID    uint      `json:"product_id"    gorm:"not null"`
	VariantID    *uint     `json:"variant_id"`
	Description  string    `json:"description"   gorm:"size:500"`
	Quantity     float64   `json:"quantity"      gorm:"not null"`
	UOMID        uint      `json:"uom_id"        gorm:"not null"`
	UnitPrice    float64   `json:"unit_price"    gorm:"not null;default:0"`
	DiscountPct  float64   `json:"discount_pct"  gorm:"not null;default:0"`
	TaxCodeID    *uint     `json:"tax_code_id"`
	TaxAmount    float64   `json:"tax_amount"    gorm:"not null;default:0"`
	LineTotal    float64   `json:"line_total"    gorm:"not null;default:0"`
	DeliveredQty float64   `json:"delivered_qty" gorm:"not null;default:0"`
	InvoicedQty  float64   `json:"invoiced_qty"  gorm:"not null;default:0"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SOLine) TableName() string { return "sales_order_lines" }

// ── Delivery Order ────────────────────────────────────────────────────────────

const (
	DOStatusDraft     = "DRAFT"
	DOStatusConfirmed = "CONFIRMED"
	DOStatusCancelled = "CANCELLED"
)

type DeliveryOrder struct {
	ID          uint           `json:"id"           gorm:"primaryKey"`
	TenantID    uint           `json:"tenant_id"    gorm:"not null;index"`
	Code        string         `json:"code"         gorm:"not null;size:50"`
	SOID        *uint          `json:"so_id"`
	CustomerID  uint           `json:"customer_id"  gorm:"not null;index"`
	DeliveryDate string        `json:"delivery_date" gorm:"type:date;not null"`
	WarehouseID uint           `json:"warehouse_id" gorm:"not null"`
	Status      string         `json:"status"       gorm:"not null;size:20;default:DRAFT"`
	Notes       string         `json:"notes"`
	ConfirmedBy *uint          `json:"confirmed_by"`
	ConfirmedAt *time.Time     `json:"confirmed_at"`
	CreatedBy   *uint          `json:"created_by"`
	Lines       []DOLine       `json:"lines,omitempty" gorm:"foreignKey:DOID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (DeliveryOrder) TableName() string { return "delivery_orders" }

type DOLine struct {
	ID         uint      `json:"id"          gorm:"primaryKey"`
	DOID       uint      `json:"do_id"       gorm:"not null;index"`
	TenantID   uint      `json:"tenant_id"   gorm:"not null"`
	SOLineID   *uint     `json:"so_line_id"`
	LineNumber int       `json:"line_number" gorm:"not null;default:1"`
	ProductID  uint      `json:"product_id"  gorm:"not null"`
	VariantID  *uint     `json:"variant_id"`
	LocationID *uint     `json:"location_id"`
	UOMID      uint      `json:"uom_id"      gorm:"not null"`
	Quantity   float64   `json:"quantity"    gorm:"not null"`
	UnitCost   float64   `json:"unit_cost"   gorm:"not null;default:0"`
	TotalCost  float64   `json:"total_cost"  gorm:"not null;default:0"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (DOLine) TableName() string { return "delivery_order_lines" }

// ── Sales Invoice ─────────────────────────────────────────────────────────────

const (
	SIStatusDraft     = "DRAFT"
	SIStatusPosted    = "POSTED"
	SIStatusPartial   = "PARTIAL"
	SIStatusPaid      = "PAID"
	SIStatusCancelled = "CANCELLED"
)

type SalesInvoice struct {
	ID              uint           `json:"id"                gorm:"primaryKey"`
	TenantID        uint           `json:"tenant_id"         gorm:"not null;index"`
	Code            string         `json:"code"              gorm:"not null;size:50"`
	CustomerID      uint           `json:"customer_id"       gorm:"not null;index"`
	SOID            uint           `json:"so_id"             gorm:"not null"`
	InvoiceDate     string         `json:"invoice_date"      gorm:"type:date;not null"`
	DueDate         *string        `json:"due_date"          gorm:"type:date"`
	CustomerPONumber string        `json:"customer_po_number" gorm:"size:100"`
	CurrencyID      uint           `json:"currency_id"       gorm:"not null"`
	ExchangeRate    float64        `json:"exchange_rate"     gorm:"not null;default:1"`
	PaymentTermID   *uint          `json:"payment_term_id"`
	Status          string         `json:"status"            gorm:"not null;size:20;default:DRAFT"`
	Subtotal        float64        `json:"subtotal"          gorm:"not null;default:0"`
	TaxAmount       float64        `json:"tax_amount"        gorm:"not null;default:0"`
	DiscountAmount  float64        `json:"discount_amount"   gorm:"not null;default:0"`
	TotalAmount     float64        `json:"total_amount"      gorm:"not null;default:0"`
	PaidAmount      float64        `json:"paid_amount"       gorm:"not null;default:0"`
	Notes           string         `json:"notes"`
	PostedBy        *uint          `json:"posted_by"`
	PostedAt        *time.Time     `json:"posted_at"`
	CreatedBy       *uint          `json:"created_by"`
	Lines           []SILine       `json:"lines,omitempty"   gorm:"foreignKey:InvoiceID"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (SalesInvoice) TableName() string { return "sales_invoices" }

type SILine struct {
	ID          uint      `json:"id"           gorm:"primaryKey"`
	InvoiceID   uint      `json:"invoice_id"   gorm:"not null;index"`
	TenantID    uint      `json:"tenant_id"    gorm:"not null"`
	DOLineID    *uint     `json:"do_line_id"`
	LineNumber  int       `json:"line_number"  gorm:"not null;default:1"`
	ProductID   uint      `json:"product_id"   gorm:"not null"`
	VariantID   *uint     `json:"variant_id"`
	Description string    `json:"description"  gorm:"size:500"`
	Quantity    float64   `json:"quantity"     gorm:"not null"`
	UOMID       uint      `json:"uom_id"       gorm:"not null"`
	UnitPrice   float64   `json:"unit_price"   gorm:"not null;default:0"`
	DiscountPct float64   `json:"discount_pct" gorm:"not null;default:0"`
	TaxCodeID   *uint     `json:"tax_code_id"`
	TaxAmount   float64   `json:"tax_amount"   gorm:"not null;default:0"`
	LineTotal   float64   `json:"line_total"   gorm:"not null;default:0"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SILine) TableName() string { return "sales_invoice_lines" }
