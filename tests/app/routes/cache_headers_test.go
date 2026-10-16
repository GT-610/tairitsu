package routes

import (
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

// An API response can carry a session's data, and the endpoints that create a
// user or reset a password answer with a one-time password, so nothing on this
// surface may be stored by a browser or an intermediary.
func TestAPIResponsesAreNotCacheable(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Security:    config.SecurityConfig{EncryptionKey: "test-secret"},
	}

	dependencies, err := assembly.NewDependencies(config.AppConfig, nil, nil)
	require.NoError(t, err)

	app := fiber.New()
	routes.SetupRoutes(app, dependencies)

	for _, path := range []string{"/api/health", "/api/ready", "/api/system/status", "/api/networks"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, "no-store", resp.Header.Get("Cache-Control"))
		})
	}
}

// The console's hashed assets are cached for a year, so the marker must not
// have leaked onto the whole process.
func TestConsoleResponsesAreNotMarkedNoStore(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Security:    config.SecurityConfig{EncryptionKey: "test-secret"},
	}

	dependencies, err := assembly.NewDependencies(config.AppConfig, nil, nil)
	require.NoError(t, err)

	app := fiber.New()
	routes.SetupRoutes(app, dependencies)

	// A path outside /api falls through to the embedded console, which is the
	// only other thing this process serves.
	req := httptest.NewRequest(http.MethodGet, "/some/client/route", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, "no-store", resp.Header.Get("Cache-Control"))
}
