package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/GT-610/tairitsu/internal/app/models"
)

// JWTClaims defines the structure of JWT claims used for authentication
type JWTClaims struct {
	UserID   string `json:"user_id"`   // Unique identifier for the user
	Username string `json:"username"` // Username of the authenticated user
	Role     string `json:"role"`     // User role for authorization purposes
	jwt.RegisteredClaims               // Standard JWT registered claims
}

// JWTService handles JWT token generation, validation, and parsing
type JWTService struct {
	secretKey     string        // Secret key used for signing tokens
	accessExpiry  time.Duration // Default expiration time for access tokens
}

// NewJWTService creates a new JWT service instance with the provided secret key
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		accessExpiry:  time.Hour * 24, // 24 hours expiration time
	}
}

// GenerateToken creates a new JWT token for the given user
func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	// Create claims
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证声明
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}