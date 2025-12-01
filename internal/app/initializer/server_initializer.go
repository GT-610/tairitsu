package initializer

import (
	"fmt"
	"sync"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServerInitializer Server initializer
type ServerInitializer struct {
	router      *gin.Engine
	jwtSecret   string
	db          database.Database
	ztClient    *zerotier.Client
	config      *config.Config
	reloadFunc  func()
	routerMutex sync.RWMutex
}

// NewServerInitializer Create a new server initializer
func NewServerInitializer() *ServerInitializer {
	return &ServerInitializer{}
}

// Initialize Initialize HTTP server
func (si *ServerInitializer) Initialize(db database.Database, ztClient *zerotier.Client, jwtSecret string) (*gin.Engine, error) {
	si.db = db
	si.ztClient = ztClient
	si.jwtSecret = jwtSecret
	si.config = config.AppConfig

	// Create router instance
	router := gin.New()
	si.router = router

	// Set route reload function
	si.reloadFunc = si.ReloadRoutes

	// Register routes
	if err := si.setupRoutes(); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	logger.Info("HTTP server initialization completed")
	return router, nil
}

// setupRoutes Setup application routes
func (si *ServerInitializer) setupRoutes() error {
	// Register routes using route reload function
	routes.SetupRoutesWithReload(si.router, si.ztClient, si.jwtSecret, si.db, si.reloadFunc)
	return nil
}

// ReloadRoutes Reload all API routes
func (si *ServerInitializer) ReloadRoutes() {
	logger.Info("Starting route reload")

	si.routerMutex.Lock()
	defer si.routerMutex.Unlock()

	// Clear existing routes, create new router instance
	si.router = gin.New()

	// Re-register routes
	if err := si.setupRoutes(); err != nil {
		logger.Error("Failed to re-register routes", zap.Error(err))
		return
	}

	logger.Info("Route reload completed")
}

// GetRouter Get router instance
func (si *ServerInitializer) GetRouter() *gin.Engine {
	si.routerMutex.RLock()
	defer si.routerMutex.RUnlock()
	return si.router
}

// StartServer Start HTTP server
func (si *ServerInitializer) StartServer() error {
	if si.router == nil {
		return fmt.Errorf("router not initialized")
	}

	if si.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	serverAddr := fmt.Sprintf(":%d", si.config.Server.Port)
	logger.Info("Starting HTTP server", zap.String("address", serverAddr))

	if err := si.router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
