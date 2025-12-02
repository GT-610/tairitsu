package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestNewUserServiceWithoutDB(t *testing.T) {
	// Act
	userService := services.NewUserServiceWithoutDB()

	// Assert
	assert.NotNil(t, userService)
}
