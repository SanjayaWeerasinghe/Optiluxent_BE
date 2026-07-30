package products

import (
	"context"
	"fmt"
	"strings"
)

var validProductTypes = map[string]bool{
	"RAW_MATERIAL": true, "WIP": true, "FINISHED": true, "SERVICE": true, "CONSUMABLE": true,
}
var validUOMTypes = map[string]bool{
	"UNIT": true, "WEIGHT": true, "VOLUME": true, "LENGTH": true, "AREA": true, "TIME": true,
}
var validPriceTypes = map[string]bool{"SALES": true, "PURCHASE": true, "SPECIAL": true}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ── Categories ────────────────────────────────────────────────────────────────

func (s *Service) CountCategories(ctx context.Context, tenantID uint) (int64, error) {
	return s.repo.CountCategories(ctx, tenantID)
}

func (s *Service) ListCategories(ctx context.Context, tenantID uint, limit, offset int) ([]ProductCategory, error) {
	return s.repo.ListCategories(ctx, tenantID, limit, offset)
}

func (s *Service) CreateCategory(ctx context.Context, tenantID uint, req *CreateCategoryRequest) (*ProductCategory, error) {
	c := &ProductCategory{
		TenantID: tenantID,
		Code:     req.Code,
		Name:     req.Name,
		ParentID: req.ParentID,
		IsActive: true,
	}
	if err := s.repo.CreateCategory(ctx, c); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("category code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateCategory(ctx context.Context, tenantID, id uint, req *UpdateCategoryRequest) (*ProductCategory, error) {
	c, err := s.repo.GetCategory(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	if req.Name != "" {
		c.Name = req.Name
	}
	c.ParentID = req.ParentID
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	return c, s.repo.UpdateCategory(ctx, c)
}

func (s *Service) DeleteCategory(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetCategory(ctx, tenantID, id); err != nil {
		return fmt.Errorf("category not found")
	}
	return s.repo.DeleteCategory(ctx, tenantID, id)
}

// ── UOMs ──────────────────────────────────────────────────────────────────────

func (s *Service) CountUOMs(ctx context.Context, tenantID uint) (int64, error) {
	return s.repo.CountUOMs(ctx, tenantID)
}

func (s *Service) ListUOMs(ctx context.Context, tenantID uint, limit, offset int) ([]UnitOfMeasure, error) {
	return s.repo.ListUOMs(ctx, tenantID, limit, offset)
}

func (s *Service) CreateUOM(ctx context.Context, tenantID uint, req *CreateUOMRequest) (*UnitOfMeasure, error) {
	uomType := req.UOMType
	if uomType == "" {
		uomType = "UNIT"
	}
	if !validUOMTypes[uomType] {
		return nil, fmt.Errorf("invalid uom_type: %s", uomType)
	}
	factor := req.ConversionFactor
	if factor == 0 {
		factor = 1
	}
	u := &UnitOfMeasure{
		TenantID:         tenantID,
		Code:             req.Code,
		Name:             req.Name,
		UOMType:          uomType,
		BaseUOMID:        req.BaseUOMID,
		ConversionFactor: factor,
		IsActive:         true,
	}
	if err := s.repo.CreateUOM(ctx, u); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("UOM code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return u, nil
}

func (s *Service) UpdateUOM(ctx context.Context, tenantID, id uint, req *UpdateUOMRequest) (*UnitOfMeasure, error) {
	u, err := s.repo.GetUOM(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("UOM not found")
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.ConversionFactor != 0 {
		u.ConversionFactor = req.ConversionFactor
	}
	u.BaseUOMID = req.BaseUOMID
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}
	return u, s.repo.UpdateUOM(ctx, u)
}

// ── Products ──────────────────────────────────────────────────────────────────

func (s *Service) CountProducts(ctx context.Context, tenantID uint, productType string, activeOnly bool) (int64, error) {
	if productType != "" && !validProductTypes[productType] {
		return 0, fmt.Errorf("invalid product_type: %s", productType)
	}
	return s.repo.CountProducts(ctx, tenantID, productType, activeOnly)
}

func (s *Service) ListProducts(ctx context.Context, tenantID uint, productType string, activeOnly bool, limit, offset int) ([]Product, error) {
	if productType != "" && !validProductTypes[productType] {
		return nil, fmt.Errorf("invalid product_type: %s", productType)
	}
	return s.repo.ListProducts(ctx, tenantID, productType, activeOnly, limit, offset)
}

func (s *Service) GetProduct(ctx context.Context, tenantID, id uint) (*Product, error) {
	return s.repo.GetProduct(ctx, tenantID, id)
}

// Kind returns the product's Kind (PRODUCT / SERVICE / REFINING_INTAKE) —
// consumed by other modules via a thin ProductKindProvider interface to
// keep stock-hitting flows from accepting service items.
func (s *Service) Kind(ctx context.Context, tenantID, id uint) (string, error) {
	p, err := s.repo.GetProduct(ctx, tenantID, id)
	if err != nil {
		return "", err
	}
	return p.Kind, nil
}

func (s *Service) CreateProduct(ctx context.Context, tenantID uint, req *CreateProductRequest) (*Product, error) {
	pType := req.ProductType
	if pType == "" {
		pType = "FINISHED"
	}
	if !validProductTypes[pType] {
		return nil, fmt.Errorf("invalid product_type: %s", pType)
	}
	kind := req.Kind
	if kind == "" {
		kind = "PRODUCT"
	}
	p := &Product{
		TenantID:       tenantID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		ProductType:    pType,
		Kind:           kind,
		CategoryID:     req.CategoryID,
		BaseUOMID:      req.BaseUOMID,
		PurchaseUOMID:  req.PurchaseUOMID,
		SalesUOMID:     req.SalesUOMID,
		TaxCodeID:      req.TaxCodeID,
		CostPrice:      req.CostPrice,
		StandardPrice:  req.StandardPrice,
		MinStockQty:    req.MinStockQty,
		ReorderQty:     req.ReorderQty,
		LeadTimeDays:   req.LeadTimeDays,
		IsPurchased:    req.IsPurchased,
		IsSold:         req.IsSold,
		IsManufactured: req.IsManufactured,
		IsActive:       true,
		Notes:          req.Notes,
	}
	if err := s.repo.CreateProduct(ctx, p); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("product code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return p, nil
}

func (s *Service) UpdateProduct(ctx context.Context, tenantID, id uint, req *UpdateProductRequest) (*Product, error) {
	p, err := s.repo.GetProduct(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.ProductType != "" {
		if !validProductTypes[req.ProductType] {
			return nil, fmt.Errorf("invalid product_type: %s", req.ProductType)
		}
		p.ProductType = req.ProductType
	}
	if req.Kind != "" {
		p.Kind = req.Kind
	}
	p.CategoryID = req.CategoryID
	if req.BaseUOMID != 0 {
		p.BaseUOMID = req.BaseUOMID
	}
	p.PurchaseUOMID = req.PurchaseUOMID
	p.SalesUOMID = req.SalesUOMID
	p.TaxCodeID = req.TaxCodeID
	if req.CostPrice != nil {
		p.CostPrice = *req.CostPrice
	}
	if req.StandardPrice != nil {
		p.StandardPrice = *req.StandardPrice
	}
	if req.MinStockQty != nil {
		p.MinStockQty = *req.MinStockQty
	}
	if req.ReorderQty != nil {
		p.ReorderQty = *req.ReorderQty
	}
	if req.LeadTimeDays != nil {
		p.LeadTimeDays = *req.LeadTimeDays
	}
	if req.IsPurchased != nil {
		p.IsPurchased = *req.IsPurchased
	}
	if req.IsSold != nil {
		p.IsSold = *req.IsSold
	}
	if req.IsManufactured != nil {
		p.IsManufactured = *req.IsManufactured
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if req.Notes != "" {
		p.Notes = req.Notes
	}
	return p, s.repo.UpdateProduct(ctx, p)
}

func (s *Service) DeleteProduct(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.GetProduct(ctx, tenantID, id); err != nil {
		return fmt.Errorf("product not found")
	}
	return s.repo.DeleteProduct(ctx, tenantID, id)
}

// ── Variants ──────────────────────────────────────────────────────────────────

func (s *Service) ListVariants(ctx context.Context, tenantID, productID uint) ([]ProductVariant, error) {
	return s.repo.ListVariants(ctx, tenantID, productID)
}

func (s *Service) CreateVariant(ctx context.Context, tenantID, productID uint, req *CreateVariantRequest) (*ProductVariant, error) {
	v := &ProductVariant{
		TenantID:   tenantID,
		ProductID:  productID,
		Code:       req.Code,
		Name:       req.Name,
		CostPrice:  req.CostPrice,
		SalesPrice: req.SalesPrice,
		IsActive:   true,
	}
	if err := s.repo.CreateVariant(ctx, v); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("variant code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return v, nil
}

// ── Prices ────────────────────────────────────────────────────────────────────

func (s *Service) ListPrices(ctx context.Context, tenantID, productID uint) ([]ProductPrice, error) {
	return s.repo.ListPrices(ctx, tenantID, productID)
}

func (s *Service) CreatePrice(ctx context.Context, tenantID, productID uint, req *CreatePriceRequest) (*ProductPrice, error) {
	pType := req.PriceType
	if pType == "" {
		pType = "SALES"
	}
	if !validPriceTypes[pType] {
		return nil, fmt.Errorf("invalid price_type: %s", pType)
	}
	p := &ProductPrice{
		TenantID:      tenantID,
		ProductID:     productID,
		PartyID:       req.PartyID,
		PriceType:     pType,
		CurrencyID:    req.CurrencyID,
		Price:         req.Price,
		MinQty:        req.MinQty,
		EffectiveFrom: req.EffectiveFrom,
		EffectiveTo:   req.EffectiveTo,
		IsActive:      true,
	}
	return p, s.repo.CreatePrice(ctx, p)
}

func (s *Service) DeletePrice(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeletePrice(ctx, tenantID, id)
}
