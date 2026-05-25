package user

type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  string `json:"last_name"  validate:"omitempty,min=1,max=100"`
	Status    string `json:"status"     validate:"omitempty,oneof=active inactive suspended"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type AssignRoleRequest struct {
	RoleID uint `json:"role_id" validate:"required"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	TenantID    *uint  `json:"tenant_id,omitempty"`
}

type ListUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}
