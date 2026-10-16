package routes

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// The deployment every route test starts from. A test opts out of the parts it
// does not want rather than restating the whole configuration.
const (
	testEncryptionKey     = "test-secret"
	testDatabasePath      = "/app/data/tairitsu.db"
	testZeroTierURL       = "http://127.0.0.1:9993"
	testZeroTierTokenPath = "/var/lib/zerotier-one/authtoken.secret"
)

// appConfig describes the configuration a route test needs.
type appConfig struct {
	initialized     bool
	omitDatabase    bool
	omitZeroTier    bool
	useTempDataRoot bool
}

type appOption func(*appConfig)

func initialized(value bool) appOption {
	return func(cfg *appConfig) { cfg.initialized = value }
}

// withoutDatabase and withoutZeroTier leave the respective section empty, so a
// test can exercise an unconfigured deployment.
func withoutDatabase() appOption {
	return func(cfg *appConfig) { cfg.omitDatabase = true }
}

func withoutZeroTier() appOption {
	return func(cfg *appConfig) { cfg.omitZeroTier = true }
}

// withTempDataRoot points the configuration file at a temporary directory. Any
// test that lets a request or a service write the configuration has to use it,
// or the file lands in the source tree.
func withTempDataRoot() appOption {
	return func(cfg *appConfig) { cfg.useTempDataRoot = true }
}

// newTestApp assembles the real route table around a configuration. The frontend
// mount registers a catch-all, so tests must target endpoints that already exist
// rather than adding routes afterwards: anything registered later is served by
// the app shell instead.
//
// The encryption key is always set, because assembly mints one when the
// configuration has none and a test should not depend on that side effect.
func newTestApp(t *testing.T, options ...appOption) (*fiber.App, *assembly.Dependencies) {
	t.Helper()

	cfg := appConfig{}
	for _, option := range options {
		option(&cfg)
	}

	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	if cfg.useTempDataRoot {
		t.Setenv(config.DataDirEnv, t.TempDir())
	}

	appConfigValue := &config.Config{
		Initialized: cfg.initialized,
		Security:    config.SecurityConfig{EncryptionKey: testEncryptionKey},
	}
	if !cfg.omitDatabase {
		appConfigValue.Database = config.DatabaseConfig{Type: "sqlite", Path: testDatabasePath}
	}
	if !cfg.omitZeroTier {
		appConfigValue.ZeroTier = config.ZeroTierConfig{URL: testZeroTierURL, TokenPath: testZeroTierTokenPath}
	}
	config.AppConfig = appConfigValue

	dependencies, err := assembly.NewDependencies(config.AppConfig, nil, nil)
	require.NoError(t, err)

	app := fiber.New()
	routes.SetupRoutes(app, dependencies)

	return app, dependencies
}
