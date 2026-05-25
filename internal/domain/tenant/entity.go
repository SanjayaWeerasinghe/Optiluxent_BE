package tenant

import (
	"encoding/json"
	"time"
)

type Tenant struct {
	ID        uint            `json:"id" gorm:"primarykey"`
	Name      string          `json:"name" gorm:"size:255;not null"`
	Slug      string          `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Plan      string          `json:"plan" gorm:"size:50;not null;default:'standard'"`
	Status    string          `json:"status" gorm:"size:20;not null;default:'active'"`
	Config    json.RawMessage `json:"config,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
}

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusSuspended = "suspended"

	PlanStandard    = "standard"
	PlanProfessional = "professional"
	PlanEnterprise  = "enterprise"
)

func (Tenant) TableName() string {
	return "tenants"
}

func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive && t.DeletedAt == nil
}
