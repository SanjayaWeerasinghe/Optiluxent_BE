package permission

import "time"

type Permission struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Resource    string    `json:"resource" gorm:"size:100;not null"`
	Action      string    `json:"action" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Permission) TableName() string { return "permissions" }

// Key returns the permission in "resource:action" format used by Casbin.
func (p *Permission) Key() string {
	return p.Resource + ":" + p.Action
}
