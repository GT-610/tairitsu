package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTService(t *testing.T) {
	// Arrange
	secretKey := "test-secret-key"

	// Act
	jwtService := services.NewJWTService(secretKey)

	// Assert
	assert.NotNil(t, jwtService)
}

func TestJWTService_GenerateToken(t *testing.T) {
	// Arrange
	secretKey := "test-secret-key"
	jwtService := services.NewJWTService(secretKey)
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}

	// Act
	token, err := jwtService.GenerateToken(user)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTService_ValidateToken(t *testing.T) {
	// Arrange
	secretKey := "test-secret-key"
	jwtService := services.NewJWTService(secretKey)
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}

	// Generate a valid token
	token, err := jwtService.GenerateToken(user)
	assert.NoError(t, err)

	// Act
	claims, err := jwtService.ValidateToken(token)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Role, claims.Role)
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	// Arrange
	secretKey := "test-secret-key"
	jwtService := services.NewJWTService(secretKey)
	invalidToken := "invalid-token"

	// Act
	claims, err := jwtService.ValidateToken(invalidToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	// Arrange
	// Generate token with one secret
	jwtService1 := services.NewJWTService("secret1")
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}

	token, err := jwtService1.GenerateToken(user)
	assert.NoError(t, err)

	// Try to validate with different secret
	jwtService2 := services.NewJWTService("secret2")

	// Act
	claims, err := jwtService2.ValidateToken(token)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}
