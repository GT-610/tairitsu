package models

import (
	"time"
)

// User represents a user account.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Password is never returned to the client
	Role      string    `json:"role"` // admin, user
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest represents a login request payload.
type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

// RegisterRequest represents a registration request payload.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest represents a password change request payload.
type ChangePasswordRequest struct {
	CurrentPassword     string `json:"current_password"`
	NewPassword         string `json:"new_password"`
	ConfirmPassword     string `json:"confirm_password"`
	LogoutOtherSessions bool   `json:"logout_other_sessions"`
}

// UserResponse is the API response shape for a user.
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToResponse converts a User to a UserResponse.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
