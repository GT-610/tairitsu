package services

import (
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateServiceWithConfigUsesBoundConfig(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	boundConfig := &config.Config{
		Initialized: true,
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: filepath.Join(t.TempDir(), "state.db"),
		},
	}
	config.AppConfig = &config.Config{Initialized: false}

	stateService := appservices.NewStateServiceWithConfig(boundConfig)

	assert.True(t, stateService.IsInitialized())
	assert.True(t, stateService.DatabaseConfigured())
}

func TestStateServiceSetInitializedPersistsBoundConfig(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	boundConfig := &config.Config{
		Initialized: false,
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: filepath.Join(t.TempDir(), "persist.db"),
		},
	}
	config.AppConfig = boundConfig

	stateService := appservices.NewStateServiceWithConfig(boundConfig)
	require.NoError(t, stateService.SetInitialized(true))

	assert.True(t, boundConfig.Initialized)
	assert.True(t, config.AppConfig.Initialized)
}
