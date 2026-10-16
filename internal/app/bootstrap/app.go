package bootstrap

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type App struct {
	Config       *config.Config
	Database     database.DBInterface
	ZTClient     *zerotier.Client
	Dependencies *assembly.Dependencies
	Router       *fiber.App
	cancel       context.CancelFunc
	cleanupDone  <-chan struct{}
}

func Build() (*App, error) {
	logger.InitLogger("info")
	logger.Info("starting application assembly")

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	app := &App{Config: cfg}

	if err := app.initializeDatabase(); err != nil {
		if cfg.Initialized {
			return nil, fmt.Errorf("system is initialized, but database initialization failed: %w", err)
		}
		logger.Warn("database initialization failed; continuing in uninitialized mode", zap.Error(err))
	}

	if err := app.initializeZeroTierClient(); err != nil {
		if cfg.Initialized {
			return nil, fmt.Errorf("system is initialized, but ZeroTier client initialization failed: %w", err)
		}
		logger.Warn("ZeroTier client initialization failed; continuing in uninitialized mode", zap.Error(err))
	}

	dependencies, err := assembly.NewDependencies(app.Config, app.Database, app.ZTClient)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble dependencies: %w", err)
	}
	app.Dependencies = dependencies
	app.Dependencies.Services.Session.PurgeUnusableSessions()
	announceSetupToken(app.Dependencies.Services.SetupToken)
	app.Router = newHTTPApp()
	routes.SetupRoutes(app.Router, app.Dependencies)

	ctx, cancel := context.WithCancel(context.Background())
	app.cancel = cancel
	app.cleanupDone = app.Dependencies.Services.Session.StartCleanup(ctx)

	logger.Info("application assembly completed")
	return app, nil
}

// Environment variables that describe the network between the server and its
// callers. They are deployment properties rather than stored settings, so they
// are read from the environment on every start instead of being persisted.
const (
	// TrustedProxiesEnv lists the addresses or CIDR ranges of reverse proxies
	// whose forwarding headers may be believed.
	TrustedProxiesEnv = "TAIRITSU_TRUSTED_PROXIES"
	// ProxyHeaderEnv overrides the header carrying the client address.
	ProxyHeaderEnv = "TAIRITSU_PROXY_HEADER"
	// TrustLoopbackProxyEnv turns off the default trust for a proxy on loopback.
	TrustLoopbackProxyEnv = "TAIRITSU_TRUSTED_PROXY_LOOPBACK"
)

// defaultProxyHeader is the header reverse proxies use to carry the client
// address. X-Forwarded-For is set by nginx, Caddy, Traefik and HAProxy alike;
// the previously hardcoded X-Real-IP only exists in nginx.
const defaultProxyHeader = "X-Forwarded-For"

// announceSetupToken makes the first-run token impossible to miss. It is
// written to standard output as well as the log, because the log file lives
// inside the container while `docker logs` only shows what the process prints.
func announceSetupToken(token *services.SetupToken) {
	if token == nil || !token.Issued() {
		return
	}

	value := token.Value()
	fmt.Fprintf(os.Stdout, "\n  Tairitsu is not initialized yet.\n  Setup token: %s\n  Enter it in the setup wizard. A restart issues a new token.\n\n", value)
	logger.Warn("first-run setup token issued; the setup endpoints require it", zap.String("setup_token", value))
}

func newHTTPApp() *fiber.App {
	return fiber.New(httpAppConfig())
}

func httpAppConfig() fiber.Config {
	return fiber.Config{
		TrustProxy:  true,
		ProxyHeader: proxyHeader(),
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies:  trustedProxies(),
			Loopback: trustLoopbackProxy(),
		},
		EnableIPValidation: true,
	}
}

// proxyHeader returns the header a trusted proxy is expected to set.
func proxyHeader() string {
	if header := config.Env(ProxyHeaderEnv); header != "" {
		return header
	}
	return defaultProxyHeader
}

// trustedProxies returns the configured proxy addresses and CIDR ranges,
// dropping entries that cannot be parsed so that a typo degrades to "not
// trusted" with a warning instead of being silently accepted.
func trustedProxies() []string {
	configured := config.EnvList(TrustedProxiesEnv)
	accepted := make([]string, 0, len(configured))

	for _, entry := range configured {
		if isValidProxyEntry(entry) {
			accepted = append(accepted, entry)
			continue
		}
		logger.Warn("ignoring unparsable trusted proxy entry", zap.String("entry", entry))
	}

	if len(accepted) == 0 {
		return nil
	}
	return accepted
}

func isValidProxyEntry(entry string) bool {
	if _, _, err := net.ParseCIDR(entry); err == nil {
		return true
	}
	return net.ParseIP(entry) != nil
}

// trustLoopbackProxy reports whether a proxy connecting over loopback may set
// forwarding headers. It defaults to true, which covers a proxy in the same
// container or on the host's loopback and nothing else.
func trustLoopbackProxy() bool {
	raw := config.Env(TrustLoopbackProxyEnv)
	if raw == "" {
		return true
	}

	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		logger.Warn("ignoring unparsable boolean; keeping the default",
			zap.String("variable", TrustLoopbackProxyEnv), zap.String("value", raw))
		return true
	}
	return enabled
}

func (a *App) Listen() error {
	if a.Router == nil {
		return fmt.Errorf("router is not initialized")
	}

	serverAddr := config.ServerAddress()
	logger.Info("starting HTTP server", zap.String("address", serverAddr))

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Router.Listen(serverAddr)
	}()

	return <-errCh
}

func (a *App) Shutdown() {
	logger.Info("shutting down application")
	if a.cancel != nil {
		a.cancel()
	}
	if a.Router != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Router.ShutdownWithContext(ctx); err != nil {
			logger.Error("failed to shutdown HTTP server", zap.Error(err))
		}
	}
	if a.cleanupDone != nil {
		<-a.cleanupDone
	}
	if a.Database != nil {
		if err := a.Database.Close(); err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}
}

func (a *App) initializeDatabase() error {
	dbConfig := database.LoadConfigFromApp(a.Config)
	if dbConfig.Type == "" {
		logger.Info("database type is not configured; skipping database initialization")
		return nil
	}

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create database instance: %w", err)
	}

	if err := db.Init(); err != nil {
		db.Close()
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	a.Database = db
	return nil
}

func (a *App) initializeZeroTierClient() error {
	if a.Config == nil || !a.Config.Initialized {
		logger.Info("system is not initialized; skipping automatic ZeroTier client initialization")
		return nil
	}

	ztClient, err := zerotier.NewClientWithConfig(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create ZeroTier client: %w", err)
	}

	if _, err := ztClient.GetStatus(); err != nil {
		return fmt.Errorf("ZeroTier client validation failed: %w", err)
	}

	a.ZTClient = ztClient
	return nil
}
