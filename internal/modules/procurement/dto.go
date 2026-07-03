package procurement

// ── Purchase Request DTOs ─────────────────────────────────────────────────────

type CreatePRRequest struct {
	RequestDate  string  `json:"request_date"  validate:"omitempty"`
	RequiredDate *string `json:"required_date"`
	RequestedBy  *uint   `json:"requested_by"`
	DepartmentID *uint   `json:"department_id"`
	Notes        string  `json:"notes"`
}

type UpdatePRRequest struct {
	RequiredDate *string `json:"required_date"`
	RequestedBy  *uint   `json:"requested_by"`
	DepartmentID *uint   `json:"department_id"`
	Notes        string  `json:"notes"`
}

type AddPRItemRequest struct {
	ProductID      uint    `json:"product_id"      validate:"required"`
	VariantID      *uint   `json:"variant_id"`
	Description    string  `json:"description"     validate:"max=500"`
	Quantity       float64 `json:"quantity"        validate:"required,min=0"`
	UOMID          uint    `json:"uom_id"          validate:"required"`
	EstimatedPrice float64 `json:"estimated_price" validate:"omitempty,min=0"`
	TransferRatio  float64 `json:"transfer_ratio"  validate:"omitempty,min=0"`
	CurrencyID     *uint   `json:"currency_id"`
	Notes          string  `json:"notes"`
}

type UpdatePRItemRequest struct {
	ProductID      uint    `json:"product_id"      validate:"required"`
	VariantID      *uint   `json:"variant_id"`
	Description    string  `json:"description"     validate:"max=500"`
	Quantity       float64 `json:"quantity"        validate:"required,min=0"`
	UOMID          uint    `json:"uom_id"          validate:"required"`
	EstimatedPrice float64 `json:"estimated_price" validate:"omitempty,min=0"`
	TransferRatio  float64 `json:"transfer_ratio"  validate:"omitempty,min=0"`
	CurrencyID     *uint   `json:"currency_id"`
	Notes          string  `json:"notes"`
}

// ── Purchase Order DTOs ───────────────────────────────────────────────────────

type CreatePORequest struct {
	SupplierID    uint    `json:"supplier_id"     validate:"required"`
	PRID          *uint   `json:"pr_id"`
	OrderDate     string  `json:"order_date"      validate:"omitempty"`
	ExpectedDate  *string `json:"expected_date"`
	CurrencyID    uint    `json:"currency_id"     validate:"required"`
	ExchangeRate  float64 `json:"exchange_rate"   validate:"omitempty,min=0"`
	PaymentTermID *uint   `json:"payment_term_id"`
	WarehouseID   uint    `json:"warehouse_id"    validate:"required"`
	Notes         string  `json:"notes"`
}

type UpdatePORequest struct {
	ExpectedDate   *string  `json:"expected_date"`
	ExchangeRate   *float64 `json:"exchange_rate"`
	PaymentTermID  *uint    `json:"payment_term_id"`
	DiscountAmount *float64 `json:"discount_amount"`
	Notes          string   `json:"notes"`
}

type AddPOItemRequest struct {
	ProductID     uint    `json:"product_id"     validate:"required"`
	VariantID     *uint   `json:"variant_id"`
	Description   string  `json:"description"    validate:"max=500"`
	Quantity      float64 `json:"quantity"       validate:"required,min=0"`
	UOMID         uint    `json:"uom_id"         validate:"required"`
	UnitPrice     float64 `json:"unit_price"     validate:"required,min=0"`
	DiscountPct   float64 `json:"discount_pct"   validate:"omitempty,min=0,max=100"`
	TaxCodeID     *uint   `json:"tax_code_id"`
	TransferRatio float64 `json:"transfer_ratio" validate:"omitempty,min=0"`
	Notes         string  `json:"notes"`
}

type UpdatePOItemRequest struct {
	ProductID     uint    `json:"product_id"     validate:"required"`
	VariantID     *uint   `json:"variant_id"`
	Description   string  `json:"description"    validate:"max=500"`
	Quantity      float64 `json:"quantity"       validate:"required,min=0"`
	UOMID         uint    `json:"uom_id"         validate:"required"`
	UnitPrice     float64 `json:"unit_price"     validate:"required,min=0"`
	DiscountPct   float64 `json:"discount_pct"   validate:"omitempty,min=0,max=100"`
	TaxCodeID     *uint   `json:"tax_code_id"`
	TransferRatio float64 `json:"transfer_ratio" validate:"omitempty,min=0"`
	Notes         string  `json:"notes"`
}

// ── Goods Receipt DTOs ────────────────────────────────────────────────────────

type CreateGRNRequest struct {
	GRNType     string  `json:"grn_type"     validate:"omitempty,oneof=WITH_PO WITHOUT_PO CUSTOMER_RETURN PRODUCTION_RETURN PRODUCTION_OUTPUT"`
	POID        *uint   `json:"po_id"`
	MOID        *uint   `json:"mo_id"`
	SupplierID  uint    `json:"supplier_id"  validate:"omitempty"`
	ReceiptDate string  `json:"receipt_date" validate:"omitempty"`
	WarehouseID uint    `json:"warehouse_id" validate:"required"`
	Notes       string  `json:"notes"`
}

type AddGRNItemRequest struct {
	POLineID      *uint   `json:"po_line_id"`
	ProductID     uint    `json:"product_id"     validate:"required"`
	VariantID     *uint   `json:"variant_id"`
	Quantity      float64 `json:"quantity"       validate:"required,min=0"`
	UOMID         uint    `json:"uom_id"         validate:"required"`
	LocationID    *uint   `json:"location_id"`
	UnitCost      float64 `json:"unit_cost"      validate:"omitempty,min=0"`
	TransferRatio float64 `json:"transfer_ratio" validate:"omitempty,min=0"`
	Notes         string  `json:"notes"`
}

type UpdateGRNItemRequest struct {
	ProductID     uint    `json:"product_id"     validate:"required"`
	VariantID     *uint   `json:"variant_id"`
	Quantity      float64 `json:"quantity"       validate:"required,min=0"`
	UOMID         uint    `json:"uom_id"         validate:"required"`
	LocationID    *uint   `json:"location_id"`
	UnitCost      float64 `json:"unit_cost"      validate:"omitempty,min=0"`
	TransferRatio float64 `json:"transfer_ratio" validate:"omitempty,min=0"`
	Notes         string  `json:"notes"`
}

// ── Purchase Invoice DTOs ─────────────────────────────────────────────────────

type CreateInvoiceRequest struct {
	POID                uint    `json:"po_id"                 validate:"required"`
	SupplierInvoiceNo   string  `json:"supplier_invoice_no"`
	SupplierInvoiceDate *string `json:"supplier_invoice_date"`
	InvoiceDate         string  `json:"invoice_date"          validate:"omitempty"`
	DueDate             *string `json:"due_date"`
	CurrencyID          uint    `json:"currency_id"           validate:"required"`
	ExchangeRate        float64 `json:"exchange_rate"         validate:"omitempty,min=0"`
	PaymentTermID       *uint   `json:"payment_term_id"`
	Notes               string  `json:"notes"`
}

type RecordPaymentRequest struct {
	Amount        float64 `json:"amount"         validate:"required,min=0"`
	PaymentDate   string  `json:"payment_date"   validate:"omitempty"`
	PaymentRef    string  `json:"payment_ref"`
	Notes         string  `json:"notes"`
}

type AddInvoiceLineRequest struct {
	GRNLineID         *uint   `json:"grn_line_id"`
	ProductID         uint    `json:"product_id"         validate:"required"`
	VariantID         *uint   `json:"variant_id"`
	Description       string  `json:"description"        validate:"max=500"`
	Quantity          float64 `json:"quantity"           validate:"required,min=0"`
	UOMID             uint    `json:"uom_id"             validate:"required"`
	UnitPrice         float64 `json:"unit_price"         validate:"required,min=0"`
	DiscountPct       float64 `json:"discount_pct"       validate:"omitempty,min=0,max=100"`
	TaxCodeID         *uint   `json:"tax_code_id"`
	SupplierInvoiceNo string  `json:"supplier_invoice_no"`
	Notes             string  `json:"notes"`
}
