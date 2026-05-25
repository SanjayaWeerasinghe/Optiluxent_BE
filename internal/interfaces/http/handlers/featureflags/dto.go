package featureflags

import (
	"encoding/json"
	"time"
)

type CreateFlagRequest struct {
	Name        string          `json:"name" validate:"required,min=2,max=100"`
	Description string          `json:"description" validate:"max=500"`
	Enabled     bool            `json:"enabled"`
	Rules       json.RawMessage `json:"rules,omitempty"`
}

type UpdateFlagRequest struct {
	Description *string         `json:"description,omitempty" validate:"omitempty,max=500"`
	Enabled     *bool           `json:"enabled,omitempty"`
	Rules       json.RawMessage `json:"rules,omitempty"`
}

type SetEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type FlagResponse struct {
	ID          uint            `json:"id"`
	TenantID    *uint           `json:"tenant_id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	Rules       json.RawMessage `json:"rules,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
