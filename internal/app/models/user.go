package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"-" binding:"required"` // 密码不返回给客户端
	Email     string    `json:"email" binding:"required,email"`
	Role      string    `json:"role"` // admin, user
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword      string `json:"oldPassword" binding:"required"`
	NewPassword      string `json:"newPassword" binding:"required,min=6"`
	CurrentPassword  string `json:"current_password" binding:"required"`
	NewPasswordField string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword  string `json:"confirm_password" binding:"required"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse 转换为用户响应
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}