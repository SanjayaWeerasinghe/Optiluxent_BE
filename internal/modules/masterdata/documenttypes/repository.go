package documenttypes

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// Types
	ListTypes(ctx context.Context, tenantID uint, model string) ([]DocumentType, error)
	GetType(ctx context.Context, tenantID, id uint) (*DocumentType, error)
	GetTypeWithFields(ctx context.Context, tenantID, id uint) (*DocumentType, error)
	FindSystemType(ctx context.Context, tenantID uint, model, systemKey string) (uint, error)
	CreateType(ctx context.Context, t *DocumentType) error
	UpdateType(ctx context.Context, t *DocumentType) error
	DeleteType(ctx context.Context, tenantID, id uint) error

	// Fields
	ListFields(ctx context.Context, tenantID, typeID uint) ([]DocumentTypeField, error)
	GetField(ctx context.Context, tenantID, id uint) (*DocumentTypeField, error)
	CreateField(ctx context.Context, f *DocumentTypeField) error
	UpdateField(ctx context.Context, f *DocumentTypeField) error
	DeleteField(ctx context.Context, tenantID, id uint) error

	// Values
	ListValuesForDoc(ctx context.Context, tenantID uint, docKind string, docID uint) ([]DocumentFieldValue, error)
	UpsertValue(ctx context.Context, v *DocumentFieldValue) error
	DeleteValue(ctx context.Context, tenantID uint, docKind string, docID, fieldID uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

// ── Types ────────────────────────────────────────────────────────────────────

func (r *dbRepository) ListTypes(ctx context.Context, tenantID uint, model string) ([]DocumentType, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if model != "" {
		q = q.Where("model = ?", model)
	}
	var rows []DocumentType
	return rows, q.Order("model, code").Find(&rows).Error
}

func (r *dbRepository) GetType(ctx context.Context, tenantID, id uint) (*DocumentType, error) {
	var t DocumentType
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error
	return &t, err
}

func (r *dbRepository) GetTypeWithFields(ctx context.Context, tenantID, id uint) (*DocumentType, error) {
	var t DocumentType
	err := r.db.WithContext(ctx).
		Preload("Fields", func(db *gorm.DB) *gorm.DB { return db.Order("display_order ASC, id ASC") }).
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error
	return &t, err
}

func (r *dbRepository) FindSystemType(ctx context.Context, tenantID uint, model, systemKey string) (uint, error) {
	var id uint
	err := r.db.WithContext(ctx).Model(&DocumentType{}).
		Where("tenant_id = ? AND model = ? AND system_key = ?", tenantID, model, systemKey).
		Limit(1).Pluck("id", &id).Error
	return id, err
}

func (r *dbRepository) CreateType(ctx context.Context, t *DocumentType) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *dbRepository) UpdateType(ctx context.Context, t *DocumentType) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *dbRepository) DeleteType(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&DocumentType{}).Error
}

// ── Fields ───────────────────────────────────────────────────────────────────

func (r *dbRepository) ListFields(ctx context.Context, tenantID, typeID uint) ([]DocumentTypeField, error) {
	var rows []DocumentTypeField
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_type_id = ?", tenantID, typeID).
		Order("display_order ASC, id ASC").Find(&rows).Error
}

func (r *dbRepository) GetField(ctx context.Context, tenantID, id uint) (*DocumentTypeField, error) {
	var f DocumentTypeField
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&f).Error
	return &f, err
}

func (r *dbRepository) CreateField(ctx context.Context, f *DocumentTypeField) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *dbRepository) UpdateField(ctx context.Context, f *DocumentTypeField) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *dbRepository) DeleteField(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&DocumentTypeField{}).Error
}

// ── Values ───────────────────────────────────────────────────────────────────

func (r *dbRepository) ListValuesForDoc(ctx context.Context, tenantID uint, docKind string, docID uint) ([]DocumentFieldValue, error) {
	var rows []DocumentFieldValue
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND doc_kind = ? AND doc_id = ?", tenantID, docKind, docID).
		Find(&rows).Error
}

func (r *dbRepository) UpsertValue(ctx context.Context, v *DocumentFieldValue) error {
	// One row per (doc_kind, doc_id, field_id).
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO document_field_values (tenant_id, doc_kind, doc_id, field_id, ref_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		 ON CONFLICT (doc_kind, doc_id, field_id)
		 DO UPDATE SET ref_id = EXCLUDED.ref_id, updated_at = NOW()`,
		v.TenantID, v.DocKind, v.DocID, v.FieldID, v.RefID,
	).Error
}

func (r *dbRepository) DeleteValue(ctx context.Context, tenantID uint, docKind string, docID, fieldID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND doc_kind = ? AND doc_id = ? AND field_id = ?", tenantID, docKind, docID, fieldID).
		Delete(&DocumentFieldValue{}).Error
}
