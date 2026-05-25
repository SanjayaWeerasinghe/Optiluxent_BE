package inventory

type CreateWarehouseRequest struct {
	Code    string `json:"code"    validate:"required,min=1,max=20"`
	Name    string `json:"name"    validate:"required,min=1,max=200"`
	Address string `json:"address"`
}

type UpdateWarehouseRequest struct {
	Name     string `json:"name"      validate:"omitempty,min=1,max=200"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}

type CreateLocationRequest struct {
	Code         string `json:"code"          validate:"required,min=1,max=20"`
	Name         string `json:"name"          validate:"required,min=1,max=200"`
	LocationType string `json:"location_type" validate:"omitempty"`
}

type UpdateLocationRequest struct {
	Name     string `json:"name"      validate:"omitempty,min=1,max=200"`
	IsActive *bool  `json:"is_active"`
}

type CreateStockEntryRequest struct {
	ProductID       uint    `json:"product_id"        validate:"required"`
	VariantID       *uint   `json:"variant_id"`
	WarehouseID     uint    `json:"warehouse_id"      validate:"required"`
	LocationID      *uint   `json:"location_id"`
	TransactionType string  `json:"transaction_type"  validate:"required"`
	ReferenceType   string  `json:"reference_type"    validate:"max=50"`
	ReferenceID     *uint   `json:"reference_id"`
	Quantity        float64 `json:"quantity"          validate:"required"`
	UnitCost        float64 `json:"unit_cost"         validate:"omitempty,min=0"`
	TransactionDate string  `json:"transaction_date"  validate:"required"`
	Notes           string  `json:"notes"`
}
