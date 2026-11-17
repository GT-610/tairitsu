package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// setupFixture wires the setup surface the way the process does: the database
// exists, the instance is not initialized, and no administrator has been
// created yet. This is the window the first administrator is created in.
type setupFixture struct {
	app         *fiber.App
	userService *services.UserService
	setupSvc    *services.SetupService
	config      *config.Config
}

func newSetupFixture(t *testing.T, initialized bool) *setupFixture {
	t.Helper()

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
	cfg := &config.Config{
		Initialized: initialized,
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: dbPath,
		},
		Security: config.SecurityConfig{
			EncryptionKey: "test-secret",
		},
	}
	config.AppConfig = cfg
	config.SetTempSetting("admin_creation_reset_done", "")

	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: dbPath,
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	networkService := services.NewNetworkService(nil, db)
	stateService := services.NewStateServiceWithConfig(cfg)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	setupService := services.NewSetupService(runtimeService, stateService, userService, networkService)
	systemHandler := apphandlers.NewSystemHandler(setupService, services.NewSystemService())
	authHandler := apphandlers.NewAuthHandler(userService, sessionService, stateService)

	app := fiber.New()
	app.Post("/system/admin/init", systemHandler.InitializeAdminCreation)
	app.Post("/system/admin/register", systemHandler.RegisterInitialAdmin)
	app.Post("/auth/register", authHandler.Register)

	return &setupFixture{app: app, userService: userService, setupSvc: setupService, config: cfg}
}

func (f *setupFixture) post(t *testing.T, path string, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.app.Test(req, testRequestTimeout)
	require.NoError(t, err)
	return resp
}

func (f *setupFixture) rolesByUsername(t *testing.T) map[string]string {
	t.Helper()
	users, err := f.userService.GetAllUsers()
	require.NoError(t, err)
	roles := make(map[string]string, len(users))
	for _, user := range users {
		roles[user.Username] = user.Role
	}
	return roles
}

func TestSetupFlow_ResetDatabaseThenRegisterAdmin(t *testing.T) {
	f := newSetupFixture(t, false)

	initResp := f.post(t, "/system/admin/init", "{}")
	assert.Equal(t, fiber.StatusOK, initResp.StatusCode)

	registerResp := f.post(t, "/system/admin/register", `{"username":"admin","password":"secret123"}`)
	assert.Equal(t, fiber.StatusCreated, registerResp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(registerResp.Body).Decode(&body))
	created, ok := body["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "admin", created["username"])
	assert.Equal(t, "admin", created["role"])

	users, err := f.userService.GetAllUsers()
	require.NoError(t, err)
	require.Len(t, users, 1)

	// A second init must not wipe the account that was just created.
	secondInitResp := f.post(t, "/system/admin/init", "{}")
	assert.Equal(t, fiber.StatusOK, secondInitResp.StatusCode)

	usersAfterSecondInit, err := f.userService.GetAllUsers()
	require.NoError(t, err)
	require.Len(t, usersAfterSecondInit, 1)
	assert.Equal(t, "admin", usersAfterSecondInit[0].Username)
}

// TestSetupFlow_PublicRegistrationNeverGrantsAdmin is the regression test for
// the first-run takeover: public registration answers before anyone can sign
// in, so it must never be the thing that mints an administrator.
func TestSetupFlow_PublicRegistrationNeverGrantsAdmin(t *testing.T) {
	f := newSetupFixture(t, false)

	// The operator has configured the database, and the window is open. An
	// unauthenticated caller registers first.
	resp := f.post(t, "/auth/register", `{"username":"attacker","password":"pwned123"}`)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	user, ok := body["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "user", user["role"], "public registration must not grant the administrator role")

	assert.Equal(t, map[string]string{"attacker": "user"}, f.rolesByUsername(t))
}

// TestSetupFlow_AdminRegistrationClaimsRoleOnce proves the role is claimed, not
// assigned: a second attempt loses to the first and leaves no stray account.
func TestSetupFlow_AdminRegistrationClaimsRoleOnce(t *testing.T) {
	f := newSetupFixture(t, false)

	first := f.post(t, "/system/admin/register", `{"username":"first-admin","password":"secret123"}`)
	require.Equal(t, fiber.StatusCreated, first.StatusCode)

	second := f.post(t, "/system/admin/register", `{"username":"second-admin","password":"secret123"}`)
	assert.Equal(t, fiber.StatusConflict, second.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(second.Body).Decode(&body))
	assert.Equal(t, "user.admin_already_exists", body["error_code"])

	// The losing attempt is rolled back with its transaction, so the account it
	// tried to create does not linger as a usable ordinary account.
	assert.Equal(t, map[string]string{"first-admin": "admin"}, f.rolesByUsername(t))
}

// TestSetupFlow_AdminRegistrationRequiresUninitializedInstance encodes the gate:
// once setup is complete the endpoint is closed.
func TestSetupFlow_AdminRegistrationRejectsInitializedInstance(t *testing.T) {
	f := newSetupFixture(t, true)

	resp := f.post(t, "/system/admin/register", `{"username":"late-admin","password":"secret123"}`)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "setup.already_initialized", body["error_code"])
	assert.Empty(t, f.rolesByUsername(t))
}

// TestSetupFlow_AdminRegistrationValidatesAccountShape keeps the account rules in
// one place: the setup path is not a way around the password policy.
func TestSetupFlow_AdminRegistrationValidatesAccountShape(t *testing.T) {
	f := newSetupFixture(t, false)

	short := f.post(t, "/system/admin/register", `{"username":"admin","password":"short"}`)
	assert.Equal(t, fiber.StatusBadRequest, short.StatusCode)

	long := f.post(t, "/system/admin/register",
		`{"username":"admin","password":"`+string(bytes.Repeat([]byte("x"), 40))+`"}`)
	assert.Equal(t, fiber.StatusBadRequest, long.StatusCode)

	empty := f.post(t, "/system/admin/register", `{"username":"  ","password":"secret123"}`)
	assert.Equal(t, fiber.StatusBadRequest, empty.StatusCode)

	assert.Empty(t, f.rolesByUsername(t))
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
	resp, err := app.Test(req, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var responseBody map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
	assert.Equal(t, "setup.zerotier_config_save_failed", responseBody["error_code"])
	detail, _ := responseBody["detail"].(string)
	assert.Contains(t, detail, "/missing/authtoken.secret")
	assert.Contains(t, detail, "no such file or directory")
}

// TestSetupFlow_ControllerURLMustBeAnHTTPAddress keeps a malformed address from
// reaching the client, where it would surface as a connection failure rather than
// as the mistake it is. It also keeps a scheme the transport does not understand
// from being stored as the place the controller token gets sent.
func TestSetupFlow_ControllerURLMustBeAnHTTPAddress(t *testing.T) {
	testutil.UseTempDataDir(t)

	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	tokenPath := filepath.Join(t.TempDir(), "authtoken.secret")
	require.NoError(t, os.WriteFile(tokenPath, []byte("gvyqdnryjyhm9tms336s428i\n"), 0o600))

	cfg := &config.Config{Security: config.SecurityConfig{EncryptionKey: "test-secret-key"}}
	config.AppConfig = cfg

	stateService := services.NewStateServiceWithConfig(cfg)
	setupService := services.NewSetupService(
		services.NewRuntimeService(nil, nil, nil, stateService),
		stateService,
		nil,
		nil,
	)
	systemHandler := apphandlers.NewSystemHandler(setupService, services.NewSystemService())

	app := fiber.New()
	app.Post("/system/zerotier/config", systemHandler.SaveZeroTierConfig)

	for _, badURL := range []string{"gopher://127.0.0.1:9993", "file:///etc/passwd", "127.0.0.1:9993"} {
		body, err := json.Marshal(map[string]string{"controllerUrl": badURL, "tokenPath": tokenPath})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/system/zerotier/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, testRequestTimeout)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
		require.NoError(t, resp.Body.Close())

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "url %q", badURL)
		assert.Equal(t, "setup.invalid_controller_url", payload["error_code"], "url %q", badURL)
	}

	assert.Empty(t, cfg.ZeroTier.URL, "a rejected address must not be stored")
}
