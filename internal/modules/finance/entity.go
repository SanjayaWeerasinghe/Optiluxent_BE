package finance

import "time"

// ── Payment ─────────────────────────────────────────────────────────────────
//
// One row per payment transaction. AR receipts and AP disbursements share
// the same table; `direction` distinguishes them. An invoice can accumulate
// many payment rows — SUM(amount) must equal invoices.paid_amount.
const (
	DirectionInbound  = "INBOUND"  // customer paid us (AR receipt)
	DirectionOutbound = "OUTBOUND" // we paid a supplier (AP disbursement)

	InvoiceKindSI = "SI" // sales_invoices
	InvoiceKindPI = "PI" // purchase_invoices

	MethodCash         = "CASH"
	MethodBankTransfer = "BANK_TRANSFER"
	MethodCheque       = "CHEQUE"
	MethodCard         = "CARD"
	MethodOther        = "OTHER"
)

type Payment struct {
	ID            uint      `json:"id"              gorm:"primaryKey"`
	TenantID      uint      `json:"tenant_id"       gorm:"not null;index"`
	Code          string    `json:"code"            gorm:"not null;size:50"`
	Direction     string    `json:"direction"       gorm:"not null;size:10"`
	InvoiceKind   string    `json:"invoice_kind"    gorm:"not null;size:10"`
	InvoiceID     uint      `json:"invoice_id"      gorm:"not null;index"`
	PartyID       uint      `json:"party_id"        gorm:"not null;index"`
	Amount        float64   `json:"amount"          gorm:"not null"`
	CurrencyID    uint      `json:"currency_id"     gorm:"not null"`
	ExchangeRate  float64   `json:"exchange_rate"   gorm:"not null;default:1"`
	PaymentDate   string    `json:"payment_date"    gorm:"type:date;not null"`
	Method        string    `json:"method"          gorm:"not null;size:20"`
	BankAccountID *uint     `json:"bank_account_id"`
	ReferenceNo   string    `json:"reference_no"    gorm:"size:100"`
	Notes         string    `json:"notes"`
	JournalID     *uint     `json:"journal_id"`
	CreatedBy     *uint     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Payment) TableName() string { return "payments" }

// ── Finance Settings ────────────────────────────────────────────────────────
//
// Tenant defaults consumed by the GL poster. Nullable — an unconfigured
// account skips that leg of the JE (system stays operational for payment
// tracking even without full GL setup).

type FinanceSettings struct {
	TenantID          uint      `json:"tenant_id"           gorm:"primaryKey"`
	ARAccountID       *uint     `json:"ar_account_id"`
	APAccountID       *uint     `json:"ap_account_id"`
	CashAccountID     *uint     `json:"cash_account_id"`
	SalesRevenueID    *uint     `json:"sales_revenue_id"`
	PurchaseExpenseID *uint     `json:"purchase_expense_id"`
	TaxAccountID      *uint     `json:"tax_account_id"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (FinanceSettings) TableName() string { return "finance_settings" }

// ── Journal Entry + GL Line ─────────────────────────────────────────────────

const (
	SourceTypeSIPost  = "SI_POST"
	SourceTypePIPost  = "PI_POST"
	SourceTypePayment = "PAYMENT"
)

type JournalEntry struct {
	ID           uint      `json:"id"             gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"      gorm:"not null;index"`
	Code         string    `json:"code"           gorm:"not null;size:50"`
	PostDate     string    `json:"post_date"      gorm:"type:date;not null"`
	Narration    string    `json:"narration"      gorm:"not null;size:500"`
	SourceType   string    `json:"source_type"    gorm:"not null;size:30"`
	SourceID     uint      `json:"source_id"      gorm:"not null"`
	IsPosted     bool      `json:"is_posted"      gorm:"not null;default:true"`
	IsReversalOf *uint     `json:"is_reversal_of"`
	CreatedBy    *uint     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	Lines        []GLLine  `json:"lines,omitempty" gorm:"foreignKey:JournalID"`
}

func (JournalEntry) TableName() string { return "journal_entries" }

type GLLine struct {
	ID           uint    `json:"id"           gorm:"primaryKey"`
	TenantID     uint    `json:"tenant_id"    gorm:"not null;index"`
	JournalID    uint    `json:"journal_id"   gorm:"not null;index"`
	LineNumber   int     `json:"line_number"  gorm:"not null"`
	AccountID    uint    `json:"account_id"   gorm:"not null;index"`
	PartyID      *uint   `json:"party_id"`
	CostCenterID *uint   `json:"cost_center_id"`
	Debit        float64 `json:"debit"        gorm:"not null;default:0"`
	Credit       float64 `json:"credit"       gorm:"not null;default:0"`
	Description  string  `json:"description"  gorm:"size:500"`
}

func (GLLine) TableName() string { return "gl_lines" }
