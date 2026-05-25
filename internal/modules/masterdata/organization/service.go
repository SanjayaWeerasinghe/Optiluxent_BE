package organization

import (
	"context"
	"errors"
	"strings"
)

// Service handles business logic for the Organization submodule.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCompany(ctx context.Context) (*Company, error) {
	return s.repo.GetCompany(ctx)
}

func (s *Service) SaveCompany(ctx context.Context, c *Company) error {
	if c.FiscalYearStart < 1 || c.FiscalYearStart > 12 {
		return errors.New("fiscal_year_start must be between 1 and 12")
	}
	return s.repo.SaveCompany(ctx, c)
}

func (s *Service) ListDepartments(ctx context.Context, tenantID uint) ([]*Department, error) {
	return s.repo.ListDepartments(ctx, tenantID)
}

func (s *Service) GetDepartment(ctx context.Context, id, tenantID uint) (*Department, error) {
	return s.repo.GetDepartment(ctx, id, tenantID)
}

func (s *Service) CreateDepartment(ctx context.Context, d *Department) error {
	d.Code = strings.ToUpper(strings.TrimSpace(d.Code))
	if err := s.repo.CreateDepartment(ctx, d); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("department code already exists in this tenant")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateDepartment(ctx context.Context, d *Department) error {
	return s.repo.UpdateDepartment(ctx, d)
}

func (s *Service) DeleteDepartment(ctx context.Context, id, tenantID uint) error {
	existing, err := s.repo.GetDepartment(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("department not found")
	}
	return s.repo.DeleteDepartment(ctx, id, tenantID)
}

func (s *Service) ListFiscalYears(ctx context.Context, tenantID uint) ([]*FiscalYear, error) {
	return s.repo.ListFiscalYears(ctx, tenantID)
}

func (s *Service) GetFiscalYear(ctx context.Context, id, tenantID uint) (*FiscalYear, error) {
	return s.repo.GetFiscalYear(ctx, id, tenantID)
}

func (s *Service) CreateFiscalYear(ctx context.Context, fy *FiscalYear) error {
	if !fy.EndDate.After(fy.StartDate) {
		return errors.New("end_date must be after start_date")
	}
	return s.repo.CreateFiscalYear(ctx, fy)
}

func (s *Service) CloseFiscalYear(ctx context.Context, id, tenantID uint) error {
	fy, err := s.repo.GetFiscalYear(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if fy == nil {
		return errors.New("fiscal year not found")
	}
	if fy.IsClosed {
		return errors.New("fiscal year is already closed")
	}
	return s.repo.CloseFiscalYear(ctx, id, tenantID)
}

func (s *Service) ListAccountingPeriods(ctx context.Context, tenantID uint, fiscalYearID *uint) ([]*AccountingPeriod, error) {
	return s.repo.ListAccountingPeriods(ctx, tenantID, fiscalYearID)
}

func (s *Service) GetAccountingPeriod(ctx context.Context, id, tenantID uint) (*AccountingPeriod, error) {
	return s.repo.GetAccountingPeriod(ctx, id, tenantID)
}

func (s *Service) CloseAccountingPeriod(ctx context.Context, id, tenantID uint) error {
	p, err := s.repo.GetAccountingPeriod(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("accounting period not found")
	}
	if p.IsClosed {
		return errors.New("accounting period is already closed")
	}
	return s.repo.CloseAccountingPeriod(ctx, id, tenantID)
}

func (s *Service) ListDocumentSequences(ctx context.Context, tenantID uint) ([]*DocumentSequence, error) {
	return s.repo.ListDocumentSequences(ctx, tenantID)
}

func (s *Service) GetDocumentSequence(ctx context.Context, id, tenantID uint) (*DocumentSequence, error) {
	return s.repo.GetDocumentSequence(ctx, id, tenantID)
}

func (s *Service) CreateDocumentSequence(ctx context.Context, ds *DocumentSequence) error {
	ds.DocumentType = strings.ToUpper(strings.TrimSpace(ds.DocumentType))
	if err := s.repo.CreateDocumentSequence(ctx, ds); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("document sequence for this type already exists")
		}
		return err
	}
	return nil
}

func (s *Service) UpdateDocumentSequence(ctx context.Context, ds *DocumentSequence) error {
	return s.repo.UpdateDocumentSequence(ctx, ds)
}

func (s *Service) ListCountries(ctx context.Context) ([]*Country, error) {
	return s.repo.ListCountries(ctx)
}

func (s *Service) ListStates(ctx context.Context, countryID uint) ([]*State, error) {
	return s.repo.ListStates(ctx, countryID)
}
