package materials

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ── Core Material ─────────────────────────────────────────────────────────────

func (s *Service) Count(ctx context.Context, tenantID uint, materialType string) (int64, error) {
	return s.repo.Count(ctx, tenantID, materialType)
}

func (s *Service) List(ctx context.Context, tenantID uint, materialType string, limit, offset int) ([]Material, error) {
	return s.repo.List(ctx, tenantID, materialType, limit, offset)
}

func (s *Service) Get(ctx context.Context, tenantID, id uint) (*MaterialDetail, error) {
	d, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("material not found")
	}
	return d, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req *CreateMaterialRequest) (*MaterialDetail, error) {
	matType := req.MaterialType
	if matType == "" {
		matType = "RAW_MATERIAL"
	}
	m := &Material{
		TenantID:     tenantID,
		Code:         req.Code,
		Name:         req.Name,
		MaterialType: matType,
		Color:        req.Color,
		CategoryID:   req.CategoryID,
		ArticleCode:  req.ArticleCode,
		Ref2:         req.Ref2,
		Barcode:      req.Barcode,
		Description:  req.Description,
		Notes:        req.Notes,
		IsActive:     true,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("material code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return s.repo.Get(ctx, tenantID, m.ID)
}

func (s *Service) Update(ctx context.Context, tenantID, id uint, req *UpdateMaterialRequest) (*MaterialDetail, error) {
	d, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("material not found")
	}
	m := &d.Material
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.MaterialType != "" {
		m.MaterialType = req.MaterialType
	}
	if req.Color != "" {
		m.Color = req.Color
	}
	m.CategoryID = req.CategoryID
	if req.ArticleCode != "" {
		m.ArticleCode = req.ArticleCode
	}
	if req.Ref2 != "" {
		m.Ref2 = req.Ref2
	}
	if req.Barcode != "" {
		m.Barcode = req.Barcode
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}
	if req.Notes != "" {
		m.Notes = req.Notes
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, tenantID, id)
}

func (s *Service) Delete(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.Get(ctx, tenantID, id); err != nil {
		return fmt.Errorf("material not found")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// ── Purchasing ────────────────────────────────────────────────────────────────

func (s *Service) GetPurchasing(ctx context.Context, tenantID, materialID uint) (*MaterialPurchasing, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	return s.repo.GetPurchasing(ctx, materialID)
}

func (s *Service) UpsertPurchasing(ctx context.Context, tenantID, materialID uint, req *UpsertPurchasingRequest) (*MaterialPurchasing, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	p := &MaterialPurchasing{
		MaterialID:       materialID,
		TenantID:         tenantID,
		PurchasingUOMID:  req.PurchasingUOMID,
		UnderDeliveryPct: req.UnderDeliveryPct,
		OverDeliveryPct:  req.OverDeliveryPct,
	}
	if err := s.repo.UpsertPurchasing(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.GetPurchasing(ctx, materialID)
}

// ── Manufacturing ─────────────────────────────────────────────────────────────

func (s *Service) GetManufacturing(ctx context.Context, tenantID, materialID uint) (*MaterialManufacturing, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	return s.repo.GetManufacturing(ctx, materialID)
}

func (s *Service) UpsertManufacturing(ctx context.Context, tenantID, materialID uint, req *UpsertManufacturingRequest) (*MaterialManufacturing, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	m := &MaterialManufacturing{
		MaterialID:            materialID,
		TenantID:              tenantID,
		ProductionUOMID:       req.ProductionUOMID,
		ReorderQtyLevel:       req.ReorderQtyLevel,
		SafetyLevel:           req.SafetyLevel,
		ProductionDays:        req.ProductionDays,
		DeliveryDays:          req.DeliveryDays,
		GRNDays:               req.GRNDays,
		ProcurementRepeatDays: req.ProcurementRepeatDays,
	}
	if err := s.repo.UpsertManufacturing(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.GetManufacturing(ctx, materialID)
}

// ── Warehouse ─────────────────────────────────────────────────────────────────

func (s *Service) GetWarehouse(ctx context.Context, tenantID, materialID uint) (*MaterialWarehouse, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	return s.repo.GetWarehouse(ctx, materialID)
}

func (s *Service) UpsertWarehouse(ctx context.Context, tenantID, materialID uint, req *UpsertWarehouseRequest) (*MaterialWarehouse, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	stockRemoval := req.StockRemoval
	if stockRemoval == "" {
		stockRemoval = "FIFO"
	}
	w := &MaterialWarehouse{
		MaterialID:          materialID,
		TenantID:            tenantID,
		StockingUOMID:       req.StockingUOMID,
		StockRemoval:        stockRemoval,
		StorageMain:         req.StorageMain,
		StorageDamaged:      req.StorageDamaged,
		StorageHold:         req.StorageHold,
		BatchProcess:        req.BatchProcess,
		ProductionDateCheck: req.ProductionDateCheck,
		ExpiryDateCheck:     req.ExpiryDateCheck,
		QCCheck:             req.QCCheck,
		GRNWithPOUOMImage:   req.GRNWithPOUOMImage,
	}
	if err := s.repo.UpsertWarehouse(ctx, w); err != nil {
		return nil, err
	}
	return s.repo.GetWarehouse(ctx, materialID)
}

// ── Vendors ───────────────────────────────────────────────────────────────────

func (s *Service) ListVendors(ctx context.Context, tenantID, materialID uint) ([]MaterialVendor, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	return s.repo.ListVendors(ctx, materialID)
}

func (s *Service) AddVendor(ctx context.Context, tenantID, materialID uint, req *VendorRequest) (*MaterialVendor, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	v := &MaterialVendor{
		MaterialID:     materialID,
		TenantID:       tenantID,
		VendorID:       req.VendorID,
		ArticleNo:      req.ArticleNo,
		DeliveryDays:   req.DeliveryDays,
		Cost:           req.Cost,
		CurrencyID:     req.CurrencyID,
		ProjectedPrice: req.ProjectedPrice,
		MOQ:            req.MOQ,
		IsDefault:      req.IsDefault,
	}
	if err := s.repo.CreateVendor(ctx, v); err != nil {
		return nil, err
	}
	vendors, _ := s.repo.ListVendors(ctx, materialID)
	for _, vv := range vendors {
		if vv.ID == v.ID {
			result := vv
			return &result, nil
		}
	}
	return v, nil
}

func (s *Service) UpdateVendor(ctx context.Context, tenantID, materialID, vendorRowID uint, req *VendorRequest) (*MaterialVendor, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	v := &MaterialVendor{
		ID:             vendorRowID,
		MaterialID:     materialID,
		TenantID:       tenantID,
		VendorID:       req.VendorID,
		ArticleNo:      req.ArticleNo,
		DeliveryDays:   req.DeliveryDays,
		Cost:           req.Cost,
		CurrencyID:     req.CurrencyID,
		ProjectedPrice: req.ProjectedPrice,
		MOQ:            req.MOQ,
		IsDefault:      req.IsDefault,
	}
	if err := s.repo.UpdateVendor(ctx, v); err != nil {
		return nil, err
	}
	vendors, _ := s.repo.ListVendors(ctx, materialID)
	for _, vv := range vendors {
		if vv.ID == v.ID {
			result := vv
			return &result, nil
		}
	}
	return v, nil
}

func (s *Service) DeleteVendor(ctx context.Context, tenantID, materialID, vendorRowID uint) error {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return fmt.Errorf("material not found")
	}
	return s.repo.DeleteVendor(ctx, materialID, vendorRowID)
}

func (s *Service) ReplaceVendors(ctx context.Context, tenantID, materialID uint, reqs []VendorRequest) ([]MaterialVendor, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	vendors := make([]MaterialVendor, 0, len(reqs))
	for _, r := range reqs {
		vendors = append(vendors, MaterialVendor{
			VendorID: r.VendorID, ArticleNo: r.ArticleNo, DeliveryDays: r.DeliveryDays,
			Cost: r.Cost, CurrencyID: r.CurrencyID, ProjectedPrice: r.ProjectedPrice,
			MOQ: r.MOQ, IsDefault: r.IsDefault,
		})
	}
	if err := s.repo.ReplaceVendors(ctx, materialID, tenantID, vendors); err != nil {
		return nil, err
	}
	return s.repo.ListVendors(ctx, materialID)
}

// ── Measurements ──────────────────────────────────────────────────────────────

func (s *Service) ListMeasurements(ctx context.Context, tenantID, materialID uint) ([]MaterialMeasurement, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	return s.repo.ListMeasurements(ctx, materialID)
}

func (s *Service) AddMeasurement(ctx context.Context, tenantID, materialID uint, req *MeasurementRequest) (*MaterialMeasurement, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	m := &MaterialMeasurement{
		MaterialID:      materialID,
		TenantID:        tenantID,
		BaseUOMID:       req.BaseUOMID,
		TargetUOMID:     req.TargetUOMID,
		ConversionRatio: req.ConversionRatio,
	}
	if err := s.repo.CreateMeasurement(ctx, m); err != nil {
		return nil, err
	}
	rows, _ := s.repo.ListMeasurements(ctx, materialID)
	for _, r := range rows {
		if r.ID == m.ID {
			result := r
			return &result, nil
		}
	}
	return m, nil
}

func (s *Service) UpdateMeasurement(ctx context.Context, tenantID, materialID, measurementID uint, req *MeasurementRequest) (*MaterialMeasurement, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	m := &MaterialMeasurement{
		ID:              measurementID,
		MaterialID:      materialID,
		TenantID:        tenantID,
		BaseUOMID:       req.BaseUOMID,
		TargetUOMID:     req.TargetUOMID,
		ConversionRatio: req.ConversionRatio,
	}
	if err := s.repo.UpdateMeasurement(ctx, m); err != nil {
		return nil, err
	}
	rows, _ := s.repo.ListMeasurements(ctx, materialID)
	for _, r := range rows {
		if r.ID == m.ID {
			result := r
			return &result, nil
		}
	}
	return m, nil
}

func (s *Service) DeleteMeasurement(ctx context.Context, tenantID, materialID, measurementID uint) error {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return fmt.Errorf("material not found")
	}
	return s.repo.DeleteMeasurement(ctx, materialID, measurementID)
}

func (s *Service) ReplaceMeasurements(ctx context.Context, tenantID, materialID uint, reqs []MeasurementRequest) ([]MaterialMeasurement, error) {
	if _, err := s.repo.Get(ctx, tenantID, materialID); err != nil {
		return nil, fmt.Errorf("material not found")
	}
	items := make([]MaterialMeasurement, 0, len(reqs))
	for _, r := range reqs {
		items = append(items, MaterialMeasurement{
			BaseUOMID: r.BaseUOMID, TargetUOMID: r.TargetUOMID, ConversionRatio: r.ConversionRatio,
		})
	}
	if err := s.repo.ReplaceMeasurements(ctx, materialID, tenantID, items); err != nil {
		return nil, err
	}
	return s.repo.ListMeasurements(ctx, materialID)
}
