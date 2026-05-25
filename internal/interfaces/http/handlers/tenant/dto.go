package tenant

import (
	"encoding/json"
	"time"
)

type CreateTenantRequest struct {
	Name   string          `json:"name"   validate:"required,min=2,max=255"`
	Slug   string          `json:"slug"   validate:"required,min=2,max=100"`
	Plan   string          `json:"plan"   validate:"omitempty,oneof=standard professional enterprise"`
	Config json.RawMessage `json:"config" validate:"omitempty"`
}

type UpdateTenantRequest struct {
	Name   string          `json:"name"   validate:"omitempty,min=2,max=255"`
	Plan   string          `json:"plan"   validate:"omitempty,oneof=standard professional enterprise"`
	Status string          `json:"status" validate:"omitempty,oneof=active inactive suspended"`
	Config json.RawMessage `json:"config" validate:"omitempty"`
}

type TenantResponse struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Plan      string          `json:"plan"`
	Status    string          `json:"status"`
	Config    json.RawMessage `json:"config,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type ListTenantsResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}
