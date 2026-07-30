package contacts

import (
	"time"

	"gorm.io/gorm"
)

type Party struct {
	ID            uint           `json:"id"              gorm:"primaryKey"`
	TenantID      uint           `json:"tenant_id"       gorm:"not null;index"`
	Code          string         `json:"code"            gorm:"not null;size:20"`
	Name          string         `json:"name"            gorm:"not null;size:200"`
	LegalName     string         `json:"legal_name"      gorm:"size:200"`
	PartyType     string         `json:"party_type"      gorm:"not null;size:10"`
	Type          string         `json:"type"            gorm:"not null;size:10;default:COMPANY"`
	TaxRegNumber  string         `json:"tax_reg_number"  gorm:"size:50"`
	CurrencyID    uint           `json:"currency_id"     gorm:"not null"`
	PaymentTermID *uint          `json:"payment_term_id"`
	CreditLimit   float64        `json:"credit_limit"    gorm:"not null;default:0"`
	CreditType    string         `json:"credit_type"     gorm:"not null;size:10;default:CREDIT"`
	IsActive      bool           `json:"is_active"       gorm:"not null;default:true"`
	Notes         string         `json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Party) TableName() string { return "parties" }

type ContactPerson struct {
	ID          uint      `json:"id"           gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id"    gorm:"not null;index"`
	PartyID     uint      `json:"party_id"     gorm:"not null;index"`
	Name        string    `json:"name"         gorm:"not null;size:200"`
	Designation string    `json:"designation"  gorm:"size:100"`
	Phone       string    `json:"phone"        gorm:"size:50"`
	Mobile      string    `json:"mobile"       gorm:"size:50"`
	Email       string    `json:"email"        gorm:"size:255"`
	IsPrimary   bool      `json:"is_primary"   gorm:"not null;default:false"`
	IsActive    bool      `json:"is_active"    gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ContactPerson) TableName() string { return "contact_persons" }

type PartyAddress struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null;index"`
	PartyID      uint      `json:"party_id"      gorm:"not null;index"`
	AddressType  string    `json:"address_type"  gorm:"not null;size:10;default:BOTH"`
	AddressLine1 string    `json:"address_line1" gorm:"not null;size:255"`
	AddressLine2 string    `json:"address_line2" gorm:"size:255"`
	City         string    `json:"city"          gorm:"size:100"`
	StateID      *uint     `json:"state_id"`
	CountryID    uint      `json:"country_id"    gorm:"not null"`
	PostalCode   string    `json:"postal_code"   gorm:"size:20"`
	IsPrimary    bool      `json:"is_primary"    gorm:"not null;default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (PartyAddress) TableName() string { return "party_addresses" }
