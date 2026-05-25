package role

import "time"

type CreateRoleRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" validate:"required,min=1"`
}

type PermissionResponse struct {
	ID          uint   `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          uint                 `json:"id"`
	TenantID    *uint                `json:"tenant_id,omitempty"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	CreatedAt   time.Time            `json:"created_at"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
}
