package manufacturing

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// BOMs
	ListBOMs(ctx context.Context, tenantID uint) ([]BillOfMaterials, error)
	GetBOM(ctx context.Context, tenantID, id uint) (*BillOfMaterials, error)
	CreateBOM(ctx context.Context, b *BillOfMaterials) error
	UpdateBOM(ctx context.Context, b *BillOfMaterials) error
	DeleteBOM(ctx context.Context, tenantID, id uint) error
	AddBOMLine(ctx context.Context, line *BOMLine) error
	DeleteBOMLine(ctx context.Context, tenantID, id uint) error

	// Work Centers
	ListWorkCenters(ctx context.Context, tenantID uint) ([]WorkCenter, error)
	GetWorkCenter(ctx context.Context, tenantID, id uint) (*WorkCenter, error)
	CreateWorkCenter(ctx context.Context, wc *WorkCenter) error
	UpdateWorkCenter(ctx context.Context, wc *WorkCenter) error
	DeleteWorkCenter(ctx context.Context, tenantID, id uint) error

	// Routings
	ListRoutings(ctx context.Context, tenantID uint) ([]Routing, error)
	GetRouting(ctx context.Context, tenantID, id uint) (*Routing, error)
	CreateRouting(ctx context.Context, r *Routing) error
	UpdateRouting(ctx context.Context, r *Routing) error
	DeleteRouting(ctx context.Context, tenantID, id uint) error
	AddRoutingOperation(ctx context.Context, op *RoutingOperation) error
	DeleteRoutingOperation(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

func (r *dbRepository) ListBOMs(ctx context.Context, tenantID uint) ([]BillOfMaterials, error) {
	var rows []BillOfMaterials
	return rows, r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("code").Find(&rows).Error
}

func (r *dbRepository) GetBOM(ctx context.Context, tenantID, id uint) (*BillOfMaterials, error) {
	var b BillOfMaterials
	err := r.db.WithContext(ctx).Preload("Lines").Where("tenant_id = ? AND id = ?", tenantID, id).First(&b).Error
	return &b, err
}

func (r *dbRepository) CreateBOM(ctx context.Context, b *BillOfMaterials) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *dbRepository) UpdateBOM(ctx context.Context, b *BillOfMaterials) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *dbRepository) DeleteBOM(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&BillOfMaterials{}).Error
}

func (r *dbRepository) AddBOMLine(ctx context.Context, line *BOMLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) DeleteBOMLine(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&BOMLine{}).Error
}

func (r *dbRepository) ListWorkCenters(ctx context.Context, tenantID uint) ([]WorkCenter, error) {
	var rows []WorkCenter
	return rows, r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name").Find(&rows).Error
}

func (r *dbRepository) GetWorkCenter(ctx context.Context, tenantID, id uint) (*WorkCenter, error) {
	var wc WorkCenter
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&wc).Error
	return &wc, err
}

func (r *dbRepository) CreateWorkCenter(ctx context.Context, wc *WorkCenter) error {
	return r.db.WithContext(ctx).Create(wc).Error
}

func (r *dbRepository) UpdateWorkCenter(ctx context.Context, wc *WorkCenter) error {
	return r.db.WithContext(ctx).Save(wc).Error
}

func (r *dbRepository) DeleteWorkCenter(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&WorkCenter{}).Error
}

func (r *dbRepository) ListRoutings(ctx context.Context, tenantID uint) ([]Routing, error) {
	var rows []Routing
	return rows, r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("code").Find(&rows).Error
}

func (r *dbRepository) GetRouting(ctx context.Context, tenantID, id uint) (*Routing, error) {
	var rt Routing
	err := r.db.WithContext(ctx).Preload("Operations").Where("tenant_id = ? AND id = ?", tenantID, id).First(&rt).Error
	return &rt, err
}

func (r *dbRepository) CreateRouting(ctx context.Context, rt *Routing) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *dbRepository) UpdateRouting(ctx context.Context, rt *Routing) error {
	return r.db.WithContext(ctx).Save(rt).Error
}

func (r *dbRepository) DeleteRouting(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Routing{}).Error
}

func (r *dbRepository) AddRoutingOperation(ctx context.Context, op *RoutingOperation) error {
	return r.db.WithContext(ctx).Create(op).Error
}

func (r *dbRepository) DeleteRoutingOperation(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&RoutingOperation{}).Error
}
