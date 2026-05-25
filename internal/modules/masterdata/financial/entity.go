package financial

import (
	"time"

	"gorm.io/gorm"
)

type Currency struct {
	ID        uint      `gorm:"primarykey"          json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null"            json:"name"`
	Symbol    string    `gorm:"not null"            json:"symbol"`
	IsBase    bool      `gorm:"default:false"       json:"is_base"`
	IsActive  bool      `gorm:"default:true"        json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExchangeRate struct {
	ID             uint      `gorm:"primarykey"     json:"id"`
	TenantID       uint      `gorm:"not null;index" json:"tenant_id"`
	FromCurrencyID uint      `gorm:"not null"       json:"from_currency_id"`
	ToCurrencyID   uint      `gorm:"not null"       json:"to_currency_id"`
	Rate           float64   `gorm:"not null"       json:"rate"`
	EffectiveDate  time.Time `gorm:"not null;index" json:"effective_date"`
	Source         string    `gorm:"default:manual" json:"source"`
	CreatedAt      time.Time `json:"created_at"`
}

type ChartOfAccount struct {
	ID           uint           `gorm:"primarykey"     json:"id"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	Code         string         `gorm:"not null"       json:"code"`
	Name         string         `gorm:"not null"       json:"name"`
	AccountType  string         `gorm:"not null"       json:"account_type"`
	ParentID     *uint          `json:"parent_id"`
	CurrencyID   uint           `gorm:"not null"       json:"currency_id"`
	IsControlled bool           `gorm:"default:false"  json:"is_controlled"`
	IsActive     bool           `gorm:"default:true"   json:"is_active"`
	Level        int            `gorm:"default:1"      json:"level"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"          json:"deleted_at,omitempty"`
}

type CostCenter struct {
	ID           uint           `gorm:"primarykey"     json:"id"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	Code         string         `gorm:"not null"       json:"code"`
	Name         string         `gorm:"not null"       json:"name"`
	DepartmentID *uint          `json:"department_id"`
	IsActive     bool           `gorm:"default:true"   json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"          json:"deleted_at,omitempty"`
}

type PaymentTerm struct {
	ID              uint      `gorm:"primarykey"     json:"id"`
	TenantID        uint      `gorm:"not null;index" json:"tenant_id"`
	Code            string    `gorm:"not null"       json:"code"`
	Name            string    `gorm:"not null"       json:"name"`
	DueDays         int       `gorm:"default:0"      json:"due_days"`
	DiscountDays    int       `gorm:"default:0"      json:"discount_days"`
	DiscountPercent float64   `gorm:"default:0"      json:"discount_percent"`
	IsActive        bool      `gorm:"default:true"   json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Bank struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Name       string    `gorm:"not null"   json:"name"`
	BranchName string    `json:"branch_name"`
	SwiftCode  string    `json:"swift_code"`
	Address    string    `json:"address"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CompanyBankAccount struct {
	ID            uint      `gorm:"primarykey"     json:"id"`
	TenantID      uint      `gorm:"not null;index" json:"tenant_id"`
	BankID        uint      `gorm:"not null"       json:"bank_id"`
	AccountNumber string    `gorm:"not null"       json:"account_number"`
	AccountName   string    `gorm:"not null"       json:"account_name"`
	CurrencyID    uint      `gorm:"not null"       json:"currency_id"`
	GLAccountID   uint      `gorm:"not null"       json:"gl_account_id"`
	IsDefault     bool      `gorm:"default:false"  json:"is_default"`
	IsActive      bool      `gorm:"default:true"   json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TaxCode struct {
	ID          uint      `gorm:"primarykey"     json:"id"`
	TenantID    uint      `gorm:"not null;index" json:"tenant_id"`
	Code        string    `gorm:"not null"       json:"code"`
	Name        string    `gorm:"not null"       json:"name"`
	TaxType     string    `gorm:"not null"       json:"tax_type"`
	Rate        float64   `gorm:"not null"       json:"rate"`
	GLAccountID uint      `gorm:"not null"       json:"gl_account_id"`
	IsActive    bool      `gorm:"default:true"   json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaxGroup struct {
	ID        uint       `gorm:"primarykey"          json:"id"`
	TenantID  uint       `gorm:"not null;index"      json:"tenant_id"`
	Name      string     `gorm:"not null"            json:"name"`
	IsActive  bool       `gorm:"default:true"        json:"is_active"`
	TaxCodes  []TaxCode  `gorm:"many2many:tax_group_codes;" json:"tax_codes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
