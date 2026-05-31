package materials

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Count(ctx context.Context, tenantID uint, materialType string) (int64, error)
	List(ctx context.Context, tenantID uint, materialType string, limit, offset int) ([]Material, error)
	Get(ctx context.Context, tenantID, id uint) (*MaterialDetail, error)
	Create(ctx context.Context, m *Material) error
	Update(ctx context.Context, m *Material) error
	Delete(ctx context.Context, tenantID, id uint) error

	GetPurchasing(ctx context.Context, materialID uint) (*MaterialPurchasing, error)
	UpsertPurchasing(ctx context.Context, p *MaterialPurchasing) error

	GetManufacturing(ctx context.Context, materialID uint) (*MaterialManufacturing, error)
	UpsertManufacturing(ctx context.Context, m *MaterialManufacturing) error

	GetWarehouse(ctx context.Context, materialID uint) (*MaterialWarehouse, error)
	UpsertWarehouse(ctx context.Context, w *MaterialWarehouse) error

	ListVendors(ctx context.Context, materialID uint) ([]MaterialVendor, error)
	CreateVendor(ctx context.Context, v *MaterialVendor) error
	UpdateVendor(ctx context.Context, v *MaterialVendor) error
	DeleteVendor(ctx context.Context, materialID, vendorRowID uint) error
	ReplaceVendors(ctx context.Context, materialID, tenantID uint, vendors []MaterialVendor) error

	ListMeasurements(ctx context.Context, materialID uint) ([]MaterialMeasurement, error)
	CreateMeasurement(ctx context.Context, m *MaterialMeasurement) error
	UpdateMeasurement(ctx context.Context, m *MaterialMeasurement) error
	DeleteMeasurement(ctx context.Context, materialID, measurementID uint) error
	ReplaceMeasurements(ctx context.Context, materialID, tenantID uint, items []MaterialMeasurement) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

// ── Core Material ─────────────────────────────────────────────────────────────

func (r *dbRepository) Count(ctx context.Context, tenantID uint, materialType string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&Material{}).Where("tenant_id = ?", tenantID)
	if materialType != "" {
		q = q.Where("material_type = ?", materialType)
	}
	var count int64
	return count, q.Count(&count).Error
}

func (r *dbRepository) List(ctx context.Context, tenantID uint, materialType string, limit, offset int) ([]Material, error) {
	var rows []Material
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if materialType != "" {
		q = q.Where("material_type = ?", materialType)
	}
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	return rows, q.Order("code").Find(&rows).Error
}

func (r *dbRepository) Get(ctx context.Context, tenantID, id uint) (*MaterialDetail, error) {
	var m Material
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error; err != nil {
		return nil, err
	}
	detail := &MaterialDetail{Material: m}
	if m.CategoryID != nil {
		r.db.WithContext(ctx).Raw(
			`SELECT code, name FROM mm_categories WHERE id = ?`, *m.CategoryID,
		).Row().Scan(&detail.CategoryCode, &detail.CategoryName)
	}
	return detail, nil
}

func (r *dbRepository) Create(ctx context.Context, m *Material) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dbRepository) Update(ctx context.Context, m *Material) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *dbRepository) Delete(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&Material{}).Error
}

// ── Purchasing ────────────────────────────────────────────────────────────────

const purchasingSelect = `mm_purchasing.*,
	COALESCE(u.code,'') AS purchasing_uom_code,
	COALESCE(u.name,'') AS purchasing_uom_name`

func (r *dbRepository) GetPurchasing(ctx context.Context, materialID uint) (*MaterialPurchasing, error) {
	var p MaterialPurchasing
	err := r.db.WithContext(ctx).
		Table("mm_purchasing").
		Select(purchasingSelect).
		Joins("LEFT JOIN units_of_measure u ON u.id = mm_purchasing.purchasing_uom_id").
		Where("mm_purchasing.material_id = ?", materialID).
		First(&p).Error
	if err != nil {
		return &MaterialPurchasing{MaterialID: materialID}, nil
	}
	return &p, nil
}

func (r *dbRepository) UpsertPurchasing(ctx context.Context, p *MaterialPurchasing) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "material_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"purchasing_uom_id", "under_delivery_pct", "over_delivery_pct", "updated_at"}),
		}).Create(p).Error
}

// ── Manufacturing ─────────────────────────────────────────────────────────────

const manufacturingSelect = `mm_manufacturing.*,
	COALESCE(u.code,'') AS production_uom_code,
	COALESCE(u.name,'') AS production_uom_name`

func (r *dbRepository) GetManufacturing(ctx context.Context, materialID uint) (*MaterialManufacturing, error) {
	var m MaterialManufacturing
	err := r.db.WithContext(ctx).
		Table("mm_manufacturing").
		Select(manufacturingSelect).
		Joins("LEFT JOIN units_of_measure u ON u.id = mm_manufacturing.production_uom_id").
		Where("mm_manufacturing.material_id = ?", materialID).
		First(&m).Error
	if err != nil {
		return &MaterialManufacturing{MaterialID: materialID}, nil
	}
	return &m, nil
}

func (r *dbRepository) UpsertManufacturing(ctx context.Context, m *MaterialManufacturing) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "material_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"production_uom_id", "reorder_qty_level", "safety_level",
				"production_days", "delivery_days", "grn_days", "procurement_repeat_days", "updated_at",
			}),
		}).Create(m).Error
}

// ── Warehouse ─────────────────────────────────────────────────────────────────

const warehouseSelect = `mm_warehouse.*,
	COALESCE(u.code,'') AS stocking_uom_code,
	COALESCE(u.name,'') AS stocking_uom_name`

func (r *dbRepository) GetWarehouse(ctx context.Context, materialID uint) (*MaterialWarehouse, error) {
	var w MaterialWarehouse
	err := r.db.WithContext(ctx).
		Table("mm_warehouse").
		Select(warehouseSelect).
		Joins("LEFT JOIN units_of_measure u ON u.id = mm_warehouse.stocking_uom_id").
		Where("mm_warehouse.material_id = ?", materialID).
		First(&w).Error
	if err != nil {
		return &MaterialWarehouse{MaterialID: materialID}, nil
	}
	return &w, nil
}

func (r *dbRepository) UpsertWarehouse(ctx context.Context, w *MaterialWarehouse) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "material_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"stocking_uom_id", "stock_removal", "storage_main", "storage_damaged", "storage_hold",
				"batch_process", "production_date_check", "expiry_date_check", "qc_check",
				"grn_with_po_uom_image", "updated_at",
			}),
		}).Create(w).Error
}

// ── Vendors ───────────────────────────────────────────────────────────────────

const vendorSelect = `mm_vendors.*,
	COALESCE(p.name,'') AS vendor_name,
	COALESCE(c.code,'') AS currency_code`

func (r *dbRepository) ListVendors(ctx context.Context, materialID uint) ([]MaterialVendor, error) {
	var rows []MaterialVendor
	err := r.db.WithContext(ctx).
		Table("mm_vendors").
		Select(vendorSelect).
		Joins("LEFT JOIN parties p ON p.id = mm_vendors.vendor_id").
		Joins("LEFT JOIN currencies c ON c.id = mm_vendors.currency_id").
		Where("mm_vendors.material_id = ?", materialID).
		Order("mm_vendors.id").
		Find(&rows).Error
	return rows, err
}

func (r *dbRepository) CreateVendor(ctx context.Context, v *MaterialVendor) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *dbRepository) UpdateVendor(ctx context.Context, v *MaterialVendor) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *dbRepository) DeleteVendor(ctx context.Context, materialID, vendorRowID uint) error {
	return r.db.WithContext(ctx).
		Where("material_id = ? AND id = ?", materialID, vendorRowID).
		Delete(&MaterialVendor{}).Error
}

func (r *dbRepository) ReplaceVendors(ctx context.Context, materialID, tenantID uint, vendors []MaterialVendor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("material_id = ?", materialID).Delete(&MaterialVendor{}).Error; err != nil {
			return err
		}
		if len(vendors) == 0 {
			return nil
		}
		for i := range vendors {
			vendors[i].MaterialID = materialID
			vendors[i].TenantID = tenantID
		}
		return tx.Create(&vendors).Error
	})
}

// ── Measurements ──────────────────────────────────────────────────────────────

const measSelect = `mm_measurements.*,
	bu.code AS base_uom_code, bu.name AS base_uom_name,
	tu.code AS target_uom_code, tu.name AS target_uom_name`

func (r *dbRepository) ListMeasurements(ctx context.Context, materialID uint) ([]MaterialMeasurement, error) {
	var rows []MaterialMeasurement
	err := r.db.WithContext(ctx).
		Table("mm_measurements").
		Select(measSelect).
		Joins("JOIN units_of_measure bu ON bu.id = mm_measurements.base_uom_id").
		Joins("JOIN units_of_measure tu ON tu.id = mm_measurements.target_uom_id").
		Where("mm_measurements.material_id = ?", materialID).
		Order("mm_measurements.id").
		Find(&rows).Error
	return rows, err
}

func (r *dbRepository) CreateMeasurement(ctx context.Context, m *MaterialMeasurement) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dbRepository) UpdateMeasurement(ctx context.Context, m *MaterialMeasurement) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *dbRepository) DeleteMeasurement(ctx context.Context, materialID, measurementID uint) error {
	return r.db.WithContext(ctx).
		Where("material_id = ? AND id = ?", materialID, measurementID).
		Delete(&MaterialMeasurement{}).Error
}

func (r *dbRepository) ReplaceMeasurements(ctx context.Context, materialID, tenantID uint, items []MaterialMeasurement) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("material_id = ?", materialID).Delete(&MaterialMeasurement{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].MaterialID = materialID
			items[i].TenantID = tenantID
		}
		return tx.Create(&items).Error
	})
}
