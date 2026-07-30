package contacts

type CreatePartyRequest struct {
	Code          string  `json:"code"            validate:"required,min=1,max=20"`
	Name          string  `json:"name"            validate:"required,min=1,max=200"`
	LegalName     string  `json:"legal_name"      validate:"max=200"`
	PartyType     string  `json:"party_type"      validate:"required"`
	Type          string  `json:"type"            validate:"omitempty"`
	TaxRegNumber  string  `json:"tax_reg_number"  validate:"max=50"`
	CurrencyID    uint    `json:"currency_id"     validate:"required"`
	PaymentTermID *uint   `json:"payment_term_id"`
	CreditLimit   float64 `json:"credit_limit"    validate:"omitempty,min=0"`
	// CASH: SO confirm blocks until any prior unpaid invoice is settled.
	// CREDIT: SO confirm blocks when outstanding + this order exceeds credit_limit.
	// Defaults to CREDIT.
	CreditType    string  `json:"credit_type"     validate:"omitempty,oneof=CASH CREDIT"`
	Notes         string  `json:"notes"`
}

type UpdatePartyRequest struct {
	Name          string   `json:"name"            validate:"omitempty,min=1,max=200"`
	LegalName     string   `json:"legal_name"      validate:"max=200"`
	PartyType     string   `json:"party_type"      validate:"omitempty"`
	TaxRegNumber  string   `json:"tax_reg_number"  validate:"max=50"`
	CurrencyID    uint     `json:"currency_id"     validate:"omitempty"`
	PaymentTermID *uint    `json:"payment_term_id"`
	CreditLimit   *float64 `json:"credit_limit"`
	CreditType    string   `json:"credit_type"     validate:"omitempty,oneof=CASH CREDIT"`
	IsActive      *bool    `json:"is_active"`
	Notes         string   `json:"notes"`
}

type CreateContactPersonRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=200"`
	Designation string `json:"designation" validate:"max=100"`
	Phone       string `json:"phone"       validate:"max=50"`
	Mobile      string `json:"mobile"      validate:"max=50"`
	Email       string `json:"email"       validate:"omitempty,email"`
	IsPrimary   bool   `json:"is_primary"`
}

type UpdateContactPersonRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=1,max=200"`
	Designation string `json:"designation" validate:"max=100"`
	Phone       string `json:"phone"       validate:"max=50"`
	Mobile      string `json:"mobile"      validate:"max=50"`
	Email       string `json:"email"       validate:"omitempty,email"`
	IsPrimary   *bool  `json:"is_primary"`
	IsActive    *bool  `json:"is_active"`
}

type CreateAddressRequest struct {
	AddressType  string `json:"address_type"  validate:"required"`
	AddressLine1 string `json:"address_line1" validate:"required,min=1,max=255"`
	AddressLine2 string `json:"address_line2" validate:"max=255"`
	City         string `json:"city"          validate:"max=100"`
	StateID      *uint  `json:"state_id"`
	CountryID    uint   `json:"country_id"    validate:"required"`
	PostalCode   string `json:"postal_code"   validate:"max=20"`
	IsPrimary    bool   `json:"is_primary"`
}

type UpdateAddressRequest struct {
	AddressType  string `json:"address_type"  validate:"omitempty"`
	AddressLine1 string `json:"address_line1" validate:"omitempty,min=1,max=255"`
	AddressLine2 string `json:"address_line2" validate:"max=255"`
	City         string `json:"city"          validate:"max=100"`
	StateID      *uint  `json:"state_id"`
	CountryID    uint   `json:"country_id"    validate:"omitempty"`
	PostalCode   string `json:"postal_code"   validate:"max=20"`
	IsPrimary    *bool  `json:"is_primary"`
}
