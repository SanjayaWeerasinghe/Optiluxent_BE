package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	TenantID     *uint      `json:"tenant_id,omitempty" gorm:"index"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"`
	FirstName    string     `json:"first_name" gorm:"size:100"`
	LastName     string     `json:"last_name" gorm:"size:100"`
	Role         string     `json:"role" gorm:"size:50;not null;default:'user'"`
	Status       string     `json:"status" gorm:"size:20;not null;default:'active'"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// UserStatus constants
const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
)

// UserRole constants
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleUser    = "user"
	RoleGuest   = "guest"
)

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}

// UpdateLastLogin updates the user's last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// Soft delete the user
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.Status = StatusInactive
}

// IsDeleted checks if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
