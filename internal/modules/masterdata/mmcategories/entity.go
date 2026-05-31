package mmcategories

import "time"

type MaterialCategory struct {
	ID       uint    `json:"id"        gorm:"primaryKey"`
	TenantID uint    `json:"tenant_id" gorm:"not null;index"`
	ParentID *uint   `json:"parent_id"`
	Code     string  `json:"code"      gorm:"not null;size:50"`
	Name     string  `json:"name"      gorm:"not null;size:200"`
	IsActive bool    `json:"is_active" gorm:"not null;default:true"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MaterialCategory) TableName() string { return "mm_categories" }

// CategoryResponse is returned by List/Get — adds computed fields
type CategoryResponse struct {
	MaterialCategory
	Depth      int    `json:"depth"`
	Path       string `json:"path"`
	ParentCode string `json:"parent_code"`
	ParentName string `json:"parent_name"`
}
