package featureflags

import (
	"encoding/json"
	"time"
)

type FeatureFlag struct {
	ID          uint            `json:"id" gorm:"primarykey"`
	TenantID    *uint           `json:"tenant_id,omitempty" gorm:"index"`
	Name        string          `json:"name" gorm:"size:100;not null"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled" gorm:"not null;default:false"`
	Rules       json.RawMessage `json:"rules,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (FeatureFlag) TableName() string { return "feature_flags" }

// FlagRules holds optional targeting rules for a flag.
type FlagRules struct {
	UserIDs    []uint  `json:"user_ids,omitempty"`    // explicit user allow-list
	Percentage float64 `json:"percentage,omitempty"` // 0-100 rollout percentage
}
