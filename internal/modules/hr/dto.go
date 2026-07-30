package hr

type UpsertFamilyRequest struct {
	Relation    string  `json:"relation"      validate:"required,oneof=SPOUSE CHILD PARENT SIBLING OTHER"`
	FullName    string  `json:"full_name"     validate:"required,min=1,max=200"`
	DateOfBirth *string `json:"date_of_birth"`
	IsDependent bool    `json:"is_dependent"`
	Notes       string  `json:"notes"         validate:"omitempty,max=500"`
}

type UpsertEmergencyRequest struct {
	FullName  string `json:"full_name"  validate:"required,min=1,max=200"`
	Relation  string `json:"relation"   validate:"required,min=1,max=30"`
	Phone     string `json:"phone"      validate:"required,min=1,max=50"`
	AltPhone  string `json:"alt_phone"  validate:"omitempty,max=50"`
	Address   string `json:"address"    validate:"omitempty,max=500"`
	IsPrimary bool   `json:"is_primary"`
}

type UpsertAttendanceRequest struct {
	AttendDate  string  `json:"attend_date"  validate:"required"`
	CheckIn     *string `json:"check_in"`
	CheckOut    *string `json:"check_out"`
	Status      string  `json:"status"       validate:"omitempty,oneof=PRESENT ABSENT HALF_DAY LEAVE HOLIDAY"`
	HoursWorked float64 `json:"hours_worked" validate:"omitempty,min=0,max=24"`
	Notes       string  `json:"notes"        validate:"omitempty,max=500"`
}

type CreateSalaryRevisionRequest struct {
	EffectiveDate string  `json:"effective_date" validate:"required"`
	BasicSalary   float64 `json:"basic_salary"   validate:"required,min=0"`
	Reason        string  `json:"reason"         validate:"omitempty,max=200"`
}
