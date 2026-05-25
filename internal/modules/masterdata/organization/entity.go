package organization

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID              uint      `gorm:"primarykey"     json:"id"`
	Name            string    `gorm:"not null"       json:"name"`
	LegalName       string    `json:"legal_name"`
	TaxRegNumber    string    `json:"tax_reg_number"`
	Logo            string    `json:"logo"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Website         string    `json:"website"`
	AddressLine1    string    `json:"address_line1"`
	AddressLine2    string    `json:"address_line2"`
	City            string    `json:"city"`
	StateID         *uint     `json:"state_id"`
	CountryID       *uint     `json:"country_id"`
	PostalCode      string    `json:"postal_code"`
	BaseCurrencyID  *uint     `json:"base_currency_id"`
	FiscalYearStart int       `json:"fiscal_year_start"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Department struct {
	ID                uint           `gorm:"primarykey"         json:"id"`
	TenantID          uint           `gorm:"not null;index"     json:"tenant_id"`
	Code              string         `gorm:"not null"           json:"code"`
	Name              string         `gorm:"not null"           json:"name"`
	ParentID          *uint          `json:"parent_id"`
	ManagerEmployeeID *uint          `json:"manager_employee_id"`
	IsActive          bool           `gorm:"default:true"       json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index"              json:"deleted_at,omitempty"`
}

type FiscalYear struct {
	ID        uint      `gorm:"primarykey"     json:"id"`
	TenantID  uint      `gorm:"not null;index" json:"tenant_id"`
	Name      string    `gorm:"not null"       json:"name"`
	StartDate time.Time `gorm:"not null"       json:"start_date"`
	EndDate   time.Time `gorm:"not null"       json:"end_date"`
	IsClosed  bool      `gorm:"default:false"  json:"is_closed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountingPeriod struct {
	ID           uint      `gorm:"primarykey"     json:"id"`
	TenantID     uint      `gorm:"not null;index" json:"tenant_id"`
	FiscalYearID uint      `gorm:"not null"       json:"fiscal_year_id"`
	PeriodNumber int       `gorm:"not null"       json:"period_number"`
	Name         string    `gorm:"not null"       json:"name"`
	StartDate    time.Time `gorm:"not null"       json:"start_date"`
	EndDate      time.Time `gorm:"not null"       json:"end_date"`
	IsClosed     bool      `gorm:"default:false"  json:"is_closed"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DocumentSequence struct {
	ID           uint      `gorm:"primarykey"     json:"id"`
	TenantID     uint      `gorm:"not null;index" json:"tenant_id"`
	DocumentType string    `gorm:"not null"       json:"document_type"`
	Prefix       string    `gorm:"not null"       json:"prefix"`
	NextNumber   int       `gorm:"default:1"      json:"next_number"`
	Padding      int       `gorm:"default:5"      json:"padding"`
	Suffix       string    `json:"suffix"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Country struct {
	ID        uint   `gorm:"primarykey"          json:"id"`
	Code      string `gorm:"uniqueIndex;not null" json:"code"`
	Name      string `gorm:"not null"            json:"name"`
	PhoneCode string `json:"phone_code"`
	IsActive  bool   `gorm:"default:true"        json:"is_active"`
}

type State struct {
	ID        uint   `gorm:"primarykey"     json:"id"`
	CountryID uint   `gorm:"not null;index" json:"country_id"`
	Code      string `gorm:"not null"       json:"code"`
	Name      string `gorm:"not null"       json:"name"`
}
