package products

type CreateCategoryRequest struct {
	Code     string `json:"code"      validate:"required,min=1,max=20"`
	Name     string `json:"name"      validate:"required,min=1,max=200"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name     string `json:"name"      validate:"omitempty,min=1,max=200"`
	ParentID *uint  `json:"parent_id"`
	IsActive *bool  `json:"is_active"`
}

type CreateUOMRequest struct {
	Code             string  `json:"code"              validate:"required,min=1,max=20"`
	Name             string  `json:"name"              validate:"required,min=1,max=100"`
	UOMType          string  `json:"uom_type"          validate:"omitempty"`
	BaseUOMID        *uint   `json:"base_uom_id"`
	ConversionFactor float64 `json:"conversion_factor" validate:"omitempty,min=0"`
}

type UpdateUOMRequest struct {
	Name             string  `json:"name"              validate:"omitempty,min=1,max=100"`
	BaseUOMID        *uint   `json:"base_uom_id"`
	ConversionFactor float64 `json:"conversion_factor" validate:"omitempty,min=0"`
	IsActive         *bool   `json:"is_active"`
}

type CreateProductRequest struct {
	Code           string  `json:"code"            validate:"required,min=1,max=50"`
	Name           string  `json:"name"            validate:"required,min=1,max=200"`
	Description    string  `json:"description"`
	ProductType    string  `json:"product_type"    validate:"omitempty"`
	CategoryID     *uint   `json:"category_id"`
	BaseUOMID      uint    `json:"base_uom_id"     validate:"required"`
	PurchaseUOMID  *uint   `json:"purchase_uom_id"`
	SalesUOMID     *uint   `json:"sales_uom_id"`
	TaxCodeID      *uint   `json:"tax_code_id"`
	CostPrice      float64 `json:"cost_price"      validate:"omitempty,min=0"`
	StandardPrice  float64 `json:"standard_price"  validate:"omitempty,min=0"`
	MinStockQty    float64 `json:"min_stock_qty"   validate:"omitempty,min=0"`
	ReorderQty     float64 `json:"reorder_qty"     validate:"omitempty,min=0"`
	LeadTimeDays   int     `json:"lead_time_days"  validate:"omitempty,min=0"`
	IsPurchased    bool    `json:"is_purchased"`
	IsSold         bool    `json:"is_sold"`
	IsManufactured bool    `json:"is_manufactured"`
	Notes          string  `json:"notes"`
}

type UpdateProductRequest struct {
	Name           string   `json:"name"            validate:"omitempty,min=1,max=200"`
	Description    string   `json:"description"`
	ProductType    string   `json:"product_type"    validate:"omitempty"`
	CategoryID     *uint    `json:"category_id"`
	BaseUOMID      uint     `json:"base_uom_id"     validate:"omitempty"`
	PurchaseUOMID  *uint    `json:"purchase_uom_id"`
	SalesUOMID     *uint    `json:"sales_uom_id"`
	TaxCodeID      *uint    `json:"tax_code_id"`
	CostPrice      *float64 `json:"cost_price"`
	StandardPrice  *float64 `json:"standard_price"`
	MinStockQty    *float64 `json:"min_stock_qty"`
	ReorderQty     *float64 `json:"reorder_qty"`
	LeadTimeDays   *int     `json:"lead_time_days"`
	IsPurchased    *bool    `json:"is_purchased"`
	IsSold         *bool    `json:"is_sold"`
	IsManufactured *bool    `json:"is_manufactured"`
	IsActive       *bool    `json:"is_active"`
	Notes          string   `json:"notes"`
}

type CreateVariantRequest struct {
	Code       string  `json:"code"        validate:"required,min=1,max=50"`
	Name       string  `json:"name"        validate:"required,min=1,max=200"`
	CostPrice  float64 `json:"cost_price"  validate:"omitempty,min=0"`
	SalesPrice float64 `json:"sales_price" validate:"omitempty,min=0"`
}

type CreatePriceRequest struct {
	PartyID       *uint   `json:"party_id"`
	PriceType     string  `json:"price_type"      validate:"omitempty"`
	CurrencyID    uint    `json:"currency_id"     validate:"required"`
	Price         float64 `json:"price"           validate:"required,min=0"`
	MinQty        float64 `json:"min_qty"         validate:"omitempty,min=0"`
	EffectiveFrom string  `json:"effective_from"  validate:"required"`
	EffectiveTo   *string `json:"effective_to"`
}
