package bootstrap

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/GT-610/tempest-zt/internal/app/assembly"
	"github.com/GT-610/tempest-zt/internal/app/config"
	"github.com/GT-610/tempest-zt/internal/app/database"
	"github.com/GT-610/tempest-zt/internal/app/logger"
	"github.com/GT-610/tempest-zt/internal/app/routes"
	"github.com/GT-610/tempest-zt/internal/app/services"
	"github.com/GT-610/tempest-zt/internal/zerotier"
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
	warnAboutProxyKeying()
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
	TrustedProxiesEnv = "TEMPEST_ZT_TRUSTED_PROXIES"
	// ProxyHeaderEnv overrides the header carrying the client address.
	ProxyHeaderEnv = "TEMPEST_ZT_PROXY_HEADER"
	// TrustLoopbackProxyEnv turns off the default trust for a proxy on loopback.
	TrustLoopbackProxyEnv = "TEMPEST_ZT_TRUSTED_PROXY_LOOPBACK"
)

// defaultProxyHeader is the header reverse proxies use to carry the client
// address. X-Forwarded-For is set by nginx, Caddy, Traefik and HAProxy alike;
// the previously hardcoded X-Real-IP only exists in nginx.
const defaultProxyHeader = "X-Forwarded-For"

// announceSetupToken makes the first-run token impossible to miss. Only the
// banner carries the value: standard output is what `docker logs` shows, while a
// secret written to the log sink as well would be collected, retained and
// rotated on a different schedule from the token it describes.
func announceSetupToken(token *services.SetupToken) {
	if token == nil || !token.Issued() {
		return
	}

	fmt.Fprintf(os.Stdout, "\n  Tempest-ZT is not initialized yet.\n  Setup token: %s\n  Enter it in the setup wizard. A restart issues a new token.\n\n", token.Value())
	logger.Warn("first-run setup token issued; the setup endpoints require it, and the token is printed on standard output only")
}

func newHTTPApp() *fiber.App {
	return fiber.New(httpAppConfig())
}

// Request timeouts. Fiber leaves all three at zero by default, which means a
// client that sends a request line and then stalls holds the connection open
// indefinitely; enough of those exhaust the listener without ever completing a
// request. The values are generous relative to what this API does — every
// handler answers in milliseconds, and the slowest, a bulk network import, is
// bounded by the number of networks the controller reports — so they bound a
// stalled peer without cutting off real work.
const (
	// readTimeout covers reading the whole request, headers and body.
	readTimeout = 30 * time.Second
	// writeTimeout covers writing the response.
	writeTimeout = 60 * time.Second
	// idleTimeout covers waiting for the next request on a kept-alive
	// connection.
	idleTimeout = 60 * time.Second
)

func httpAppConfig() fiber.Config {
	return fiber.Config{
		TrustProxy:  true,
		ProxyHeader: proxyHeader(),
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies:  trustedProxies(),
			Loopback: trustLoopbackProxy(),
		},
		EnableIPValidation: true,
		ReadTimeout:        readTimeout,
		WriteTimeout:       writeTimeout,
		IdleTimeout:        idleTimeout,
	}
}

// proxyHeader returns the header a trusted proxy is expected to set.
func proxyHeader() string {
	if header := config.EnvWithLegacy(ProxyHeaderEnv); header != "" {
		return header
	}
	return defaultProxyHeader
}

// trustedProxies returns the configured proxy addresses and CIDR ranges,
// dropping entries that cannot be parsed so that a typo degrades to "not
// trusted" with a warning instead of being silently accepted.
func trustedProxies() []string {
	configured := config.EnvListWithLegacy(TrustedProxiesEnv)
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
	raw := config.EnvWithLegacy(TrustLoopbackProxyEnv)
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

// warnAboutProxyKeying explains the one consequence of the trust configuration
// that an operator cannot see from the outside.
//
// The rate limiters key on the client address. With no trusted proxies
// configured, a deployment behind a reverse proxy records the proxy's address for
// every caller, so they all share one budget: one noisy client can exhaust it and
// lock everyone else out. The reverse case is worse and is what the
// documentation's proxy list prevents — a proxy that forwards a caller-supplied
// X-Forwarded-For lets the caller choose the key, which retires the limit
// entirely. Neither state is detectable from a healthy request stream, so it is
// said once at startup.
func warnAboutProxyKeying() {
	if len(trustedProxies()) > 0 {
		return
	}

	logger.Warn("no trusted proxies are configured, so per-client rate limiting and the address recorded on a session key on the direct peer; if Tempest-ZT runs behind a reverse proxy, set "+TrustedProxiesEnv+" to that proxy's address",
		zap.Bool("loopback_proxy_trusted", trustLoopbackProxy()))
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
