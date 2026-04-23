package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginRouteDoesNotUseSetupOnlyAfterInitialization(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	app := fiber.New()
	routes.SetupRoutes(app, assembly.NewDependencies(config.AppConfig, nil, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.NotEqual(t, fiber.StatusConflict, resp.StatusCode)
}

func TestRegisterRouteDoesNotUseSetupOnlyAfterInitialization(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	app := fiber.New()
	routes.SetupRoutes(app, assembly.NewDependencies(config.AppConfig, nil, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{"username":"user-1","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.NotEqual(t, fiber.StatusConflict, resp.StatusCode)
}
