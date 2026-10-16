package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemHandler_GetSystemStatus_Uninitialized(t *testing.T) {
	originalConfig := config.AppConfig
	originalResetFlag := config.GetTempSetting("admin_creation_reset_done")
	t.Cleanup(func() {
		config.AppConfig = originalConfig
		config.SetTempSetting("admin_creation_reset_done", originalResetFlag)
	})

	config.AppConfig = &config.Config{
		Initialized: false,
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/setup.db",
		},
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: "/tmp/authtoken.secret",
		},
	}
	config.SetTempSetting("admin_creation_reset_done", "true")

	userService := services.NewUserService(nil)
	sessionService := services.NewSessionService(nil)
	networkService := services.NewNetworkService(nil, nil)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	handler := apphandlers.NewSystemHandler(services.NewSetupService(runtimeService, stateService, userService, networkService), services.NewSystemService())

	app := fiber.New()
	app.Get("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	resp, err := app.Test(req, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, false, body["initialized"])
	assert.Equal(t, true, body["databaseConfigured"])
	assert.Equal(t, true, body["zerotierConfigured"])
	assert.Equal(t, true, body["adminCreationPrepared"])
	assert.Equal(t, true, body["allowPublicRegistration"])
}

func TestSystemHandler_GetSystemStatus_Initialized(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Registration: config.RegistrationConfig{
			AllowPublicRegistration: boolPtr(false),
		},
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))
	sessionService := services.NewSessionService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))
	networkService := services.NewNetworkService(nil, nil)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	handler := apphandlers.NewSystemHandler(services.NewSetupService(runtimeService, stateService, userService, networkService), services.NewSystemService())

	app := fiber.New()
	app.Get("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	resp, err := app.Test(req, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body struct {
		Initialized             bool `json:"initialized"`
		HasAdmin                bool `json:"hasAdmin"`
		AllowPublicRegistration bool `json:"allowPublicRegistration"`
		ZTStatus                struct {
			Online  bool   `json:"online"`
			Version string `json:"version"`
		} `json:"ztStatus"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body.Initialized)
	assert.True(t, body.HasAdmin)
	assert.False(t, body.AllowPublicRegistration)
	assert.False(t, body.ZTStatus.Online)
	assert.Equal(t, "unknown", body.ZTStatus.Version)
}

func boolPtr(value bool) *bool {
	return &value
}

func TestSystemHandler_RuntimeSettingsReadWrite(t *testing.T) {
	// The settings endpoint persists the configuration, so it must not land in
	// the source tree.
	testutil.UseTempDataDir(t)

	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Registration: config.RegistrationConfig{
			AllowPublicRegistration: boolPtr(true),
		},
	}

	userService := services.NewUserService(nil)
	sessionService := services.NewSessionService(nil)
	networkService := services.NewNetworkService(nil, nil)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	handler := apphandlers.NewSystemHandler(services.NewSetupService(runtimeService, stateService, userService, networkService), services.NewSystemService())

	app := fiber.New()
	app.Get("/system/settings", handler.GetRuntimeSettings)
	app.Put("/system/settings", handler.UpdateRuntimeSettings)

	getReq := httptest.NewRequest(http.MethodGet, "/system/settings", nil)
	getResp, err := app.Test(getReq, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, getResp.StatusCode)

	var getBody struct {
		AllowPublicRegistration bool `json:"allow_public_registration"`
	}
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&getBody))
	assert.True(t, getBody.AllowPublicRegistration)

	updateReq := httptest.NewRequest(http.MethodPut, "/system/settings", bytes.NewBufferString(`{"allow_public_registration":false}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := app.Test(updateReq, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, updateResp.StatusCode)
	assert.False(t, config.AllowPublicRegistration(config.AppConfig))
}

func TestSystemHandler_SetInitializedRejectsMissingAdmin(t *testing.T) {
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

	stateDB := testutil.NewFakeDB()
	userService := services.NewUserService(stateDB)
	sessionService := services.NewSessionService(stateDB)
	networkService := services.NewNetworkService(nil, stateDB)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	handler := apphandlers.NewSystemHandler(services.NewSetupService(runtimeService, stateService, userService, networkService), services.NewSystemService())

	app := fiber.New()
	app.Post("/system/initialized", handler.SetInitialized)

	req := httptest.NewRequest(http.MethodPost, "/system/initialized", bytes.NewBufferString(`{"initialized":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Contains(t, body["message"], "Create the first administrator account first")
}
