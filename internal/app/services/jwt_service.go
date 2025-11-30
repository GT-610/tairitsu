package services

import (
	"errors"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims defines the structure of JWT claims used for authentication
type JWTClaims struct {
	UserID               string `json:"user_id"`  // Unique identifier for the user
	Username             string `json:"username"` // Username of the authenticated user
	Role                 string `json:"role"`     // User role for authorization purposes
	jwt.RegisteredClaims        // Standard JWT registered claims
}

// JWTService handles JWT token generation, validation, and parsing
type JWTService struct {
	secretKey    string        // Secret key used for signing tokens
	accessExpiry time.Duration // Default expiration time for access tokens
}

// NewJWTService creates a new JWT service instance with the provided secret key
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:    secretKey,
		accessExpiry: time.Hour * 24, // 24 hours expiration time
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

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken Validate JWT token
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.secretKey), nil
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
