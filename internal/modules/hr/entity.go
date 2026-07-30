package hr

import "time"

const (
	AttendancePresent  = "PRESENT"
	AttendanceAbsent   = "ABSENT"
	AttendanceHalfDay  = "HALF_DAY"
	AttendanceLeave    = "LEAVE"
	AttendanceHoliday  = "HOLIDAY"

	RelationSpouse  = "SPOUSE"
	RelationChild   = "CHILD"
	RelationParent  = "PARENT"
	RelationSibling = "SIBLING"
	RelationOther   = "OTHER"
)

// Family — dependents / next-of-kin listed on an employee record.
type Family struct {
	ID           uint      `json:"id"             gorm:"primaryKey"`
	TenantID     uint      `json:"tenant_id"      gorm:"not null;index"`
	EmployeeID   uint      `json:"employee_id"    gorm:"not null;index"`
	Relation     string    `json:"relation"       gorm:"not null;size:30"`
	FullName     string    `json:"full_name"      gorm:"not null;size:200"`
	DateOfBirth  *string   `json:"date_of_birth"  gorm:"type:date"`
	IsDependent  bool      `json:"is_dependent"   gorm:"not null;default:false"`
	Notes        string    `json:"notes"          gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Family) TableName() string { return "employee_family" }

// EmergencyContact — who to reach if the employee has an issue at work.
type EmergencyContact struct {
	ID         uint      `json:"id"          gorm:"primaryKey"`
	TenantID   uint      `json:"tenant_id"   gorm:"not null;index"`
	EmployeeID uint      `json:"employee_id" gorm:"not null;index"`
	FullName   string    `json:"full_name"   gorm:"not null;size:200"`
	Relation   string    `json:"relation"    gorm:"not null;size:30"`
	Phone      string    `json:"phone"       gorm:"not null;size:50"`
	AltPhone   string    `json:"alt_phone"   gorm:"size:50"`
	Address    string    `json:"address"     gorm:"size:500"`
	IsPrimary  bool      `json:"is_primary"  gorm:"not null;default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (EmergencyContact) TableName() string { return "employee_emergency_contact" }

// Attendance — one row per employee per day. Uniqueness on
// (tenant_id, employee_id, attend_date) enforced at the DB level (see
// migration 000046 uidx_attendance_emp_date).
type Attendance struct {
	ID          uint      `json:"id"           gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id"    gorm:"not null;index"`
	EmployeeID  uint      `json:"employee_id"  gorm:"not null;index"`
	AttendDate  string    `json:"attend_date"  gorm:"type:date;not null"`
	CheckIn     *string   `json:"check_in"     gorm:"type:time"`
	CheckOut    *string   `json:"check_out"    gorm:"type:time"`
	Status      string    `json:"status"       gorm:"not null;size:20;default:PRESENT"`
	HoursWorked float64   `json:"hours_worked" gorm:"not null;default:0"`
	Notes       string    `json:"notes"        gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Attendance) TableName() string { return "attendance" }

// SalaryHistory — append-only log of every basic_salary revision.
type SalaryHistory struct {
	ID            uint      `json:"id"             gorm:"primaryKey"`
	TenantID      uint      `json:"tenant_id"      gorm:"not null;index"`
	EmployeeID    uint      `json:"employee_id"    gorm:"not null;index"`
	EffectiveDate string    `json:"effective_date" gorm:"type:date;not null"`
	BasicSalary   float64   `json:"basic_salary"   gorm:"not null"`
	Reason        string    `json:"reason"         gorm:"size:200"`
	CreatedBy     *uint     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func (SalaryHistory) TableName() string { return "salary_history" }
