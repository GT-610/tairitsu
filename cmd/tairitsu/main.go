package main

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/initializer"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

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
