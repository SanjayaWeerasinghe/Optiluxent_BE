package inventory

import (
	"context"
	"fmt"
	"strings"
)

var validLocationTypes = map[string]bool{
	"STORAGE": true, "RECEIVING": true, "SHIPPING": true, "QUALITY": true, "DAMAGED": true,
}
var validTxTypes = map[string]bool{
	"OPENING": true, "PURCHASE": true, "SALES": true, "PRODUCTION": true,
	"ADJUSTMENT": true, "TRANSFER_IN": true, "TRANSFER_OUT": true,
	"RETURN_IN": true, "RETURN_OUT": true,
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListWarehouses(ctx context.Context, tenantID uint) ([]Warehouse, error) {
	return s.repo.ListWarehouses(ctx, tenantID)
}

func (s *Service) GetWarehouse(ctx context.Context, tenantID, id uint) (*Warehouse, error) {
	return s.repo.GetWarehouse(ctx, tenantID, id)
}

func (s *Service) CreateWarehouse(ctx context.Context, tenantID uint, req *CreateWarehouseRequest) (*Warehouse, error) {
	w := &Warehouse{
		TenantID: tenantID,
		Code:     req.Code,
		Name:     req.Name,
		Address:  req.Address,
		IsActive: true,
	}
	if err := s.repo.CreateWarehouse(ctx, w); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("warehouse code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return w, nil
}

func (s *Service) UpdateWarehouse(ctx context.Context, tenantID, id uint, req *UpdateWarehouseRequest) (*Warehouse, error) {
	w, err := s.repo.GetWarehouse(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found")
	}
	if req.Name != "" {
		w.Name = req.Name
	}
	if req.Address != "" {
		w.Address = req.Address
	}
	if req.IsActive != nil {
		w.IsActive = *req.IsActive
	}
	return w, s.repo.UpdateWarehouse(ctx, w)
}

func (s *Service) DeleteWarehouse(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetWarehouse(ctx, tenantID, id); err != nil {
		return fmt.Errorf("warehouse not found")
	}
	return s.repo.DeleteWarehouse(ctx, tenantID, id)
}

func (s *Service) ListLocations(ctx context.Context, tenantID, warehouseID uint) ([]StorageLocation, error) {
	return s.repo.ListLocations(ctx, tenantID, warehouseID)
}

func (s *Service) CreateLocation(ctx context.Context, tenantID, warehouseID uint, req *CreateLocationRequest) (*StorageLocation, error) {
	locType := req.LocationType
	if locType == "" {
		locType = "STORAGE"
	}
	if !validLocationTypes[locType] {
		return nil, fmt.Errorf("invalid location_type: %s", locType)
	}
	l := &StorageLocation{
		TenantID:     tenantID,
		WarehouseID:  warehouseID,
		Code:         req.Code,
		Name:         req.Name,
		LocationType: locType,
		IsActive:     true,
	}
	if err := s.repo.CreateLocation(ctx, l); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("location code '%s' already exists in this warehouse", req.Code)
		}
		return nil, err
	}
	return l, nil
}

func (s *Service) UpdateLocation(ctx context.Context, tenantID uint, req *UpdateLocationRequest, existing *StorageLocation) (*StorageLocation, error) {
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	return existing, s.repo.UpdateLocation(ctx, existing)
}

func (s *Service) CreateStockEntry(ctx context.Context, tenantID uint, req *CreateStockEntryRequest) (*StockLedger, error) {
	if !validTxTypes[req.TransactionType] {
		return nil, fmt.Errorf("invalid transaction_type: %s", req.TransactionType)
	}
	e := &StockLedger{
		TenantID:        tenantID,
		ProductID:       req.ProductID,
		VariantID:       req.VariantID,
		WarehouseID:     req.WarehouseID,
		LocationID:      req.LocationID,
		TransactionType: req.TransactionType,
		ReferenceType:   req.ReferenceType,
		ReferenceID:     req.ReferenceID,
		Quantity:        req.Quantity,
		UnitCost:        req.UnitCost,
		TotalCost:       req.Quantity * req.UnitCost,
		TransactionDate: req.TransactionDate,
		Notes:           req.Notes,
	}
	return e, s.repo.CreateStockEntry(ctx, e)
}

func (s *Service) GetStockBalance(ctx context.Context, tenantID, warehouseID uint, productID *uint) ([]StockBalance, error) {
	return s.repo.GetStockBalance(ctx, tenantID, warehouseID, productID)
}

func (s *Service) ListStockLedger(ctx context.Context, tenantID, productID, warehouseID uint, limit int) ([]StockLedger, error) {
	return s.repo.ListStockLedger(ctx, tenantID, productID, warehouseID, limit)
}
