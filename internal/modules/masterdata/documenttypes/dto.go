package documenttypes

// ── Types ────────────────────────────────────────────────────────────────────

type CreateTypeRequest struct {
	Model    string `json:"model"     validate:"required,oneof=PR PO PI GRN MR GT GI SA QC SQ SO DO SI"`
	Code     string `json:"code"      validate:"required,max=50"`
	Name     string `json:"name"      validate:"required,max=200"`
	IsActive *bool  `json:"is_active"`
}

type UpdateTypeRequest struct {
	Code     string `json:"code"      validate:"omitempty,max=50"`
	Name     string `json:"name"      validate:"omitempty,max=200"`
	IsActive *bool  `json:"is_active"`
}

// ── Fields ───────────────────────────────────────────────────────────────────

type CreateFieldRequest struct {
	Code         string `json:"code"          validate:"required,max=50"`
	Label        string `json:"label"         validate:"required,max=200"`
	Kind         string `json:"kind"          validate:"required,oneof=PR PO PI GRN MR GT GI SA QC SQ SO DO SI MO Customer Supplier Product Warehouse Department"`
	IsRequired   bool   `json:"is_required"`
	DisplayOrder int    `json:"display_order"`
}

type UpdateFieldRequest struct {
	Code         string `json:"code"          validate:"omitempty,max=50"`
	Label        string `json:"label"         validate:"omitempty,max=200"`
	Kind         string `json:"kind"          validate:"omitempty,oneof=PR PO PI GRN MR GT GI SA QC SQ SO DO SI MO Customer Supplier Product Warehouse Department"`
	IsRequired   *bool  `json:"is_required"`
	DisplayOrder *int   `json:"display_order"`
}

// ── Batch field-value upsert ─────────────────────────────────────────────────
//
// One POST body describing every field on a document.

type ValueEntry struct {
	FieldID uint  `json:"field_id" validate:"required"`
	RefID   *uint `json:"ref_id"` // null clears the slot
}

type UpsertValuesRequest struct {
	Values []ValueEntry `json:"values"`
}

type ValueResponse struct {
	FieldID uint   `json:"field_id"`
	Code    string `json:"code"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	RefID   *uint  `json:"ref_id"`
}
