package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestStateService_GetSetupStatus_Uninitialized(t *testing.T) {
	originalConfig := config.AppConfig
	originalResetFlag := config.GetTempSetting("admin_creation_reset_done")
	t.Cleanup(func() {
		config.AppConfig = originalConfig
		config.SetTempSetting("admin_creation_reset_done", originalResetFlag)
	})

	config.AppConfig = &config.Config{
		Initialized: false,
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: "/tmp/authtoken.secret",
		},
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: "data/test.db",
		},
	}
	config.SetTempSetting("admin_creation_reset_done", "true")

	stateService := services.NewStateService()
	status := stateService.GetSetupStatus(nil, nil)

	assert.False(t, status.Initialized)
	assert.True(t, status.HasDatabase)
	assert.True(t, status.DatabaseConfigured)
	assert.False(t, status.HasAdmin)
	assert.True(t, status.ZeroTierConfigured)
	assert.True(t, status.AdminCreationPrepared)
	if assert.NotNil(t, status.DatabaseConfig) {
		assert.Equal(t, "sqlite", status.DatabaseConfig.Type)
		assert.Equal(t, "data/test.db", status.DatabaseConfig.Path)
	}
	if assert.NotNil(t, status.ZeroTierConfig) {
		assert.Equal(t, "http://127.0.0.1:9993", status.ZeroTierConfig.ControllerURL)
		assert.Equal(t, "/tmp/authtoken.secret", status.ZeroTierConfig.TokenPath)
	}
	assert.Nil(t, status.ZTStatus)
}

func TestStateService_GetSetupStatus_InitializedWithAdminAndOfflineZT(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(&stateServiceDBStub{
		users: []*models.User{
			{ID: "1", Username: "admin", Role: "admin"},
		},
	})
	networkService := services.NewNetworkService(nil, nil)

	stateService := services.NewStateService()
	status := stateService.GetSetupStatus(userService, networkService)

	assert.True(t, status.Initialized)
	assert.True(t, status.HasDatabase)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.HasAdmin)
	assert.False(t, status.ZeroTierConfigured)
	assert.Equal(t, "admin", status.AdminUsername)
	if assert.NotNil(t, status.ZTStatus) {
		assert.False(t, status.ZTStatus.Online)
		assert.Equal(t, "unknown", status.ZTStatus.Version)
	}
}
