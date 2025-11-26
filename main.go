package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/initializer"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// GlobalDB is the global database instance, initialized after user completes setup wizard
var GlobalDB database.DBInterface

// GlobalZTClient is the global ZeroTier client instance, maintained after initialization
var GlobalZTClient *zerotier.Client

// GlobalRouter is the global router instance for route reloading
var GlobalRouter *gin.Engine

// ReloadRoutes reloads all API routes with current configuration
func ReloadRoutes() {
	logger.Info("开始重新加载路由")
	// TODO: Implement route reloading functionality using server initializer
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
		logger.Fatal("应用初始化失败", zap.Error(err))
	}

	// Set global variables for backward compatibility
	GlobalDB = appContext.Database
	GlobalZTClient = appContext.ZTClient
	GlobalRouter = appContext.Router

	// Initialize and start HTTP server
	serverInitializer := initializer.NewServerInitializer()
	if _, err := serverInitializer.Initialize(appContext.Database, appContext.ZTClient, appContext.JWTSecret); err != nil {
		logger.Fatal("服务器初始化失败", zap.Error(err))
	}

	// Start the HTTP server
	if err := serverInitializer.StartServer(); err != nil {
		logger.Fatal("启动服务器失败", zap.Error(err))
	}
}
