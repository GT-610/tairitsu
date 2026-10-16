package routes

import (
	"io"
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

// newTestApp assembles the real route table around a configuration with only
// Initialized set. The database and controller are nil, which is enough to
// reach the middleware chain.
func newTestApp(t *testing.T, initialized bool) *fiber.App {
	t.Helper()

	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: initialized,
		Security:    config.SecurityConfig{EncryptionKey: "test-secret"},
	}

	dependencies, err := assembly.NewDependencies(config.AppConfig, nil, nil)
	require.NoError(t, err)

	app := fiber.New()
	routes.SetupRoutes(app, dependencies)

	return app
}

func getStatus(t *testing.T, app *fiber.App, path string) int {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp.StatusCode
}

func postStatus(t *testing.T, app *fiber.App, path string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(body)
}

// Once setup is complete the setup endpoints have to close, and the runtime
// surface has to be reachable. Both halves are asserted on the response itself:
// a route that is missing from the table answers 404 with system.not_found, and
// one whose gate misfires answers 409, so anything else proves the wiring.
func TestRouteGatingAfterInitialization(t *testing.T) {
	app := newTestApp(t, true)

	setupPaths := []string{
		"/api/system/database",
		"/api/system/zerotier/config",
		"/api/system/initialized",
		"/api/system/admin/init",
	}

	for _, path := range setupPaths {
		t.Run("setup endpoint closes "+path, func(t *testing.T) {
			status, body := postStatus(t, app, path)

			assert.Equal(t, fiber.StatusConflict, status)
			assert.Contains(t, body, "system.already_initialized")
		})
	}

	runtimePaths := []string{
		"/api/networks",
		"/api/networks/shared",
		"/api/system/stats",
		"/api/users",
		"/api/admin/networks/importable",
		"/api/admin/planet/identity",
	}

	for _, path := range runtimePaths {
		t.Run("runtime route stays open "+path, func(t *testing.T) {
			// Authorization still refuses the anonymous caller, which is what
			// proves the request reached the route rather than a gate.
			assert.Equal(t, fiber.StatusUnauthorized, getStatus(t, app, path))
		})
	}
}

// Before setup the two gates have to be the other way round.
func TestRouteGatingBeforeInitialization(t *testing.T) {
	app := newTestApp(t, false)

	status, body := postStatus(t, app, "/api/auth/login")

	assert.Equal(t, fiber.StatusServiceUnavailable, status)
	assert.Contains(t, body, "system.setup_required")

	for _, path := range []string{"/api/networks", "/api/users", "/api/admin/planet/identity"} {
		t.Run("runtime route stays closed "+path, func(t *testing.T) {
			assert.Equal(t, fiber.StatusServiceUnavailable, getStatus(t, app, path))
		})
	}
}
