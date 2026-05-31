package mrp

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, tenantID uint, materialType string) ([]MaterialMasterView, error) {
	return s.repo.List(ctx, tenantID, materialType)
}

func (s *Service) Get(ctx context.Context, tenantID, id uint) (*MaterialMasterView, error) {
	m, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("material master not found")
	}
	return m, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req *CreateMaterialMasterRequest) (*MaterialMaster, error) {
	matType := req.MaterialType
	if matType == "" {
		matType = "RAW_MATERIAL"
	}
	mrpType := req.MRPType
	if mrpType == "" {
		mrpType = "MRP"
	}
	procType := req.ProcurementType
	if procType == "" {
		procType = "EXTERNAL"
	}
	lotType := req.LotSizeType
	if lotType == "" {
		lotType = "LOT_FOR_LOT"
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	m := &MaterialMaster{
		TenantID:           tenantID,
		ProductID:          req.ProductID,
		MaterialType:       matType,
		MRPType:            mrpType,
		ProcurementType:    procType,
		PurchaseLeadTime:   req.PurchaseLeadTime,
		ProductionLeadTime: req.ProductionLeadTime,
		SafetyStock:        req.SafetyStock,
		ReorderPoint:       req.ReorderPoint,
		MinOrderQty:        req.MinOrderQty,
		MaxOrderQty:        req.MaxOrderQty,
		LotSizeType:        lotType,
		FixedLotSize:       req.FixedLotSize,
		ABCClass:           req.ABCClass,
		ShelfLifeDays:      req.ShelfLifeDays,
		IsActive:           isActive,
		Notes:              req.Notes,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("material master already exists for this product")
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) Update(ctx context.Context, tenantID, id uint, req *UpdateMaterialMasterRequest) (*MaterialMasterView, error) {
	m, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("material master not found")
	}
	if req.MaterialType != "" {
		m.MaterialType = req.MaterialType
	}
	if req.MRPType != "" {
		m.MRPType = req.MRPType
	}
	if req.ProcurementType != "" {
		m.ProcurementType = req.ProcurementType
	}
	if req.PurchaseLeadTime != nil {
		m.PurchaseLeadTime = *req.PurchaseLeadTime
	}
	if req.ProductionLeadTime != nil {
		m.ProductionLeadTime = *req.ProductionLeadTime
	}
	if req.SafetyStock != nil {
		m.SafetyStock = *req.SafetyStock
	}
	if req.ReorderPoint != nil {
		m.ReorderPoint = *req.ReorderPoint
	}
	if req.MinOrderQty != nil {
		m.MinOrderQty = *req.MinOrderQty
	}
	if req.MaxOrderQty != nil {
		m.MaxOrderQty = req.MaxOrderQty
	}
	if req.LotSizeType != "" {
		m.LotSizeType = req.LotSizeType
	}
	if req.FixedLotSize != nil {
		m.FixedLotSize = req.FixedLotSize
	}
	if req.ABCClass != "" {
		m.ABCClass = req.ABCClass
	}
	if req.ShelfLifeDays != nil {
		m.ShelfLifeDays = req.ShelfLifeDays
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}
	if req.Notes != "" {
		m.Notes = req.Notes
	}
	if err := s.repo.Update(ctx, &m.MaterialMaster); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.Get(ctx, tenantID, id); err != nil {
		return fmt.Errorf("material master not found")
	}
	return s.repo.Delete(ctx, tenantID, id)
}
