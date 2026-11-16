package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateService_GetSetupStatus_Uninitialized(t *testing.T) {
	originalConfig := config.AppConfig
	originalResetFlag := config.GetTempSetting("admin_creation_reset_done")
	t.Cleanup(func() {
		config.AppConfig = originalConfig
		config.SetTempSetting("admin_creation_reset_done", originalResetFlag)
	})

	config.AppConfig = &config.Config{
		Initialized: false,
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: "/tmp/authtoken.secret",
		},
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}
	config.SetTempSetting("admin_creation_reset_done", "true")

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(nil, nil, true)

	assert.False(t, status.Initialized)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.SetupAuthorized)
	assert.False(t, status.HasAdmin)
	assert.True(t, status.ZeroTierConfigured)
	assert.True(t, status.AdminCreationPrepared)
	if assert.NotNil(t, status.DatabaseConfig) {
		assert.Equal(t, "sqlite", status.DatabaseConfig.Type)
		assert.Equal(t, "data/test.db", status.DatabaseConfig.Path)
	}
	if assert.NotNil(t, status.ZeroTierConfig) {
		assert.Equal(t, "http://127.0.0.1:9993", status.ZeroTierConfig.ControllerURL)
		assert.Equal(t, "/tmp/authtoken.secret", status.ZeroTierConfig.TokenPath)
	}
	assert.Nil(t, status.ZTStatus)
}

// The endpoint that serves this is public, so an unauthorized caller must not
// learn the administrator name, the database path or the controller details.
func TestStateService_GetSetupStatus_WithholdsDetailsFromPublicCallers(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: "/tmp/authtoken.secret",
		},
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(userService, nil, false)

	assert.False(t, status.SetupAuthorized)
	assert.True(t, status.HasAdmin, "the gate still needs to know an account can sign in")
	assert.Empty(t, status.AdminUsername)
	assert.Nil(t, status.DatabaseConfig)
	assert.Nil(t, status.ZeroTierConfig)

	// The flags the client routes on are not sensitive and must still be there.
	assert.True(t, status.Initialized)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.ZeroTierConfigured)
}

func TestStateService_GetSetupStatus_InitializedWithAdminAndOfflineZT(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Database: config.DatabaseConfig{
			Type: string(database.SQLite),
			Path: "data/test.db",
		},
	}

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))
	networkService := services.NewNetworkService(nil, nil)

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(userService, networkService, true)

	assert.True(t, status.Initialized)
	assert.True(t, status.DatabaseConfigured)
	assert.True(t, status.HasAdmin)
	assert.False(t, status.ZeroTierConfigured)
	// Initialized: the administrator name is no longer setup-window information,
	// so holding the token does not reveal it.
	assert.False(t, status.SetupAuthorized)
	assert.Empty(t, status.AdminUsername)
	if assert.NotNil(t, status.ZTStatus) {
		assert.False(t, status.ZTStatus.Online)
		assert.Equal(t, "unknown", status.ZTStatus.Version)
	}
}

// While setup is still pending, the token holder is told which account already
// exists so the wizard can show it.
func TestStateService_GetSetupStatus_NamesTheAdministratorDuringSetup(t *testing.T) {
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

	userService := services.NewUserService(testutil.NewFakeDB(&models.User{ID: "1", Username: "admin", Role: "admin"}))

	stateService := services.NewStateServiceWithConfig(config.AppConfig)
	status := stateService.GetSetupStatus(userService, nil, true)

	assert.False(t, status.Initialized)
	assert.True(t, status.SetupAuthorized)
	assert.True(t, status.HasAdmin)
	assert.Equal(t, "admin", status.AdminUsername)
}

// writeTokenFile places a controller token in the test's data directory and
// returns its path.
func writeTokenFile(t *testing.T, token string) string {
	t.Helper()

	path := filepath.Join(config.DataRoot(), "authtoken.secret")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(token), 0o600))
	return path
}

// TestStateService_CreateZTClientBuildsOnce is the regression test for the
// memory amplification: building the client decrypts the stored token with
// argon2id, so doing it per request let a burst of unauthenticated requests
// exhaust the process.
func TestStateService_CreateZTClientBuildsOnce(t *testing.T) {
	testutil.UseTempDataDir(t)

	tokenPath := writeTokenFile(t, "gvyqdnryjyhm9tms336s428i")
	stateService := services.NewStateServiceWithConfig(&config.Config{
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: tokenPath,
		},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	})

	first, err := stateService.CreateZTClient()
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := stateService.CreateZTClient()
	require.NoError(t, err)
	assert.Same(t, first, second, "the client must be built once, not per request")

	assert.Equal(t, "gvyqdnryjyhm9tms336s428i", second.Token)
}

func TestStateService_InvalidateZTClientDropsTheCachedClient(t *testing.T) {
	testutil.UseTempDataDir(t)

	tokenPath := writeTokenFile(t, "gvyqdnryjyhm9tms336s428i")
	stateService := services.NewStateServiceWithConfig(&config.Config{
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: tokenPath,
		},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	})

	first, err := stateService.CreateZTClient()
	require.NoError(t, err)

	stateService.InvalidateZTClient()

	second, err := stateService.CreateZTClient()
	require.NoError(t, err)
	assert.NotSame(t, first, second, "invalidation must drop the cached client")
}

// TestStateService_SaveZeroTierConfigRebuildsTheClient proves a reconfiguration
// takes effect: the cached client was built from the previous URL and token, so
// saving new ones has to discard it.
func TestStateService_SaveZeroTierConfigRebuildsTheClient(t *testing.T) {
	testutil.UseTempDataDir(t)

	tokenPath := writeTokenFile(t, "gvyqdnryjyhm9tms336s428i")
	stateService := services.NewStateServiceWithConfig(&config.Config{
		ZeroTier: config.ZeroTierConfig{
			URL:       "http://127.0.0.1:9993",
			TokenPath: tokenPath,
		},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	})

	first, err := stateService.CreateZTClient()
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9993", first.BaseURL)

	newTokenPath := filepath.Join(config.DataRoot(), "second-authtoken.secret")
	require.NoError(t, os.WriteFile(newTokenPath, []byte("2vqdnryjyhm9tms336s428iz"), 0o600))
	require.NoError(t, stateService.SaveZeroTierConfig("http://127.0.0.1:9994", newTokenPath))

	second, err := stateService.CreateZTClient()
	require.NoError(t, err)
	assert.NotSame(t, first, second)
	assert.Equal(t, "http://127.0.0.1:9994", second.BaseURL)
	assert.Equal(t, "2vqdnryjyhm9tms336s428iz", second.Token)
}

// TestStateService_GetSetupStatusDoesNotProbePerRequest is the other half of the
// amplification fix: polling the public status endpoint used to reach out to the
// controller on every single request.
func TestStateService_GetSetupStatusDoesNotProbePerRequest(t *testing.T) {
	testutil.UseTempDataDir(t)

	var calls atomic.Int32
	controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"1.14.2","address":"9bfb5c0b7e","online":true}`))
	}))
	t.Cleanup(controller.Close)

	tokenPath := writeTokenFile(t, "gvyqdnryjyhm9tms336s428i")
	stateService := services.NewStateServiceWithConfig(&config.Config{
		ZeroTier: config.ZeroTierConfig{
			URL:       controller.URL,
			TokenPath: tokenPath,
		},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	})

	for i := 0; i < 5; i++ {
		status := stateService.GetSetupStatus(nil, nil, false)
		require.NotNil(t, status.ZTStatus, "probe %d must report the controller", i)
		assert.True(t, status.ZTStatus.Online)
		assert.Equal(t, "9bfb5c0b7e", status.ZTStatus.Address)
	}

	assert.Equal(t, int32(1), calls.Load(), "five status polls must not become five controller requests")
}

// TestStateService_ProbeDoesNotPublishAcrossAReconfiguration is the regression test
// for a probe that outlives the configuration it started under: publishing its
// answer afterwards would report the previous controller for the whole cache
// lifetime.
func TestStateService_ProbeDoesNotPublishAcrossAReconfiguration(t *testing.T) {
	testutil.UseTempDataDir(t)

	// The first controller holds the request open, so its probe is provably still
	// running when the configuration changes.
	requestArrived := make(chan struct{})
	slowRelease := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestArrived)
		<-slowRelease
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"old-controller","address":"aaaaaaaaaa","online":true}`))
	}))
	t.Cleanup(slow.Close)

	// Registered after slow.Close, so it runs first: a failure before the release
	// below would otherwise leave the handler blocked, and httptest.Server.Close waits
	// for active requests — the suite would hang until the overall timeout instead of
	// reporting what failed.
	var releaseOnce sync.Once
	releaseSlow := func() { releaseOnce.Do(func() { close(slowRelease) }) }
	t.Cleanup(releaseSlow)

	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"new-controller","address":"bbbbbbbbbb","online":true}`))
	}))
	t.Cleanup(fast.Close)

	tokenPath := writeTokenFile(t, "gvyqdnryjyhm9tms336s428i")
	stateService := services.NewStateServiceWithConfig(&config.Config{
		ZeroTier: config.ZeroTierConfig{URL: slow.URL, TokenPath: tokenPath},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		stateService.GetSetupStatus(nil, nil, false)
	}()

	// Wait until the probe is inside the handler, which proves it read the old
	// configuration before the reconfiguration below.
	<-requestArrived

	newTokenPath := filepath.Join(config.DataRoot(), "second-authtoken.secret")
	require.NoError(t, os.WriteFile(newTokenPath, []byte("2vqdnryjyhm9tms336s428iz"), 0o600))
	require.NoError(t, stateService.SaveZeroTierConfig(fast.URL, newTokenPath))

	// Let the old probe finish; it must not take over the cache.
	releaseSlow()
	<-done

	status := stateService.GetSetupStatus(nil, nil, false).ZTStatus
	require.NotNil(t, status)
	assert.Equal(t, "new-controller", status.Version,
		"the cached answer must describe the configuration that is current")
}

// TestStateService_FailedSaveStillDropsTheCache covers the in-memory configuration
// already having moved when the write to disk fails: a cached client built from the
// old settings would then disagree with it.
func TestStateService_FailedSaveStillDropsTheCache(t *testing.T) {
	testutil.UseTempDataDir(t)

	tokenPath := writeTokenFile(t, "gvyqdnryjyhm9tms336s428i")
	cfg := &config.Config{
		ZeroTier: config.ZeroTierConfig{URL: "http://127.0.0.1:9993", TokenPath: tokenPath},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	}
	stateService := services.NewStateServiceWithConfig(cfg)

	first, err := stateService.CreateZTClient()
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9993", first.BaseURL)

	// Make the write itself fail, after the configuration has already been assigned.
	// A read-only directory is not reliable: a process holding CAP_DAC_OVERRIDE writes
	// into it anyway, so this would pass in a privileged container by not failing. A
	// directory where the file belongs is refused by every filesystem regardless of
	// privileges.
	dataDir := t.TempDir()
	t.Setenv(config.DataDirEnv, dataDir)
	require.NoError(t, os.Mkdir(config.ConfigFilePath(), 0o755))

	newTokenPath := filepath.Join(t.TempDir(), "authtoken.secret")
	require.NoError(t, os.WriteFile(newTokenPath, []byte("2vqdnryjyhm9tms336s428iz"), 0o600))

	saveErr := stateService.SaveZeroTierConfig("http://127.0.0.1:9994", newTokenPath)
	require.Error(t, saveErr, "precondition: the save must fail")

	// The in-memory configuration is the new one, so the client must not still be
	// the old one.
	assert.Equal(t, "http://127.0.0.1:9994", stateService.Config().ZeroTier.URL)
	second, err := stateService.CreateZTClient()
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9994", second.BaseURL,
		"a failed save must not leave a client built from the previous settings")
}
