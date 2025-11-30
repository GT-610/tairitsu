package initializer

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// AppContext Application context containing all global dependencies
type AppContext struct {
	Config    *config.Config
	Database  database.DBInterface
	ZTClient  *zerotier.Client
	Router    *gin.Engine
	JWTSecret string
}

// AppInitializer Application initializer
type AppInitializer struct {
	context *AppContext
}

// NewAppInitializer Create a new application initializer
func NewAppInitializer() *AppInitializer {
	return &AppInitializer{
		context: &AppContext{},
	}
}

// Initialize Execute complete application initialization process
func (ai *AppInitializer) Initialize() (*AppContext, error) {
	// 1. Initialize logging system (initialize first, other steps may need logging)
	if err := ai.initializeLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logging system: %w", err)
	}

	logger.Info("Starting application initialization process")

	// 2. Load configuration
	if err := ai.loadConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 3. Initialize database
	if err := ai.initializeDatabase(); err != nil {
		logger.Warn("Database initialization failed, continuing to run", zap.Error(err))
		// Don't return error, allow user to configure database through setup wizard
	}

	// 4. Initialize ZeroTier client
	if err := ai.initializeZeroTierClient(); err != nil {
		logger.Warn("ZeroTier client initialization failed, continuing to run", zap.Error(err))
		// Don't return error, allow user to configure ZeroTier through setup wizard
	}

	// 5. Initialize HTTP server
	if err := ai.initializeHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	logger.Info("Application initialization completed")
	return ai.context, nil
}

// initializeLogger Initialize logging system
func (ai *AppInitializer) initializeLogger() error {
	logger.InitLogger("info")
	return nil
}

// loadConfiguration Load application configuration
func (ai *AppInitializer) loadConfiguration() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	ai.context.Config = cfg
	ai.context.JWTSecret = cfg.Security.JWTSecret

	// Set Gin mode
	if os.Getenv("NODE_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

// initializeDatabase Initialize database
func (ai *AppInitializer) initializeDatabase() error {
	// Load database configuration
	dbConfig := database.LoadConfig()

	// If database type is not configured, skip initialization
	if dbConfig.Type == "" {
		logger.Info("Database type not configured, skipping database initialization")
		return nil
	}

	// Create database instance
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create database instance: %w", err)
	}

	// Initialize database schema
	if err := db.Init(); err != nil {
		db.Close()
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	ai.context.Database = db
	return nil
}

// initializeZeroTierClient Initialize ZeroTier client
func (ai *AppInitializer) initializeZeroTierClient() error {
	// Only automatically initialize ZeroTier client if system is already initialized
	if !ai.context.Config.Initialized {
		logger.Info("System not initialized, skipping ZeroTier client automatic initialization")
		return nil
	}

	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create ZeroTier client: %w", err)
	}

	// Verify client is working properly
	_, err = ztClient.GetStatus()
	if err != nil {
		return fmt.Errorf("ZeroTier client verification failed: %w", err)
	}

	ai.context.ZTClient = ztClient
	return nil
}

// initializeHTTPServer Initialize HTTP server
func (ai *AppInitializer) initializeHTTPServer() error {
	// Create router instance
	router := gin.New()
	ai.context.Router = router

	return nil
}

// GetContext Get application context
func (ai *AppInitializer) GetContext() *AppContext {
	return ai.context
}
