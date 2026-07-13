package sales

// ── Sales Quotation DTOs ──────────────────────────────────────────────────────

type CreateSQRequest struct {
	Code               string  `json:"code"`
	DocumentTypeID     *uint   `json:"document_type_id"`
	CustomerID         uint    `json:"customer_id"           validate:"required"`
	QuotationDate      string  `json:"quotation_date"        validate:"omitempty"`
	ValidUntil         *string `json:"valid_until"`
	CurrencyID         uint    `json:"currency_id"           validate:"required"`
	ExchangeRate       float64 `json:"exchange_rate"         validate:"omitempty,min=0"`
	PaymentTermID      *uint   `json:"payment_term_id"`
	WarehouseID        uint    `json:"warehouse_id"          validate:"required"`
	CustomerReference  string  `json:"customer_reference"`
	TermsAndConditions string  `json:"terms_and_conditions"`
	Notes              string  `json:"notes"`
}

type UpdateSQRequest struct {
	DocumentTypeID     *uint    `json:"document_type_id"`
	ValidUntil         *string  `json:"valid_until"`
	ExchangeRate       *float64 `json:"exchange_rate"`
	PaymentTermID      *uint    `json:"payment_term_id"`
	DiscountAmount     *float64 `json:"discount_amount"`
	CustomerReference  string   `json:"customer_reference"`
	TermsAndConditions string   `json:"terms_and_conditions"`
	Notes              string   `json:"notes"`
}

type AddSQLineRequest struct {
	ProductID   uint    `json:"product_id"  validate:"required"`
	VariantID   *uint   `json:"variant_id"`
	Description string  `json:"description" validate:"max=500"`
	Quantity    float64 `json:"quantity"    validate:"required,min=0"`
	UOMID       uint    `json:"uom_id"      validate:"required"`
	UnitPrice   float64 `json:"unit_price"  validate:"required,min=0"`
	DiscountPct float64 `json:"discount_pct" validate:"omitempty,min=0,max=100"`
	TaxCodeID   *uint   `json:"tax_code_id"`
	Notes       string  `json:"notes"`
}

type UpdateSQLineRequest struct {
	ProductID   uint    `json:"product_id"  validate:"required"`
	VariantID   *uint   `json:"variant_id"`
	Description string  `json:"description" validate:"max=500"`
	Quantity    float64 `json:"quantity"    validate:"required,min=0"`
	UOMID       uint    `json:"uom_id"      validate:"required"`
	UnitPrice   float64 `json:"unit_price"  validate:"required,min=0"`
	DiscountPct float64 `json:"discount_pct" validate:"omitempty,min=0,max=100"`
	TaxCodeID   *uint   `json:"tax_code_id"`
	Notes       string  `json:"notes"`
}

// ── Sales Order DTOs ──────────────────────────────────────────────────────────

type CreateSORequest struct {
	Code                 string  `json:"code"`
	DocumentTypeID       *uint   `json:"document_type_id"`
	CustomerID           uint    `json:"customer_id"            validate:"required"`
	OrderDate            string  `json:"order_date"             validate:"omitempty"`
	ExpectedDeliveryDate *string `json:"expected_delivery_date"`
	CurrencyID           uint    `json:"currency_id"            validate:"required"`
	ExchangeRate         float64 `json:"exchange_rate"          validate:"omitempty,min=0"`
	PaymentTermID        *uint   `json:"payment_term_id"`
	WarehouseID          uint    `json:"warehouse_id"           validate:"required"`
	Notes                string  `json:"notes"`
}

type UpdateSORequest struct {
	DocumentTypeID       *uint    `json:"document_type_id"`
	ExpectedDeliveryDate *string  `json:"expected_delivery_date"`
	ExchangeRate         *float64 `json:"exchange_rate"`
	PaymentTermID        *uint    `json:"payment_term_id"`
	DiscountAmount       *float64 `json:"discount_amount"`
	Notes                string   `json:"notes"`
}

type AddSOLineRequest struct {
	ProductID   uint    `json:"product_id"  validate:"required"`
	VariantID   *uint   `json:"variant_id"`
	Description string  `json:"description" validate:"max=500"`
	Quantity    float64 `json:"quantity"    validate:"required,min=0"`
	UOMID       uint    `json:"uom_id"      validate:"required"`
	UnitPrice   float64 `json:"unit_price"  validate:"required,min=0"`
	DiscountPct float64 `json:"discount_pct" validate:"omitempty,min=0,max=100"`
	TaxCodeID   *uint   `json:"tax_code_id"`
	Notes       string  `json:"notes"`
}

type UpdateSOLineRequest struct {
	ProductID   uint    `json:"product_id"  validate:"required"`
	VariantID   *uint   `json:"variant_id"`
	Description string  `json:"description" validate:"max=500"`
	Quantity    float64 `json:"quantity"    validate:"required,min=0"`
	UOMID       uint    `json:"uom_id"      validate:"required"`
	UnitPrice   float64 `json:"unit_price"  validate:"required,min=0"`
	DiscountPct float64 `json:"discount_pct" validate:"omitempty,min=0,max=100"`
	TaxCodeID   *uint   `json:"tax_code_id"`
	Notes       string  `json:"notes"`
}

// ── Delivery Order DTOs ───────────────────────────────────────────────────────

type CreateDORequest struct {
	Code           string  `json:"code"`
	DocumentTypeID *uint   `json:"document_type_id"`
	SOID         *uint   `json:"so_id"`
	CustomerID   uint    `json:"customer_id"   validate:"required"`
	DeliveryDate string  `json:"delivery_date" validate:"omitempty"`
	WarehouseID  uint    `json:"warehouse_id"  validate:"required"`
	Notes        string  `json:"notes"`
}

type UpdateDORequest struct {
	DocumentTypeID *uint  `json:"document_type_id"`
	DeliveryDate *string `json:"delivery_date"`
	WarehouseID  *uint   `json:"warehouse_id"`
	Notes        string  `json:"notes"`
}

type AddDOLineRequest struct {
	SOLineID   *uint   `json:"so_line_id"`
	ProductID  uint    `json:"product_id"  validate:"required"`
	VariantID  *uint   `json:"variant_id"`
	LocationID *uint   `json:"location_id"`
	UOMID      uint    `json:"uom_id"      validate:"required"`
	Quantity   float64 `json:"quantity"    validate:"required,min=0"`
	UnitCost   float64 `json:"unit_cost"   validate:"omitempty,min=0"`
	Notes      string  `json:"notes"`
}

type UpdateDOLineRequest struct {
	SOLineID   *uint   `json:"so_line_id"`
	ProductID  uint    `json:"product_id"  validate:"required"`
	VariantID  *uint   `json:"variant_id"`
	LocationID *uint   `json:"location_id"`
	UOMID      uint    `json:"uom_id"      validate:"required"`
	Quantity   float64 `json:"quantity"    validate:"required,min=0"`
	UnitCost   float64 `json:"unit_cost"   validate:"omitempty,min=0"`
	Notes      string  `json:"notes"`
}

// ── Sales Invoice DTOs ────────────────────────────────────────────────────────

type CreateSIRequest struct {
	Code             string  `json:"code"`
	DocumentTypeID   *uint   `json:"document_type_id"`
	CustomerID       uint    `json:"customer_id"        validate:"required"`
	SOID             uint    `json:"so_id"              validate:"required"`
	InvoiceDate      string  `json:"invoice_date"       validate:"omitempty"`
	DueDate          *string `json:"due_date"`
	CustomerPONumber string  `json:"customer_po_number"`
	CurrencyID       uint    `json:"currency_id"        validate:"required"`
	ExchangeRate     float64 `json:"exchange_rate"      validate:"omitempty,min=0"`
	PaymentTermID    *uint   `json:"payment_term_id"`
	Notes            string  `json:"notes"`
}

type AddSILineRequest struct {
	DOLineID    *uint   `json:"do_line_id"`
	ProductID   uint    `json:"product_id"  validate:"required"`
	VariantID   *uint   `json:"variant_id"`
	Description string  `json:"description" validate:"max=500"`
	Quantity    float64 `json:"quantity"    validate:"required,min=0"`
	UOMID       uint    `json:"uom_id"      validate:"required"`
	UnitPrice   float64 `json:"unit_price"  validate:"required,min=0"`
	DiscountPct float64 `json:"discount_pct" validate:"omitempty,min=0,max=100"`
	TaxCodeID   *uint   `json:"tax_code_id"`
	Notes       string  `json:"notes"`
}

type RecordPaymentRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}
