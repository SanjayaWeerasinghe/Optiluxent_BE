package hr

import (
	"context"
	"fmt"
	"strings"
)

var validGenders = map[string]bool{"MALE": true, "FEMALE": true, "OTHER": true}
var validEmploymentTypes = map[string]bool{"PERMANENT": true, "CONTRACT": true, "PART_TIME": true, "INTERN": true}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListJobPositions(ctx context.Context, tenantID uint) ([]JobPosition, error) {
	return s.repo.ListJobPositions(ctx, tenantID)
}

func (s *Service) CreateJobPosition(ctx context.Context, tenantID uint, req *CreateJobPositionRequest) (*JobPosition, error) {
	jp := &JobPosition{
		TenantID:     tenantID,
		Code:         req.Code,
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}
	if err := s.repo.CreateJobPosition(ctx, jp); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("job position code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return jp, nil
}

func (s *Service) UpdateJobPosition(ctx context.Context, tenantID, id uint, req *UpdateJobPositionRequest) (*JobPosition, error) {
	jp, err := s.repo.GetJobPosition(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("job position not found")
	}
	if req.Name != "" {
		jp.Name = req.Name
	}
	jp.DepartmentID = req.DepartmentID
	if req.IsActive != nil {
		jp.IsActive = *req.IsActive
	}
	return jp, s.repo.UpdateJobPosition(ctx, jp)
}

func (s *Service) ListEmployees(ctx context.Context, tenantID uint, activeOnly bool, departmentID *uint) ([]Employee, error) {
	return s.repo.ListEmployees(ctx, tenantID, activeOnly, departmentID)
}

func (s *Service) GetEmployee(ctx context.Context, tenantID, id uint) (*Employee, error) {
	return s.repo.GetEmployee(ctx, tenantID, id)
}

func (s *Service) CreateEmployee(ctx context.Context, tenantID uint, req *CreateEmployeeRequest) (*Employee, error) {
	if req.Gender != "" {
		if !validGenders[req.Gender] {
			return nil, fmt.Errorf("invalid gender: %s", req.Gender)
		}
	}
	empType := req.EmploymentType
	if empType == "" {
		empType = "PERMANENT"
	}
	if !validEmploymentTypes[empType] {
		return nil, fmt.Errorf("invalid employment_type: %s", empType)
	}
	var gender *string
	if req.Gender != "" {
		g := req.Gender
		gender = &g
	}
	e := &Employee{
		TenantID:       tenantID,
		Code:           req.Code,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DisplayName:    req.DisplayName,
		Gender:         gender,
		DateOfBirth:    req.DateOfBirth,
		NICNumber:      req.NICNumber,
		JobPositionID:  req.JobPositionID,
		DepartmentID:   req.DepartmentID,
		ManagerID:      req.ManagerID,
		EmploymentType: empType,
		DateJoined:     req.DateJoined,
		Email:          req.Email,
		Phone:          req.Phone,
		Mobile:         req.Mobile,
		BankID:         req.BankID,
		BankAccountNo:  req.BankAccountNo,
		BasicSalary:    req.BasicSalary,
		CurrencyID:     req.CurrencyID,
		IsActive:       true,
		Notes:          req.Notes,
	}
	if err := s.repo.CreateEmployee(ctx, e); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("employee code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return e, nil
}

func (s *Service) UpdateEmployee(ctx context.Context, tenantID, id uint, req *UpdateEmployeeRequest) (*Employee, error) {
	e, err := s.repo.GetEmployee(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("employee not found")
	}
	if req.FirstName != "" {
		e.FirstName = req.FirstName
	}
	if req.LastName != "" {
		e.LastName = req.LastName
	}
	if req.DisplayName != "" {
		e.DisplayName = req.DisplayName
	}
	if req.Gender != "" {
		if !validGenders[req.Gender] {
			return nil, fmt.Errorf("invalid gender: %s", req.Gender)
		}
		g := req.Gender
		e.Gender = &g
	}
	e.JobPositionID = req.JobPositionID
	e.DepartmentID = req.DepartmentID
	e.ManagerID = req.ManagerID
	if req.EmploymentType != "" {
		if !validEmploymentTypes[req.EmploymentType] {
			return nil, fmt.Errorf("invalid employment_type: %s", req.EmploymentType)
		}
		e.EmploymentType = req.EmploymentType
	}
	e.DateLeft = req.DateLeft
	if req.BasicSalary != nil {
		e.BasicSalary = *req.BasicSalary
	}
	if req.IsActive != nil {
		e.IsActive = *req.IsActive
	}
	if req.Email != "" {
		e.Email = req.Email
	}
	if req.Phone != "" {
		e.Phone = req.Phone
	}
	if req.Mobile != "" {
		e.Mobile = req.Mobile
	}
	e.BankID = req.BankID
	if req.BankAccountNo != "" {
		e.BankAccountNo = req.BankAccountNo
	}
	if req.Notes != "" {
		e.Notes = req.Notes
	}
	return e, s.repo.UpdateEmployee(ctx, e)
}

func (s *Service) DeleteEmployee(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetEmployee(ctx, tenantID, id); err != nil {
		return fmt.Errorf("employee not found")
	}
	return s.repo.DeleteEmployee(ctx, tenantID, id)
}
