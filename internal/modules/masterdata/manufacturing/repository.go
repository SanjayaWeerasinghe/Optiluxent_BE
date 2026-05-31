package manufacturing

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// BOMs
	CountBOMs(ctx context.Context, tenantID uint) (int64, error)
	ListBOMs(ctx context.Context, tenantID uint, limit, offset int) ([]BillOfMaterials, error)
	GetBOM(ctx context.Context, tenantID, id uint) (*BillOfMaterials, error)
	CreateBOM(ctx context.Context, b *BillOfMaterials) error
	UpdateBOM(ctx context.Context, b *BillOfMaterials) error
	DeleteBOM(ctx context.Context, tenantID, id uint) error
	AddBOMLine(ctx context.Context, line *BOMLine) error
	DeleteBOMLine(ctx context.Context, tenantID, id uint) error

	// Work Centers
	CountWorkCenters(ctx context.Context, tenantID uint) (int64, error)
	ListWorkCenters(ctx context.Context, tenantID uint, limit, offset int) ([]WorkCenter, error)
	GetWorkCenter(ctx context.Context, tenantID, id uint) (*WorkCenter, error)
	CreateWorkCenter(ctx context.Context, wc *WorkCenter) error
	UpdateWorkCenter(ctx context.Context, wc *WorkCenter) error
	DeleteWorkCenter(ctx context.Context, tenantID, id uint) error

	// Routings
	CountRoutings(ctx context.Context, tenantID uint) (int64, error)
	ListRoutings(ctx context.Context, tenantID uint, limit, offset int) ([]Routing, error)
	GetRouting(ctx context.Context, tenantID, id uint) (*Routing, error)
	CreateRouting(ctx context.Context, r *Routing) error
	UpdateRouting(ctx context.Context, r *Routing) error
	DeleteRouting(ctx context.Context, tenantID, id uint) error
	AddRoutingOperation(ctx context.Context, op *RoutingOperation) error
	DeleteRoutingOperation(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

func (r *dbRepository) CountBOMs(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&BillOfMaterials{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *dbRepository) ListBOMs(ctx context.Context, tenantID uint, limit, offset int) ([]BillOfMaterials, error) {
	var rows []BillOfMaterials
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	return rows, q.Order("code").Find(&rows).Error
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

func (r *dbRepository) CountWorkCenters(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&WorkCenter{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *dbRepository) ListWorkCenters(ctx context.Context, tenantID uint, limit, offset int) ([]WorkCenter, error) {
	var rows []WorkCenter
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	return rows, q.Order("name").Find(&rows).Error
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

func (r *dbRepository) CountRoutings(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&Routing{}).Where("tenant_id = ?", tenantID).Count(&count).Error
}

func (r *dbRepository) ListRoutings(ctx context.Context, tenantID uint, limit, offset int) ([]Routing, error) {
	var rows []Routing
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	return rows, q.Order("code").Find(&rows).Error
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
