package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
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
	config.AppConfig = &config.Config{Initialized: false}

	stateService := appservices.NewStateServiceWithConfig(boundConfig)
	require.NoError(t, stateService.SetInitialized(true))

	assert.True(t, boundConfig.Initialized)
	assert.False(t, config.AppConfig.Initialized)
}

func TestStateServiceDatabaseConfigUsesBoundConfig(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	boundConfig := &config.Config{
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: filepath.Join(t.TempDir(), "bound.db"),
		},
	}
	config.AppConfig = &config.Config{}

	stateService := appservices.NewStateServiceWithConfig(boundConfig)
	dbConfig := stateService.DatabaseConfig()

	assert.Equal(t, "sqlite", string(dbConfig.Type))
	assert.Equal(t, boundConfig.Database.Path, dbConfig.Path)
}

func TestStateServiceSaveDatabaseConfigPersistsBoundConfig(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	boundConfig := &config.Config{
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}
	config.AppConfig = &config.Config{}

	stateService := appservices.NewStateServiceWithConfig(boundConfig)
	require.NoError(t, stateService.SaveDatabaseConfig(database.Config{
		Type: database.SQLite,
		Path: filepath.Join(t.TempDir(), "saved.db"),
	}))

	assert.Equal(t, config.SQLite, boundConfig.Database.Type)
	assert.NotEmpty(t, boundConfig.Database.Path)
	assert.Empty(t, config.AppConfig.Database.Path)
}

func TestStateServiceSaveZeroTierConfigPersistsBoundConfig(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	tokenPath := filepath.Join(t.TempDir(), "authtoken.secret")
	require.NoError(t, os.WriteFile(tokenPath, []byte("secret-token\n"), 0600))

	boundConfig := &config.Config{
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}
	config.AppConfig = &config.Config{}

	stateService := appservices.NewStateServiceWithConfig(boundConfig)
	require.NoError(t, stateService.SaveZeroTierConfig("http://127.0.0.1:9993", tokenPath))

	assert.Equal(t, "http://127.0.0.1:9993", boundConfig.ZeroTier.URL)
	assert.Equal(t, tokenPath, boundConfig.ZeroTier.TokenPath)
	assert.NotEmpty(t, boundConfig.ZeroTier.Token)
	assert.Empty(t, config.AppConfig.ZeroTier.URL)
}
