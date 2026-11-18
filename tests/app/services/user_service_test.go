package services

import (
	"testing"

	"github.com/GT-610/tempest-zt/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestNewUserServiceWithNilDB(t *testing.T) {
	// Act
	userService := services.NewUserService(nil)

	// Assert
	assert.NotNil(t, userService)
}
