package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeRoutesDoNotUseSetupOnlyAfterInitialization(t *testing.T) {
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
	routes.SetupRoutes(app, nil, "test-secret", nil)

	paths := []string{
		"/api/networks",
		"/api/system/stats",
		"/api/users",
		"/api/admin/networks/importable",
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		require.NoError(t, err, path)
		assert.NotEqual(t, fiber.StatusConflict, resp.StatusCode, path)
	}
}
