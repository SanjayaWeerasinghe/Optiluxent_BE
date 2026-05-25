package manufacturing

type CreateBOMRequest struct {
	Code      string  `json:"code"       validate:"required,min=1,max=50"`
	ProductID uint    `json:"product_id" validate:"required"`
	VariantID *uint   `json:"variant_id"`
	Quantity  float64 `json:"quantity"   validate:"omitempty,min=0"`
	UOMID     uint    `json:"uom_id"     validate:"required"`
	BOMType   string  `json:"bom_type"   validate:"omitempty"`
	Notes     string  `json:"notes"`
}

type UpdateBOMRequest struct {
	Quantity float64 `json:"quantity"  validate:"omitempty,min=0"`
	IsActive *bool   `json:"is_active"`
	Notes    string  `json:"notes"`
}

type AddBOMLineRequest struct {
	ComponentID uint    `json:"component_id" validate:"required"`
	VariantID   *uint   `json:"variant_id"`
	Quantity    float64 `json:"quantity"     validate:"required,min=0"`
	UOMID       uint    `json:"uom_id"       validate:"required"`
	ScrapPct    float64 `json:"scrap_percent" validate:"omitempty,min=0,max=100"`
	Sequence    int     `json:"sequence"     validate:"omitempty,min=1"`
	Notes       string  `json:"notes"`
}

type CreateWorkCenterRequest struct {
	Code        string  `json:"code"          validate:"required,min=1,max=20"`
	Name        string  `json:"name"          validate:"required,min=1,max=200"`
	Capacity    float64 `json:"capacity"      validate:"omitempty,min=0"`
	CostPerHour float64 `json:"cost_per_hour" validate:"omitempty,min=0"`
	CurrencyID  uint    `json:"currency_id"   validate:"required"`
	Notes       string  `json:"notes"`
}

type UpdateWorkCenterRequest struct {
	Name        string   `json:"name"          validate:"omitempty,min=1,max=200"`
	Capacity    float64  `json:"capacity"      validate:"omitempty,min=0"`
	CostPerHour *float64 `json:"cost_per_hour"`
	IsActive    *bool    `json:"is_active"`
	Notes       string   `json:"notes"`
}

type CreateRoutingRequest struct {
	Code      string `json:"code"       validate:"required,min=1,max=50"`
	ProductID uint   `json:"product_id" validate:"required"`
	Notes     string `json:"notes"`
}

type AddOperationRequest struct {
	Sequence     int     `json:"sequence"       validate:"omitempty,min=1"`
	Name         string  `json:"name"           validate:"required,min=1,max=200"`
	WorkCenterID uint    `json:"work_center_id" validate:"required"`
	SetupTime    float64 `json:"setup_time"     validate:"omitempty,min=0"`
	CycleTime    float64 `json:"cycle_time"     validate:"omitempty,min=0"`
	Notes        string  `json:"notes"`
}
