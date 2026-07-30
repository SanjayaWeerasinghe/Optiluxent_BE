package finance

type RecordPaymentRequest struct {
	// Which invoice this payment settles part or all of. `SI` = customer
	// paying us (INBOUND); `PI` = we pay a supplier (OUTBOUND).
	InvoiceKind   string  `json:"invoice_kind"   validate:"required,oneof=SI PI"`
	InvoiceID     uint    `json:"invoice_id"     validate:"required"`
	Amount        float64 `json:"amount"         validate:"required,gt=0"`
	PaymentDate   string  `json:"payment_date"   validate:"omitempty"`
	Method        string  `json:"method"         validate:"required,oneof=CASH BANK_TRANSFER CHEQUE CARD OTHER"`
	BankAccountID *uint   `json:"bank_account_id"`
	ReferenceNo   string  `json:"reference_no"   validate:"omitempty,max=100"`
	Notes         string  `json:"notes"`
}

type UpdateSettingsRequest struct {
	ARAccountID       *uint `json:"ar_account_id"`
	APAccountID       *uint `json:"ap_account_id"`
	CashAccountID     *uint `json:"cash_account_id"`
	SalesRevenueID    *uint `json:"sales_revenue_id"`
	PurchaseExpenseID *uint `json:"purchase_expense_id"`
	TaxAccountID      *uint `json:"tax_account_id"`
}

// AgingRow — one open invoice bucketed by days-overdue.
type AgingRow struct {
	InvoiceID       uint    `json:"invoice_id"`
	InvoiceCode     string  `json:"invoice_code"`
	InvoiceKind     string  `json:"invoice_kind"` // SI or PI
	PartyID         uint    `json:"party_id"`
	InvoiceDate     string  `json:"invoice_date"`
	DueDate         string  `json:"due_date"`
	TotalAmount     float64 `json:"total_amount"`
	PaidAmount      float64 `json:"paid_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	DaysOverdue     int     `json:"days_overdue"`
	Bucket          string  `json:"bucket"` // 0-30 / 31-60 / 61-90 / 90+
}

// PartyOutstanding — running unpaid total for a party across a side.
type PartyOutstanding struct {
	PartyID     uint    `json:"party_id"`
	Kind        string  `json:"kind"` // AR or AP
	Outstanding float64 `json:"outstanding"`
}
