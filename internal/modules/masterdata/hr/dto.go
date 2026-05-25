package hr

type CreateJobPositionRequest struct {
	Code         string `json:"code"          validate:"required,min=1,max=20"`
	Name         string `json:"name"          validate:"required,min=1,max=200"`
	DepartmentID *uint  `json:"department_id"`
}

type UpdateJobPositionRequest struct {
	Name         string `json:"name"          validate:"omitempty,min=1,max=200"`
	DepartmentID *uint  `json:"department_id"`
	IsActive     *bool  `json:"is_active"`
}

type CreateEmployeeRequest struct {
	Code           string   `json:"code"            validate:"required,min=1,max=20"`
	FirstName      string   `json:"first_name"      validate:"required,min=1,max=100"`
	LastName       string   `json:"last_name"       validate:"required,min=1,max=100"`
	DisplayName    string   `json:"display_name"    validate:"max=200"`
	Gender         string   `json:"gender"          validate:"omitempty"`
	DateOfBirth    *string  `json:"date_of_birth"`
	NICNumber      string   `json:"nic_number"      validate:"max=20"`
	JobPositionID  *uint    `json:"job_position_id"`
	DepartmentID   *uint    `json:"department_id"`
	ManagerID      *uint    `json:"manager_id"`
	EmploymentType string   `json:"employment_type" validate:"omitempty"`
	DateJoined     string   `json:"date_joined"     validate:"required"`
	Email          string   `json:"email"           validate:"omitempty,email"`
	Phone          string   `json:"phone"           validate:"max=50"`
	Mobile         string   `json:"mobile"          validate:"max=50"`
	BankID         *uint    `json:"bank_id"`
	BankAccountNo  string   `json:"bank_account_no" validate:"max=50"`
	BasicSalary    float64  `json:"basic_salary"    validate:"omitempty,min=0"`
	CurrencyID     uint     `json:"currency_id"     validate:"required"`
	Notes          string   `json:"notes"`
}

type UpdateEmployeeRequest struct {
	FirstName      string   `json:"first_name"      validate:"omitempty,min=1,max=100"`
	LastName       string   `json:"last_name"       validate:"omitempty,min=1,max=100"`
	DisplayName    string   `json:"display_name"    validate:"max=200"`
	Gender         string   `json:"gender"          validate:"omitempty"`
	JobPositionID  *uint    `json:"job_position_id"`
	DepartmentID   *uint    `json:"department_id"`
	ManagerID      *uint    `json:"manager_id"`
	EmploymentType string   `json:"employment_type" validate:"omitempty"`
	DateLeft       *string  `json:"date_left"`
	Email          string   `json:"email"           validate:"omitempty,email"`
	Phone          string   `json:"phone"           validate:"max=50"`
	Mobile         string   `json:"mobile"          validate:"max=50"`
	BankID         *uint    `json:"bank_id"`
	BankAccountNo  string   `json:"bank_account_no" validate:"max=50"`
	BasicSalary    *float64 `json:"basic_salary"`
	IsActive       *bool    `json:"is_active"`
	Notes          string   `json:"notes"`
}
