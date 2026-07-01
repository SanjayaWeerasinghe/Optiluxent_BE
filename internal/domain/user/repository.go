package user

import (
	"context"
)

// Repository defines the interface for user data operations
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uint) error

	// List retrieves a list of users with pagination
	List(ctx context.Context, limit, offset int) ([]*User, int64, error)

	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, id uint) error

	// ChangePassword updates the user's password
	ChangePassword(ctx context.Context, id uint, newPasswordHash string) error

	// SetRole replaces all role assignments for the user with this single role,
	// atomically updating the user_roles join table and the denormalised
	// users.role string. Caller is responsible for syncing Casbin groupings.
	SetRole(ctx context.Context, userID, roleID uint, roleName string) error
}
