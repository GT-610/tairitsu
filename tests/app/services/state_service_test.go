package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/tests/testutil"
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
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}
	config.SetTempSetting("admin_creation_reset_done", "true")

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(nil, nil, true)

	assert.False(t, status.Initialized)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.SetupAuthorized)
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

// The endpoint that serves this is public, so an unauthorized caller must not
// learn the administrator name, the database path or the controller details.
func TestStateService_GetSetupStatus_WithholdsDetailsFromPublicCallers(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: "/tmp/authtoken.secret",
		},
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(userService, nil, false)

	assert.False(t, status.SetupAuthorized)
	assert.True(t, status.HasAdmin, "the gate still needs to know an account can sign in")
	assert.Empty(t, status.AdminUsername)
	assert.Nil(t, status.DatabaseConfig)
	assert.Nil(t, status.ZeroTierConfig)

	// The flags the client routes on are not sensitive and must still be there.
	assert.True(t, status.Initialized)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.ZeroTierConfigured)
}

func TestStateService_GetSetupStatus_InitializedWithAdminAndOfflineZT(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))
	networkService := services.NewNetworkService(nil, nil)

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(userService, networkService, true)

	assert.True(t, status.Initialized)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.HasAdmin)
	assert.False(t, status.ZeroTierConfigured)
	// Initialized: the administrator name is no longer setup-window information,
	// so holding the token does not reveal it.
	assert.False(t, status.SetupAuthorized)
	assert.Empty(t, status.AdminUsername)
	if assert.NotNil(t, status.ZTStatus) {
		assert.False(t, status.ZTStatus.Online)
		assert.Equal(t, "unknown", status.ZTStatus.Version)
	}
}

// While setup is still pending, the token holder is told which account already
// exists so the wizard can show it.
func TestStateService_GetSetupStatus_NamesTheAdministratorDuringSetup(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: false,
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(userService, nil, true)

	assert.False(t, status.Initialized)
	assert.True(t, status.SetupAuthorized)
	assert.True(t, status.HasAdmin)
	assert.Equal(t, "admin", status.AdminUsername)
}
