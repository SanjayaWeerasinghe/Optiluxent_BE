package contacts

import (
	"context"
	"fmt"
	"strings"
)

var validPartyTypes = map[string]bool{"CUSTOMER": true, "SUPPLIER": true, "BOTH": true}
var validTypes = map[string]bool{"INDIVIDUAL": true, "COMPANY": true}
var validAddressTypes = map[string]bool{"BILLING": true, "SHIPPING": true, "BOTH": true}

// defaultCreditType returns CREDIT when the caller doesn't specify one —
// matches the DB DEFAULT so behaviour is consistent whether the field is
// omitted from the request or explicitly left blank.
func defaultCreditType(v string) string {
	if v == "" {
		return "CREDIT"
	}
	return v
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CountParties(ctx context.Context, tenantID uint, partyType string, activeOnly bool) (int64, error) {
	if partyType != "" && !validPartyTypes[partyType] {
		return 0, fmt.Errorf("invalid party_type: %s", partyType)
	}
	return s.repo.CountParties(ctx, tenantID, partyType, activeOnly)
}

func (s *Service) ListParties(ctx context.Context, tenantID uint, partyType string, activeOnly bool, limit, offset int) ([]Party, error) {
	if partyType != "" && !validPartyTypes[partyType] {
		return nil, fmt.Errorf("invalid party_type: %s", partyType)
	}
	return s.repo.ListParties(ctx, tenantID, partyType, activeOnly, limit, offset)
}

func (s *Service) GetParty(ctx context.Context, tenantID, id uint) (*Party, error) {
	return s.repo.GetParty(ctx, tenantID, id)
}

func (s *Service) CreateParty(ctx context.Context, tenantID uint, req *CreatePartyRequest) (*Party, error) {
	if !validPartyTypes[req.PartyType] {
		return nil, fmt.Errorf("invalid party_type: %s", req.PartyType)
	}
	if req.Type != "" && !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid type: %s", req.Type)
	}
	pType := req.Type
	if pType == "" {
		pType = "COMPANY"
	}
	p := &Party{
		TenantID:      tenantID,
		Code:          req.Code,
		Name:          req.Name,
		LegalName:     req.LegalName,
		PartyType:     req.PartyType,
		Type:          pType,
		TaxRegNumber:  req.TaxRegNumber,
		CurrencyID:    req.CurrencyID,
		PaymentTermID: req.PaymentTermID,
		CreditLimit:   req.CreditLimit,
		CreditType:    defaultCreditType(req.CreditType),
		IsActive:      true,
		Notes:         req.Notes,
	}
	if err := s.repo.CreateParty(ctx, p); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("party code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return p, nil
}

func (s *Service) UpdateParty(ctx context.Context, tenantID, id uint, req *UpdatePartyRequest) (*Party, error) {
	p, err := s.repo.GetParty(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("party not found")
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.LegalName != "" {
		p.LegalName = req.LegalName
	}
	if req.PartyType != "" {
		if !validPartyTypes[req.PartyType] {
			return nil, fmt.Errorf("invalid party_type: %s", req.PartyType)
		}
		p.PartyType = req.PartyType
	}
	if req.TaxRegNumber != "" {
		p.TaxRegNumber = req.TaxRegNumber
	}
	if req.CurrencyID != 0 {
		p.CurrencyID = req.CurrencyID
	}
	p.PaymentTermID = req.PaymentTermID
	if req.CreditLimit != nil {
		p.CreditLimit = *req.CreditLimit
	}
	if req.CreditType != "" {
		p.CreditType = req.CreditType
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if req.Notes != "" {
		p.Notes = req.Notes
	}
	return p, s.repo.UpdateParty(ctx, p)
}

func (s *Service) DeleteParty(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetParty(ctx, tenantID, id); err != nil {
		return fmt.Errorf("party not found")
	}
	return s.repo.DeleteParty(ctx, tenantID, id)
}

func (s *Service) ListContactPersons(ctx context.Context, tenantID, partyID uint) ([]ContactPerson, error) {
	return s.repo.ListContactPersons(ctx, tenantID, partyID)
}

func (s *Service) CreateContactPerson(ctx context.Context, tenantID, partyID uint, req *CreateContactPersonRequest) (*ContactPerson, error) {
	cp := &ContactPerson{
		TenantID:    tenantID,
		PartyID:     partyID,
		Name:        req.Name,
		Designation: req.Designation,
		Phone:       req.Phone,
		Mobile:      req.Mobile,
		Email:       req.Email,
		IsPrimary:   req.IsPrimary,
		IsActive:    true,
	}
	return cp, s.repo.CreateContactPerson(ctx, cp)
}

func (s *Service) UpdateContactPerson(ctx context.Context, tenantID, id uint, req *UpdateContactPersonRequest) (*ContactPerson, error) {
	cp, err := s.repo.GetContactPerson(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("contact person not found")
	}
	if req.Name != "" {
		cp.Name = req.Name
	}
	if req.Designation != "" {
		cp.Designation = req.Designation
	}
	if req.Phone != "" {
		cp.Phone = req.Phone
	}
	if req.Mobile != "" {
		cp.Mobile = req.Mobile
	}
	if req.Email != "" {
		cp.Email = req.Email
	}
	if req.IsPrimary != nil {
		cp.IsPrimary = *req.IsPrimary
	}
	if req.IsActive != nil {
		cp.IsActive = *req.IsActive
	}
	return cp, s.repo.UpdateContactPerson(ctx, cp)
}

func (s *Service) DeleteContactPerson(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetContactPerson(ctx, tenantID, id); err != nil {
		return fmt.Errorf("contact person not found")
	}
	return s.repo.DeleteContactPerson(ctx, tenantID, id)
}

func (s *Service) ListAddresses(ctx context.Context, tenantID, partyID uint) ([]PartyAddress, error) {
	return s.repo.ListAddresses(ctx, tenantID, partyID)
}

func (s *Service) CreateAddress(ctx context.Context, tenantID, partyID uint, req *CreateAddressRequest) (*PartyAddress, error) {
	if !validAddressTypes[req.AddressType] {
		return nil, fmt.Errorf("invalid address_type: %s", req.AddressType)
	}
	a := &PartyAddress{
		TenantID:     tenantID,
		PartyID:      partyID,
		AddressType:  req.AddressType,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		StateID:      req.StateID,
		CountryID:    req.CountryID,
		PostalCode:   req.PostalCode,
		IsPrimary:    req.IsPrimary,
	}
	return a, s.repo.CreateAddress(ctx, a)
}

func (s *Service) UpdateAddress(ctx context.Context, tenantID, id uint, req *UpdateAddressRequest) (*PartyAddress, error) {
	a, err := s.repo.GetAddress(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("address not found")
	}
	if req.AddressType != "" {
		if !validAddressTypes[req.AddressType] {
			return nil, fmt.Errorf("invalid address_type: %s", req.AddressType)
		}
		a.AddressType = req.AddressType
	}
	if req.AddressLine1 != "" {
		a.AddressLine1 = req.AddressLine1
	}
	if req.AddressLine2 != "" {
		a.AddressLine2 = req.AddressLine2
	}
	if req.City != "" {
		a.City = req.City
	}
	a.StateID = req.StateID
	if req.CountryID != 0 {
		a.CountryID = req.CountryID
	}
	if req.PostalCode != "" {
		a.PostalCode = req.PostalCode
	}
	if req.IsPrimary != nil {
		a.IsPrimary = *req.IsPrimary
	}
	return a, s.repo.UpdateAddress(ctx, a)
}

func (s *Service) DeleteAddress(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetAddress(ctx, tenantID, id); err != nil {
		return fmt.Errorf("address not found")
	}
	return s.repo.DeleteAddress(ctx, tenantID, id)
}
