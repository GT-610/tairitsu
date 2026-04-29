package bootstrap

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/routes"
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

	app.Dependencies = assembly.NewDependencies(app.Config, app.Database, app.ZTClient)
	app.Router = fiber.New()
	routes.SetupRoutes(app.Router, app.Dependencies)

	logger.Info("application assembly completed")
	return app, nil
}

func (a *App) Listen() error {
	if a.Router == nil {
		return fmt.Errorf("router is not initialized")
	}

	serverAddr := config.ServerAddressFrom(a.Config)
	logger.Info("starting HTTP server", zap.String("address", serverAddr))

	if err := a.Router.Listen(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
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
