package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newUninitializedApp builds the real route table for an instance that has not
// been set up yet, which is the window the setup token protects.
func newUninitializedApp(t *testing.T, dataRoot string) (*fiber.App, *assembly.Dependencies) {
	t.Helper()

	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})
	t.Setenv(config.DataDirEnv, dataRoot)

	config.AppConfig = &config.Config{
		Initialized: false,
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dataRoot + "/tairitsu.db",
		},
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: "/var/lib/zerotier-one/authtoken.secret",
		},
		Security: config.SecurityConfig{EncryptionKey: "test-secret"},
	}

	dependencies, err := assembly.NewDependencies(config.AppConfig, nil, nil)
	require.NoError(t, err)
	require.True(t, dependencies.Services.SetupToken.Issued(), "an uninitialized instance must issue a token")

	app := fiber.New()
	routes.SetupRoutes(app, dependencies)
	return app, dependencies
}

func postJSON(t *testing.T, app *fiber.App, path string, body string, setupToken string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if setupToken != "" {
		req.Header.Set(middleware.SetupTokenHeader, setupToken)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func TestSetupEndpointsRequireTheSetupToken(t *testing.T) {
	app, dependencies := newUninitializedApp(t, t.TempDir())
	validToken := dependencies.Services.SetupToken.Value()

	paths := []struct {
		path string
		body string
	}{
		{path: "/api/system/database", body: `{"type":"sqlite"}`},
		{path: "/api/system/zerotier/config", body: `{"controllerUrl":"http://127.0.0.1:9993","tokenPath":"/tmp/authtoken.secret"}`},
		{path: "/api/system/initialized", body: `{"initialized":true}`},
		{path: "/api/system/admin/init", body: `{}`},
	}

	cases := []struct {
		name       string
		setupToken string
		wantStatus int
		wantCode   string
	}{
		{name: "no token", setupToken: "", wantStatus: fiber.StatusForbidden, wantCode: "setup.token_required"},
		{name: "wrong token", setupToken: "not-the-token", wantStatus: fiber.StatusForbidden, wantCode: "setup.token_required"},
		{name: "valid token", setupToken: validToken, wantStatus: 0},
	}

	for _, path := range paths {
		for _, testCase := range cases {
			t.Run(testCase.name+" "+path.path, func(t *testing.T) {
				resp := postJSON(t, app, path.path, path.body, testCase.setupToken)
				defer resp.Body.Close()

				if testCase.wantStatus == 0 {
					assert.NotEqual(t, fiber.StatusForbidden, resp.StatusCode,
						"a request carrying the issued token must get past the gate")
					return
				}

				assert.Equal(t, testCase.wantStatus, resp.StatusCode)

				var body map[string]any
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
				assert.Equal(t, testCase.wantCode, body["error_code"])
			})
		}
	}
}

func TestSystemStatusHidesSetupDetailsFromAnonymousCallers(t *testing.T) {
	app, _ := newUninitializedApp(t, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/system/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, false, body["setupAuthorized"])
	assert.NotContains(t, body, "adminUsername")
	assert.NotContains(t, body, "databaseConfig")
	assert.NotContains(t, body, "zeroTierConfig")
}

func TestSystemStatusIncludesSetupDetailsForTheTokenHolder(t *testing.T) {
	app, dependencies := newUninitializedApp(t, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/system/status", nil)
	req.Header.Set(middleware.SetupTokenHeader, dependencies.Services.SetupToken.Value())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body struct {
		SetupAuthorized bool `json:"setupAuthorized"`
		DatabaseConfig  *struct {
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"databaseConfig"`
		ZeroTierConfig *struct {
			ControllerURL string `json:"controllerUrl"`
			TokenPath     string `json:"tokenPath"`
		} `json:"zeroTierConfig"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.True(t, body.SetupAuthorized)
	require.NotNil(t, body.DatabaseConfig)
	assert.Equal(t, "sqlite", body.DatabaseConfig.Type)
	require.NotNil(t, body.ZeroTierConfig)
	assert.Equal(t, "http://127.0.0.1:9993", body.ZeroTierConfig.ControllerURL)
}

// Once setup is complete the endpoints are closed by SetupOnly and no token is
// issued at all, so the gate cannot be the only thing standing in the way.
func TestSetupEndpointsAreClosedAfterInitialization(t *testing.T) {
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
	assert.False(t, dependencies.Services.SetupToken.Issued())

	app := fiber.New()
	routes.SetupRoutes(app, dependencies)

	resp := postJSON(t, app, "/api/system/admin/init", `{}`, dependencies.Services.SetupToken.Value())
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
}
