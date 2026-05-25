package manufacturing

import (
	"context"
	"fmt"
	"strings"
)

var validBOMTypes = map[string]bool{"MANUFACTURE": true, "KIT": true, "SUBCONTRACT": true}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ── BOMs ──────────────────────────────────────────────────────────────────────

func (s *Service) ListBOMs(ctx context.Context, tenantID uint) ([]BillOfMaterials, error) {
	return s.repo.ListBOMs(ctx, tenantID)
}

func (s *Service) GetBOM(ctx context.Context, tenantID, id uint) (*BillOfMaterials, error) {
	return s.repo.GetBOM(ctx, tenantID, id)
}

func (s *Service) CreateBOM(ctx context.Context, tenantID uint, req *CreateBOMRequest) (*BillOfMaterials, error) {
	bomType := req.BOMType
	if bomType == "" {
		bomType = "MANUFACTURE"
	}
	if !validBOMTypes[bomType] {
		return nil, fmt.Errorf("invalid bom_type: %s", bomType)
	}
	qty := req.Quantity
	if qty == 0 {
		qty = 1
	}
	b := &BillOfMaterials{
		TenantID:  tenantID,
		Code:      req.Code,
		ProductID: req.ProductID,
		VariantID: req.VariantID,
		Quantity:  qty,
		UOMID:     req.UOMID,
		BOMType:   bomType,
		IsActive:  true,
		Notes:     req.Notes,
	}
	if err := s.repo.CreateBOM(ctx, b); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("BOM code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return b, nil
}

func (s *Service) UpdateBOM(ctx context.Context, tenantID, id uint, req *UpdateBOMRequest) (*BillOfMaterials, error) {
	b, err := s.repo.GetBOM(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("BOM not found")
	}
	if req.IsActive != nil {
		b.IsActive = *req.IsActive
	}
	if req.Notes != "" {
		b.Notes = req.Notes
	}
	if req.Quantity != 0 {
		b.Quantity = req.Quantity
	}
	return b, s.repo.UpdateBOM(ctx, b)
}

func (s *Service) DeleteBOM(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetBOM(ctx, tenantID, id); err != nil {
		return fmt.Errorf("BOM not found")
	}
	return s.repo.DeleteBOM(ctx, tenantID, id)
}

func (s *Service) AddBOMLine(ctx context.Context, tenantID, bomID uint, req *AddBOMLineRequest) (*BOMLine, error) {
	seq := req.Sequence
	if seq == 0 {
		seq = 10
	}
	line := &BOMLine{
		BOMID:       bomID,
		TenantID:    tenantID,
		ComponentID: req.ComponentID,
		VariantID:   req.VariantID,
		Quantity:    req.Quantity,
		UOMID:       req.UOMID,
		ScrapPct:    req.ScrapPct,
		Sequence:    seq,
		Notes:       req.Notes,
	}
	return line, s.repo.AddBOMLine(ctx, line)
}

func (s *Service) DeleteBOMLine(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteBOMLine(ctx, tenantID, id)
}

// ── Work Centers ──────────────────────────────────────────────────────────────

func (s *Service) ListWorkCenters(ctx context.Context, tenantID uint) ([]WorkCenter, error) {
	return s.repo.ListWorkCenters(ctx, tenantID)
}

func (s *Service) CreateWorkCenter(ctx context.Context, tenantID uint, req *CreateWorkCenterRequest) (*WorkCenter, error) {
	cap := req.Capacity
	if cap == 0 {
		cap = 1
	}
	wc := &WorkCenter{
		TenantID:    tenantID,
		Code:        req.Code,
		Name:        req.Name,
		Capacity:    cap,
		CostPerHour: req.CostPerHour,
		CurrencyID:  req.CurrencyID,
		IsActive:    true,
		Notes:       req.Notes,
	}
	if err := s.repo.CreateWorkCenter(ctx, wc); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("work center code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return wc, nil
}

func (s *Service) UpdateWorkCenter(ctx context.Context, tenantID, id uint, req *UpdateWorkCenterRequest) (*WorkCenter, error) {
	wc, err := s.repo.GetWorkCenter(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("work center not found")
	}
	if req.Name != "" {
		wc.Name = req.Name
	}
	if req.Capacity != 0 {
		wc.Capacity = req.Capacity
	}
	if req.CostPerHour != nil {
		wc.CostPerHour = *req.CostPerHour
	}
	if req.IsActive != nil {
		wc.IsActive = *req.IsActive
	}
	if req.Notes != "" {
		wc.Notes = req.Notes
	}
	return wc, s.repo.UpdateWorkCenter(ctx, wc)
}

func (s *Service) DeleteWorkCenter(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetWorkCenter(ctx, tenantID, id); err != nil {
		return fmt.Errorf("work center not found")
	}
	return s.repo.DeleteWorkCenter(ctx, tenantID, id)
}

// ── Routings ──────────────────────────────────────────────────────────────────

func (s *Service) ListRoutings(ctx context.Context, tenantID uint) ([]Routing, error) {
	return s.repo.ListRoutings(ctx, tenantID)
}

func (s *Service) GetRouting(ctx context.Context, tenantID, id uint) (*Routing, error) {
	return s.repo.GetRouting(ctx, tenantID, id)
}

func (s *Service) CreateRouting(ctx context.Context, tenantID uint, req *CreateRoutingRequest) (*Routing, error) {
	rt := &Routing{
		TenantID:  tenantID,
		Code:      req.Code,
		ProductID: req.ProductID,
		IsActive:  true,
		Notes:     req.Notes,
	}
	if err := s.repo.CreateRouting(ctx, rt); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("routing code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return rt, nil
}

func (s *Service) DeleteRouting(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetRouting(ctx, tenantID, id); err != nil {
		return fmt.Errorf("routing not found")
	}
	return s.repo.DeleteRouting(ctx, tenantID, id)
}

func (s *Service) AddRoutingOperation(ctx context.Context, tenantID, routingID uint, req *AddOperationRequest) (*RoutingOperation, error) {
	seq := req.Sequence
	if seq == 0 {
		seq = 10
	}
	op := &RoutingOperation{
		RoutingID:    routingID,
		TenantID:     tenantID,
		Sequence:     seq,
		Name:         req.Name,
		WorkCenterID: req.WorkCenterID,
		SetupTime:    req.SetupTime,
		CycleTime:    req.CycleTime,
		Notes:        req.Notes,
	}
	return op, s.repo.AddRoutingOperation(ctx, op)
}

func (s *Service) DeleteRoutingOperation(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeleteRoutingOperation(ctx, tenantID, id)
}
