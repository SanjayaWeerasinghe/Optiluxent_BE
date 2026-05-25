package role

import "time"

type Role struct {
	ID          uint         `json:"id" gorm:"primarykey"`
	TenantID    *uint        `json:"tenant_id,omitempty" gorm:"index"`
	Name        string       `json:"name" gorm:"size:100;not null"`
	Description string       `json:"description" gorm:"size:255"`
	IsSystem    bool         `json:"is_system" gorm:"not null;default:false"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

// Permission is a lightweight struct for the many2many join — full type in permission package.
// Declared here to avoid import cycle; GORM resolves the table via TableName().
type Permission struct {
	ID          uint   `json:"id" gorm:"primarykey"`
	Resource    string `json:"resource" gorm:"size:100;not null"`
	Action      string `json:"action" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:255"`
}

func (Permission) TableName() string { return "permissions" }

func (Role) TableName() string { return "roles" }

func (r *Role) IsDeleted() bool { return r.DeletedAt != nil }
