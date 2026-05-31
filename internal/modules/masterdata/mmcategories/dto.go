package mmcategories

type CreateCategoryRequest struct {
	ParentID *uint  `json:"parent_id"`
	Code     string `json:"code" validate:"required,max=50"`
	Name     string `json:"name" validate:"required,max=200"`
}

type UpdateCategoryRequest struct {
	ParentID *uint   `json:"parent_id"`
	Name     string  `json:"name"      validate:"omitempty,max=200"`
	IsActive *bool   `json:"is_active"`
}
