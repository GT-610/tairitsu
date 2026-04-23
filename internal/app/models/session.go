package models

import "time"

// Session 登录会话模型
type Session struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	UserID     string     `json:"user_id" gorm:"index;not null"`
	UserAgent  string     `json:"user_agent"`
	IPAddress  string     `json:"ip_address"`
	RememberMe bool       `json:"remember_me"`
	LastSeenAt time.Time  `json:"last_seen_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// SessionResponse 会话响应
type SessionResponse struct {
	ID         string     `json:"id"`
	UserAgent  string     `json:"userAgent"`
	IPAddress  string     `json:"ipAddress"`
	RememberMe bool       `json:"rememberMe"`
	LastSeenAt time.Time  `json:"lastSeenAt"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	Current    bool       `json:"current"`
}

// ToResponse 转换为会话响应
func (s *Session) ToResponse(current bool) SessionResponse {
	return SessionResponse{
		ID:         s.ID,
		UserAgent:  s.UserAgent,
		IPAddress:  s.IPAddress,
		RememberMe: s.RememberMe,
		LastSeenAt: s.LastSeenAt,
		ExpiresAt:  s.ExpiresAt,
		RevokedAt:  s.RevokedAt,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
		Current:    current,
	}
}
