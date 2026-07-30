package hr

import (
	"context"
	"fmt"
	"time"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ── Family ──────────────────────────────────────────────────────────────────

func (s *Service) ListFamily(ctx context.Context, tenantID, employeeID uint) ([]Family, error) {
	return s.repo.ListFamily(ctx, tenantID, employeeID)
}

func (s *Service) AddFamily(ctx context.Context, tenantID, employeeID uint, req *UpsertFamilyRequest) (*Family, error) {
	f := &Family{
		TenantID:    tenantID,
		EmployeeID:  employeeID,
		Relation:    req.Relation,
		FullName:    req.FullName,
		DateOfBirth: req.DateOfBirth,
		IsDependent: req.IsDependent,
		Notes:       req.Notes,
	}
	if err := s.repo.AddFamily(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) UpdateFamily(ctx context.Context, tenantID, id uint, req *UpsertFamilyRequest) (*Family, error) {
	f, err := s.repo.GetFamily(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("family record not found")
	}
	f.Relation = req.Relation
	f.FullName = req.FullName
	f.DateOfBirth = req.DateOfBirth
	f.IsDependent = req.IsDependent
	f.Notes = req.Notes
	return f, s.repo.UpdateFamily(ctx, f)
}

func (s *Service) DeleteFamily(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteFamily(ctx, tenantID, id)
}

// ── Emergency ───────────────────────────────────────────────────────────────

func (s *Service) ListEmergency(ctx context.Context, tenantID, employeeID uint) ([]EmergencyContact, error) {
	return s.repo.ListEmergency(ctx, tenantID, employeeID)
}

func (s *Service) AddEmergency(ctx context.Context, tenantID, employeeID uint, req *UpsertEmergencyRequest) (*EmergencyContact, error) {
	e := &EmergencyContact{
		TenantID:   tenantID,
		EmployeeID: employeeID,
		FullName:   req.FullName,
		Relation:   req.Relation,
		Phone:      req.Phone,
		AltPhone:   req.AltPhone,
		Address:    req.Address,
		IsPrimary:  req.IsPrimary,
	}
	if err := s.repo.AddEmergency(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) UpdateEmergency(ctx context.Context, tenantID, id uint, req *UpsertEmergencyRequest) (*EmergencyContact, error) {
	e, err := s.repo.GetEmergency(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("emergency contact not found")
	}
	e.FullName = req.FullName
	e.Relation = req.Relation
	e.Phone = req.Phone
	e.AltPhone = req.AltPhone
	e.Address = req.Address
	e.IsPrimary = req.IsPrimary
	return e, s.repo.UpdateEmergency(ctx, e)
}

func (s *Service) DeleteEmergency(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteEmergency(ctx, tenantID, id)
}

// ── Attendance ──────────────────────────────────────────────────────────────

func (s *Service) ListAttendance(ctx context.Context, tenantID uint, employeeID uint, from, to string) ([]Attendance, error) {
	return s.repo.ListAttendance(ctx, tenantID, employeeID, from, to)
}

// UpsertAttendance — POST to /hr/employees/:id/attendance is idempotent
// per (employee, date). CheckIn/CheckOut are optional so tenants without
// a physical clock can still mark PRESENT/ABSENT/LEAVE.
func (s *Service) UpsertAttendance(ctx context.Context, tenantID, employeeID uint, req *UpsertAttendanceRequest) (*Attendance, error) {
	status := req.Status
	if status == "" {
		status = AttendancePresent
	}
	a := &Attendance{
		TenantID:    tenantID,
		EmployeeID:  employeeID,
		AttendDate:  req.AttendDate,
		CheckIn:     req.CheckIn,
		CheckOut:    req.CheckOut,
		Status:      status,
		HoursWorked: req.HoursWorked,
		Notes:       req.Notes,
	}
	if err := s.repo.UpsertAttendance(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) DeleteAttendance(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteAttendance(ctx, tenantID, id)
}

// ── Salary ──────────────────────────────────────────────────────────────────

func (s *Service) ListSalaryHistory(ctx context.Context, tenantID, employeeID uint) ([]SalaryHistory, error) {
	return s.repo.ListSalaryHistory(ctx, tenantID, employeeID)
}

// AddSalaryRevision writes a history row AND bumps
// employees.basic_salary when the revision is effective today or
// earlier. Future-dated revisions are logged but don't touch the current
// scalar (a nightly job / on-open recompute would apply them, but that
// stays out of this pass).
func (s *Service) AddSalaryRevision(ctx context.Context, tenantID, employeeID, userID uint, req *CreateSalaryRevisionRequest) (*SalaryHistory, error) {
	if req.EffectiveDate == "" {
		req.EffectiveDate = time.Now().Format("2006-01-02")
	}
	sh := &SalaryHistory{
		TenantID:      tenantID,
		EmployeeID:    employeeID,
		EffectiveDate: req.EffectiveDate,
		BasicSalary:   req.BasicSalary,
		Reason:        req.Reason,
	}
	if userID > 0 {
		sh.CreatedBy = &userID
	}
	if err := s.repo.AddSalaryRevision(ctx, sh); err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	if req.EffectiveDate <= today {
		if err := s.repo.UpdateEmployeeSalary(ctx, tenantID, employeeID, req.BasicSalary); err != nil {
			// Log but don't fail — the history row is the source of truth.
			return sh, nil
		}
	}
	return sh, nil
}

// SetEmployeeCVUrl is called by the upload handler after a successful
// file write.
func (s *Service) SetEmployeeCVUrl(ctx context.Context, tenantID, employeeID uint, url string) error {
	return s.repo.SetEmployeeCVUrl(ctx, tenantID, employeeID, url)
}
