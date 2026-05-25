package products

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// Categories
	ListCategories(ctx context.Context, tenantID uint) ([]ProductCategory, error)
	GetCategory(ctx context.Context, tenantID, id uint) (*ProductCategory, error)
	CreateCategory(ctx context.Context, c *ProductCategory) error
	UpdateCategory(ctx context.Context, c *ProductCategory) error
	DeleteCategory(ctx context.Context, tenantID, id uint) error

	// UOMs
	ListUOMs(ctx context.Context, tenantID uint) ([]UnitOfMeasure, error)
	GetUOM(ctx context.Context, tenantID, id uint) (*UnitOfMeasure, error)
	CreateUOM(ctx context.Context, u *UnitOfMeasure) error
	UpdateUOM(ctx context.Context, u *UnitOfMeasure) error

	// Products
	ListProducts(ctx context.Context, tenantID uint, productType string, activeOnly bool) ([]Product, error)
	GetProduct(ctx context.Context, tenantID, id uint) (*Product, error)
	CreateProduct(ctx context.Context, p *Product) error
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, tenantID, id uint) error

	// Variants
	ListVariants(ctx context.Context, tenantID, productID uint) ([]ProductVariant, error)
	CreateVariant(ctx context.Context, v *ProductVariant) error
	UpdateVariant(ctx context.Context, v *ProductVariant) error

	// Prices
	ListPrices(ctx context.Context, tenantID, productID uint) ([]ProductPrice, error)
	CreatePrice(ctx context.Context, p *ProductPrice) error
	UpdatePrice(ctx context.Context, p *ProductPrice) error
	DeletePrice(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

func (r *dbRepository) ListCategories(ctx context.Context, tenantID uint) ([]ProductCategory, error) {
	var rows []ProductCategory
	return rows, r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name").Find(&rows).Error
}

func (r *dbRepository) GetCategory(ctx context.Context, tenantID, id uint) (*ProductCategory, error) {
	var c ProductCategory
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&c).Error
	return &c, err
}

func (r *dbRepository) CreateCategory(ctx context.Context, c *ProductCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *dbRepository) UpdateCategory(ctx context.Context, c *ProductCategory) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *dbRepository) DeleteCategory(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ProductCategory{}).Error
}

func (r *dbRepository) ListUOMs(ctx context.Context, tenantID uint) ([]UnitOfMeasure, error) {
	var rows []UnitOfMeasure
	return rows, r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name").Find(&rows).Error
}

func (r *dbRepository) GetUOM(ctx context.Context, tenantID, id uint) (*UnitOfMeasure, error) {
	var u UnitOfMeasure
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&u).Error
	return &u, err
}

func (r *dbRepository) CreateUOM(ctx context.Context, u *UnitOfMeasure) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *dbRepository) UpdateUOM(ctx context.Context, u *UnitOfMeasure) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *dbRepository) ListProducts(ctx context.Context, tenantID uint, productType string, activeOnly bool) ([]Product, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if productType != "" {
		q = q.Where("product_type = ?", productType)
	}
	if activeOnly {
		q = q.Where("is_active = true")
	}
	var rows []Product
	return rows, q.Order("name").Find(&rows).Error
}

func (r *dbRepository) GetProduct(ctx context.Context, tenantID, id uint) (*Product, error) {
	var p Product
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&p).Error
	return &p, err
}

func (r *dbRepository) CreateProduct(ctx context.Context, p *Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *dbRepository) UpdateProduct(ctx context.Context, p *Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *dbRepository) DeleteProduct(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Product{}).Error
}

func (r *dbRepository) ListVariants(ctx context.Context, tenantID, productID uint) ([]ProductVariant, error) {
	var rows []ProductVariant
	return rows, r.db.WithContext(ctx).Where("tenant_id = ? AND product_id = ?", tenantID, productID).Find(&rows).Error
}

func (r *dbRepository) CreateVariant(ctx context.Context, v *ProductVariant) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *dbRepository) UpdateVariant(ctx context.Context, v *ProductVariant) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *dbRepository) ListPrices(ctx context.Context, tenantID, productID uint) ([]ProductPrice, error) {
	var rows []ProductPrice
	return rows, r.db.WithContext(ctx).Where("tenant_id = ? AND product_id = ?", tenantID, productID).Order("effective_from DESC").Find(&rows).Error
}

func (r *dbRepository) CreatePrice(ctx context.Context, p *ProductPrice) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *dbRepository) UpdatePrice(ctx context.Context, p *ProductPrice) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *dbRepository) DeletePrice(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ProductPrice{}).Error
}
