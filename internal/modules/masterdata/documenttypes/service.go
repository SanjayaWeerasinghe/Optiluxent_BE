package documenttypes

import (
	"context"
	"fmt"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ── Types ────────────────────────────────────────────────────────────────────

func (s *Service) ListTypes(ctx context.Context, tenantID uint, model string) ([]DocumentType, error) {
	return s.repo.ListTypes(ctx, tenantID, model)
}

func (s *Service) GetType(ctx context.Context, tenantID, id uint) (*DocumentType, error) {
	return s.repo.GetTypeWithFields(ctx, tenantID, id)
}

func (s *Service) CreateType(ctx context.Context, tenantID uint, req *CreateTypeRequest) (*DocumentType, error) {
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	t := &DocumentType{
		TenantID: tenantID,
		Model:    req.Model,
		Code:     req.Code,
		Name:     req.Name,
		IsActive: active,
	}
	if err := s.repo.CreateType(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) UpdateType(ctx context.Context, tenantID, id uint, req *UpdateTypeRequest) (*DocumentType, error) {
	t, err := s.repo.GetType(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("document type not found: %w", err)
	}
	if req.Code != "" {
		t.Code = req.Code
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	return t, s.repo.UpdateType(ctx, t)
}

func (s *Service) DeleteType(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteType(ctx, tenantID, id)
}

// ResolveSystemKey returns the seeded `system_key` for a Type id, or empty
// string if the Type is user-defined (no system behaviour attached). Used by
// procurement.ConfirmGRN + inventory.CreateAutoQC to keep their existing
// switches working after the enum columns were dropped.
func (s *Service) ResolveSystemKey(ctx context.Context, tenantID, id uint) (string, error) {
	if id == 0 {
		return "", nil
	}
	t, err := s.repo.GetType(ctx, tenantID, id)
	if err != nil {
		return "", err
	}
	if t.SystemKey == nil {
		return "", nil
	}
	return *t.SystemKey, nil
}

// FindSystemType returns the DocumentType id for a (model, system_key) pair.
// Zero if no match — callers should treat that as "not seeded yet" and skip.
func (s *Service) FindSystemType(ctx context.Context, tenantID uint, model, systemKey string) (uint, error) {
	return s.repo.FindSystemType(ctx, tenantID, model, systemKey)
}

// ── Fields ───────────────────────────────────────────────────────────────────

func (s *Service) ListFields(ctx context.Context, tenantID, typeID uint) ([]DocumentTypeField, error) {
	return s.repo.ListFields(ctx, tenantID, typeID)
}

func (s *Service) CreateField(ctx context.Context, tenantID, typeID uint, req *CreateFieldRequest) (*DocumentTypeField, error) {
	// Make sure the parent Type exists in this tenant.
	if _, err := s.repo.GetType(ctx, tenantID, typeID); err != nil {
		return nil, fmt.Errorf("document type not found: %w", err)
	}
	f := &DocumentTypeField{
		TenantID:       tenantID,
		DocumentTypeID: typeID,
		Code:           req.Code,
		Label:          req.Label,
		Kind:           req.Kind,
		IsRequired:     req.IsRequired,
		DisplayOrder:   req.DisplayOrder,
	}
	if err := s.repo.CreateField(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) UpdateField(ctx context.Context, tenantID, id uint, req *UpdateFieldRequest) (*DocumentTypeField, error) {
	f, err := s.repo.GetField(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("field not found: %w", err)
	}
	if req.Code != "" {
		f.Code = req.Code
	}
	if req.Label != "" {
		f.Label = req.Label
	}
	if req.Kind != "" {
		f.Kind = req.Kind
	}
	if req.IsRequired != nil {
		f.IsRequired = *req.IsRequired
	}
	if req.DisplayOrder != nil {
		f.DisplayOrder = *req.DisplayOrder
	}
	return f, s.repo.UpdateField(ctx, f)
}

func (s *Service) DeleteField(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteField(ctx, tenantID, id)
}

// ── Doc-side values ──────────────────────────────────────────────────────────

// GetValuesForDoc returns the enriched field-value snapshot for a document.
// Enrichment joins each value against its DocumentTypeField so the FE gets
// label + kind + code without needing a second round-trip.
func (s *Service) GetValuesForDoc(ctx context.Context, tenantID uint, docKind string, docID, typeID uint) ([]ValueResponse, error) {
	fields, err := s.repo.ListFields(ctx, tenantID, typeID)
	if err != nil {
		return nil, err
	}
	valMap := map[uint]uint{}
	if docID != 0 {
		vals, err := s.repo.ListValuesForDoc(ctx, tenantID, docKind, docID)
		if err != nil {
			return nil, err
		}
		for _, v := range vals {
			refCopy := v.RefID
			valMap[v.FieldID] = refCopy
		}
	}
	out := make([]ValueResponse, 0, len(fields))
	for _, f := range fields {
		row := ValueResponse{FieldID: f.ID, Code: f.Code, Label: f.Label, Kind: f.Kind}
		if ref, ok := valMap[f.ID]; ok {
			refCopy := ref
			row.RefID = &refCopy
		}
		out = append(out, row)
	}
	return out, nil
}

// UpsertValuesForDoc replaces the value set for a document with the given
// batch. Nil ref_id clears a slot.
func (s *Service) UpsertValuesForDoc(ctx context.Context, tenantID uint, docKind string, docID uint, entries []ValueEntry) error {
	for _, e := range entries {
		if e.RefID == nil || *e.RefID == 0 {
			if err := s.repo.DeleteValue(ctx, tenantID, docKind, docID, e.FieldID); err != nil {
				return err
			}
			continue
		}
		v := &DocumentFieldValue{
			TenantID: tenantID,
			DocKind:  docKind,
			DocID:    docID,
			FieldID:  e.FieldID,
			RefID:    *e.RefID,
		}
		if err := s.repo.UpsertValue(ctx, v); err != nil {
			return err
		}
	}
	return nil
}
