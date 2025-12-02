package config

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Act
	cfg, err := config.LoadConfig()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestTempSettings(t *testing.T) {
	// Arrange
	testKey := "test-key"
	testValue := "test-value"

	// Act - Set temp setting
	config.SetTempSetting(testKey, testValue)

	// Assert - Get temp setting
	retrievedValue := config.GetTempSetting(testKey)
	assert.Equal(t, testValue, retrievedValue)

	// Act - Set another value for the same key
	newValue := "new-test-value"
	config.SetTempSetting(testKey, newValue)

	// Assert - Get updated temp setting
	retrievedValue = config.GetTempSetting(testKey)
	assert.Equal(t, newValue, retrievedValue)

	// Act - Get non-existent key
	nonExistentValue := config.GetTempSetting("non-existent-key")

	// Assert - Should return empty string for non-existent key
	assert.Empty(t, nonExistentValue)
}

func TestIsInitialized_Default(t *testing.T) {
	// Act
	isInitialized := config.IsInitialized()

	// Assert
	assert.False(t, isInitialized)
}
