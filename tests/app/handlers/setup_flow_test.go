package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupFlow_ResetDatabaseThenRegisterAdmin(t *testing.T) {
	originalConfig := config.AppConfig
	originalResetFlag := config.GetTempSetting("admin_creation_reset_done")
	t.Cleanup(func() {
		config.AppConfig = originalConfig
		config.SetTempSetting("admin_creation_reset_done", originalResetFlag)
	})

	// The database path is confined to Tairitsu's data directory, so the test
	// declares one and places the file inside it.
	dataRoot := t.TempDir()
	t.Setenv(config.DataDirEnv, dataRoot)

	dbPath := filepath.Join(dataRoot, "nested", "tairitsu.db")
	config.AppConfig = &config.Config{
		Initialized: false,
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: dbPath,
		},
		Security: config.SecurityConfig{
			EncryptionKey: "test-secret",
		},
	}
	config.SetTempSetting("admin_creation_reset_done", "")

	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: dbPath,
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	networkService := services.NewNetworkService(nil, db)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	setupService := services.NewSetupService(runtimeService, stateService, userService, networkService)
	systemHandler := apphandlers.NewSystemHandler(setupService, services.NewSystemService())
	authHandler := apphandlers.NewAuthHandler(userService, sessionService, runtimeService, stateService)

	app := fiber.New()
	app.Post("/system/admin/init", systemHandler.InitializeAdminCreation)
	app.Post("/auth/register", authHandler.Register)

	initReq := httptest.NewRequest(http.MethodPost, "/system/admin/init", bytes.NewReader([]byte("{}")))
	initReq.Header.Set("Content-Type", "application/json")
	initResp, err := app.Test(initReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, initResp.StatusCode)

	registerPayload, err := json.Marshal(map[string]string{
		"username": "admin",
		"password": "secret123",
	})
	require.NoError(t, err)

	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp, err := app.Test(registerReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, registerResp.StatusCode)

	users, err := userService.GetAllUsers()
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "admin", users[0].Username)
	assert.Equal(t, "admin", users[0].Role)

	secondInitReq := httptest.NewRequest(http.MethodPost, "/system/admin/init", bytes.NewReader([]byte("{}")))
	secondInitReq.Header.Set("Content-Type", "application/json")
	secondInitResp, err := app.Test(secondInitReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, secondInitResp.StatusCode)

	usersAfterSecondInit, err := userService.GetAllUsers()
	require.NoError(t, err)
	require.Len(t, usersAfterSecondInit, 1)
	assert.Equal(t, "admin", usersAfterSecondInit[0].Username)
}

func TestSetupFlow_ZeroTierTokenNotFoundReturnsOriginalDetail(t *testing.T) {
	testutil.UseTempDataDir(t)

	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	stateService := services.NewStateServiceWithConfig(&config.Config{})
	setupService := services.NewSetupService(
		services.NewRuntimeService(nil, nil, nil, stateService),
		stateService,
		nil,
		nil,
	)
	systemHandler := apphandlers.NewSystemHandler(setupService, services.NewSystemService())

	app := fiber.New()
	app.Post("/system/zerotier/config", systemHandler.SaveZeroTierConfig)
	body := bytes.NewBufferString(`{"controllerUrl":"http://127.0.0.1:9993","tokenPath":"/missing/authtoken.secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/system/zerotier/config", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var responseBody map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
	assert.Equal(t, "setup.zerotier_config_save_failed", responseBody["error_code"])
	detail, _ := responseBody["detail"].(string)
	assert.Contains(t, detail, "/missing/authtoken.secret")
	assert.Contains(t, detail, "no such file or directory")
}
