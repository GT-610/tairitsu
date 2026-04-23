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

	dbPath := filepath.Join(t.TempDir(), "nested", "tairitsu.db")
	config.AppConfig = &config.Config{
		Initialized: false,
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: dbPath,
		},
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}
	config.SetTempSetting("admin_creation_reset_done", "")

	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: dbPath,
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())

	userService := services.NewUserServiceWithDB(db)
	networkService := services.NewNetworkService(nil, db)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, networkService, stateService)
	setupService := services.NewSetupService(runtimeService, stateService, userService, networkService)
	systemHandler := apphandlers.NewSystemHandler(setupService, services.NewSystemService())
	authHandler := apphandlers.NewAuthHandler(userService, services.NewJWTService("test-secret"), runtimeService, stateService)

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

	users := userService.GetAllUsers()
	require.Len(t, users, 1)
	assert.Equal(t, "admin", users[0].Username)
	assert.Equal(t, "admin", users[0].Role)
}
