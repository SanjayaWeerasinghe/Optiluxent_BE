package hr

import (
	"time"

	"gorm.io/gorm"
)

type JobPosition struct {
	ID           uint      `json:"id"            gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"     gorm:"not null;index"`
	Code         string    `json:"code"          gorm:"not null;size:20"`
	Name         string    `json:"name"          gorm:"not null;size:200"`
	DepartmentID *uint     `json:"department_id"`
	IsActive     bool      `json:"is_active"     gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (JobPosition) TableName() string { return "job_positions" }

type Employee struct {
	ID             uint           `json:"id"               gorm:"primaryKey"`
	TenantID       uint           `json:"tenant_id"        gorm:"not null;index"`
	Code           string         `json:"code"             gorm:"not null;size:20"`
	FirstName      string         `json:"first_name"       gorm:"not null;size:100"`
	LastName       string         `json:"last_name"        gorm:"not null;size:100"`
	DisplayName    string         `json:"display_name"     gorm:"size:200"`
	Gender         *string        `json:"gender"           gorm:"size:10"`
	DateOfBirth    *string        `json:"date_of_birth"    gorm:"type:date"`
	NICNumber      string         `json:"nic_number"       gorm:"size:20"`
	JobPositionID  *uint          `json:"job_position_id"`
	DepartmentID   *uint          `json:"department_id"`
	ManagerID      *uint          `json:"manager_id"`
	EmploymentType string         `json:"employment_type"  gorm:"not null;size:20;default:PERMANENT"`
	DateJoined     string         `json:"date_joined"      gorm:"type:date;not null"`
	DateLeft       *string        `json:"date_left"        gorm:"type:date"`
	// ContractEndDate — set for CONTRACT / INTERN employees so the HR
	// dashboard can flag upcoming expirations. Nullable for PERMANENT.
	ContractEndDate *string       `json:"contract_end_date" gorm:"type:date"`
	// CVUrl points at the uploaded CV file (served by /uploads/...).
	// Set by the /hr/employees/:id/cv upload endpoint; blank until then.
	CVUrl          string         `json:"cv_url"           gorm:"size:500"`
	Email          string         `json:"email"            gorm:"size:255"`
	Phone          string         `json:"phone"            gorm:"size:50"`
	Mobile         string         `json:"mobile"           gorm:"size:50"`
	BankID         *uint          `json:"bank_id"`
	BankAccountNo  string         `json:"bank_account_no"  gorm:"size:50"`
	BasicSalary    float64        `json:"basic_salary"     gorm:"not null;default:0"`
	CurrencyID     uint           `json:"currency_id"      gorm:"not null"`
	IsActive       bool           `json:"is_active"        gorm:"not null;default:true"`
	Notes          string         `json:"notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Employee) TableName() string { return "employees" }
