package services

import (
	"errors"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims defines the structure of JWT claims used for authentication
type JWTClaims struct {
	UserID               string `json:"user_id"`    // Unique identifier for the user
	Username             string `json:"username"`   // Username of the authenticated user
	Role                 string `json:"role"`       // User role for authorization purposes
	SessionID            string `json:"session_id"` // Unique identifier for the login session
	jwt.RegisteredClaims        // Standard JWT registered claims
}

// JWTService handles JWT token generation, validation, and parsing
type JWTService struct {
	secretKey    []byte        // Secret key used for signing tokens
	accessExpiry time.Duration // Default expiration time for access tokens
}

// NewJWTService creates a new JWT service instance with the provided secret key
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:    []byte(secretKey),
		accessExpiry: time.Hour * 24, // 24 hours expiration time
	}
}

// GenerateToken creates a new JWT token for the given user
func (s *JWTService) AccessExpiry() time.Duration {
	return s.accessExpiry
}

// GenerateToken creates a new JWT token for the given user
func (s *JWTService) GenerateToken(user *models.User, sessionID string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Validate claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
