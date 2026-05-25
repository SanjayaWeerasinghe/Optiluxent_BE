package audit

import (
	"encoding/json"
	"time"
)

type Log struct {
	ID         uint            `json:"id" gorm:"primarykey"`
	TenantID   *uint           `json:"tenant_id,omitempty" gorm:"index"`
	UserID     *uint           `json:"user_id,omitempty" gorm:"index"`
	Action     string          `json:"action" gorm:"size:100;not null"`
	Resource   string          `json:"resource" gorm:"size:100;not null"`
	ResourceID string          `json:"resource_id,omitempty" gorm:"size:100"`
	OldValues  json.RawMessage `json:"old_values,omitempty" gorm:"type:jsonb"`
	NewValues  json.RawMessage `json:"new_values,omitempty" gorm:"type:jsonb"`
	IPAddress  string          `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent  string          `json:"user_agent,omitempty" gorm:"size:500"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (Log) TableName() string { return "audit_logs" }

// Entry is the input struct used by the audit logger to build a Log.
type Entry struct {
	TenantID   *uint
	UserID     *uint
	Action     string
	Resource   string
	ResourceID string
	OldValues  interface{}
	NewValues  interface{}
	IPAddress  string
	UserAgent  string
}
