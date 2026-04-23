package handlers

import (
	"bytes"
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

func TestAuthHandler_RegisterRejectsWhenPublicRegistrationDisabled(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	dbPath := filepath.Join(t.TempDir(), "tairitsu.db")
	config.AppConfig = &config.Config{
		Initialized: true,
		Registration: config.RegistrationConfig{
			AllowPublicRegistration: boolPtr(false),
		},
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: dbPath,
		},
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: dbPath,
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	userService := services.NewUserServiceWithDB(db)
	networkService := services.NewNetworkService(nil, db)
	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	runtimeService := services.NewRuntimeService(userService, networkService, stateService)
	authHandler := apphandlers.NewAuthHandler(userService, services.NewJWTService("test-secret"), runtimeService, stateService)

	app := fiber.New()
	app.Post("/auth/register", authHandler.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"username":"user-1","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
