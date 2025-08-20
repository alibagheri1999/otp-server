package entities

import (
	"time"
)

// UserRole represents the possible user roles
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// User represents a user in the system
type User struct {
	ID          int       `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Name        string    `json:"name" db:"name"`
	Role        UserRole  `json:"role" db:"role"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser creates a new user instance
func NewUser(phoneNumber, name string) *User {
	now := time.Now()
	return &User{
		PhoneNumber: phoneNumber,
		Name:        name,
		Role:        UserRoleUser,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewAdminUser creates a new admin user instance
func NewAdminUser(phoneNumber, name string) *User {
	now := time.Now()
	return &User{
		PhoneNumber: phoneNumber,
		Name:        name,
		Role:        UserRoleAdmin,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateLastSeen updates the last seen timestamp
func (u *User) UpdateLastSeen() {
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(name, phoneNumber string) {
	u.Name = name
	u.PhoneNumber = phoneNumber
	u.UpdatedAt = time.Now()
}

// UpdateRole updates the user's role
func (u *User) UpdateRole(role UserRole) {
	u.Role = role
	u.UpdatedAt = time.Now()
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsUser checks if the user is a regular user
func (u *User) IsUser() bool {
	return u.Role == UserRoleUser
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Activate activates the user
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}
