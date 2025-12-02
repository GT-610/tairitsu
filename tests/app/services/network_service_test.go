package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestNewNetworkService(t *testing.T) {
	// Act
	networkService := services.NewNetworkService(nil, nil)

	// Assert
	assert.NotNil(t, networkService)
}
