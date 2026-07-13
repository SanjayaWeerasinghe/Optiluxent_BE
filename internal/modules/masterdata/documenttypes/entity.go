package documenttypes

import "time"

// DocumentType is a per-model classification (PR/PO/GRN/MR/GI/GT/SA/SQ/SO/DO/SI/PI/QC).
// system_key marks rows we seeded from the old string enums so the BE can
// keep its behavioural switches without re-inventing them.
type DocumentType struct {
	ID         uint      `json:"id"          gorm:"primaryKey"`
	TenantID   uint      `json:"tenant_id"   gorm:"not null;index"`
	Model      string    `json:"model"       gorm:"not null;size:20;index"`
	Code       string    `json:"code"        gorm:"not null;size:50"`
	Name       string    `json:"name"        gorm:"not null;size:200"`
	SystemKey  *string   `json:"system_key"  gorm:"size:50"`
	IsActive   bool      `json:"is_active"   gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Fields []DocumentTypeField `json:"fields,omitempty" gorm:"foreignKey:DocumentTypeID"`
}

func (DocumentType) TableName() string { return "document_types" }

// DocumentTypeField is a picker slot attached to a Type. On a document form
// each field renders as a dropdown of records of `kind`.
type DocumentTypeField struct {
	ID              uint      `json:"id"                gorm:"primaryKey"`
	TenantID        uint      `json:"tenant_id"         gorm:"not null;index"`
	DocumentTypeID  uint      `json:"document_type_id"  gorm:"not null;index"`
	Code            string    `json:"code"              gorm:"not null;size:50"`
	Label           string    `json:"label"             gorm:"not null;size:200"`
	Kind            string    `json:"kind"              gorm:"not null;size:20"`
	IsRequired      bool      `json:"is_required"       gorm:"not null;default:false"`
	DisplayOrder    int       `json:"display_order"     gorm:"not null;default:0"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (DocumentTypeField) TableName() string { return "document_type_fields" }

// DocumentFieldValue is the per-document picker value.
// (doc_kind, doc_id, field_id) identifies a slot on a doc; ref_id is the picked target.
type DocumentFieldValue struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	TenantID  uint      `json:"tenant_id"  gorm:"not null;index"`
	DocKind   string    `json:"doc_kind"   gorm:"not null;size:20;index"`
	DocID     uint      `json:"doc_id"     gorm:"not null;index"`
	FieldID   uint      `json:"field_id"   gorm:"not null;index"`
	RefID     uint      `json:"ref_id"     gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DocumentFieldValue) TableName() string { return "document_field_values" }
