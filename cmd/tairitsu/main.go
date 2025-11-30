package main

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/initializer"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GlobalDB is the global database instance, initialized after user completes setup wizard
var GlobalDB database.DBInterface

// GlobalZTClient is the global ZeroTier client instance, maintained after initialization
var GlobalZTClient *zerotier.Client

// GlobalRouter is the global router instance for route reloading
var GlobalRouter *gin.Engine

// GlobalServerInitializer is the global server initializer instance
var GlobalServerInitializer *initializer.ServerInitializer

// ReloadRoutes reloads all API routes with current configuration
func ReloadRoutes() {
	logger.Info("Starting to reload routes")
	if GlobalServerInitializer != nil {
		GlobalServerInitializer.ReloadRoutes()
		// Update global router reference after reloading
		GlobalRouter = GlobalServerInitializer.GetRouter()
		logger.Info("Route reload completed")
	} else {
		logger.Error("Server initializer not initialized, cannot reload routes")
	}
	// This method is kept for backward compatibility with global variables
}

// main is the application entry point
func main() {
	fmt.Println("Tairitsu - ZeroTier Controller Interface")

	// Create application initializer
	appInitializer := initializer.NewAppInitializer()

	// Execute complete application initialization process
	appContext, err := appInitializer.Initialize()
	if err != nil {
		logger.Fatal("Application initialization failed", zap.Error(err))
	}

	// Set global variables for backward compatibility
	GlobalDB = appContext.Database
	GlobalZTClient = appContext.ZTClient
	GlobalRouter = appContext.Router

	// Initialize and start HTTP server
	GlobalServerInitializer = initializer.NewServerInitializer()
	if _, err := GlobalServerInitializer.Initialize(appContext.Database, appContext.ZTClient, appContext.JWTSecret); err != nil {
		logger.Fatal("Server initialization failed", zap.Error(err))
	}

	// Start the HTTP server
	if err := GlobalServerInitializer.StartServer(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
