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
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: false,
	}

	stateService := services.NewStateService()
	status := stateService.GetSetupStatus(nil, nil)

	assert.False(t, status.Initialized)
	assert.False(t, status.HasDatabase)
	assert.False(t, status.HasAdmin)
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

	userService := services.NewUserServiceWithDB(&stateServiceDBStub{
		users: []*models.User{
			{ID: "1", Username: "admin", Role: "admin"},
		},
	})
	networkService := services.NewNetworkService(nil, nil)

	stateService := services.NewStateService()
	status := stateService.GetSetupStatus(userService, networkService)

	assert.True(t, status.Initialized)
	assert.True(t, status.HasDatabase)
	assert.True(t, status.HasAdmin)
	if assert.NotNil(t, status.ZTStatus) {
		assert.False(t, status.ZTStatus.Online)
		assert.Equal(t, "unknown", status.ZTStatus.Version)
	}
}
