package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/config"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/routes"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// GlobalDB is the global database instance, initialized after user completes setup wizard
var GlobalDB database.DBInterface

// GlobalZTClient is the global ZeroTier client instance, maintained after initialization
var GlobalZTClient *zerotier.Client

// GlobalRouter is the global router instance for route reloading
var GlobalRouter *gin.Engine
var routerMutex sync.RWMutex

// ReloadRoutes reloads all API routes with current configuration
func ReloadRoutes() {
	logger.Info("开始重新加载路由")

	routerMutex.Lock()
	defer routerMutex.Unlock()

	// Clear existing routes by creating a new router instance
	GlobalRouter = gin.New()
	logger.Info("已创建新的路由器实例")

	// Add debug logging for diagnostics
	logger.Info("准备重新注册路由")

	// Re-register routes with global ZeroTier client
	cfg, _ := config.LoadConfig()
	if cfg != nil {
		logger.Info("使用配置文件中的JWT密钥重新注册路由")
		routes.SetupRoutesWithReload(GlobalRouter, GlobalZTClient, cfg.Security.JWTSecret, GlobalDB, ReloadRoutes)
	} else {
		// Use default configuration if config loading fails
		logger.Info("使用默认JWT密钥重新注册路由")
		routes.SetupRoutesWithReload(GlobalRouter, GlobalZTClient, "default-secret-key", GlobalDB, ReloadRoutes)
	}

	logger.Info("路由重新加载完成")
}

// main is the application entry point
func main() {
	fmt.Println("Tairitsu - ZeroTier Controller UI")

	// Initialize logger with info level
	logger.InitLogger("info")

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("加载配置失败", zap.Error(err))
	}

	// Load database configuration
	dbConfig := database.LoadConfig()

	// Create database instance only if database type is configured
	// Otherwise, GlobalDB remains nil until user configures via setup wizard
	if dbConfig.Type != "" {
		// Create database connection
		GlobalDB, err = database.NewDatabase(dbConfig)
		if err != nil {
			logger.Error("创建数据库实例失败", zap.Error(err))
			// Don't terminate, allow configuration through setup wizard
		} else {
			// Initialize database schema and tables
			if err := GlobalDB.Init(); err != nil {
				logger.Error("初始化数据库失败", zap.Error(err))
				// Don't terminate, allow reconfiguration through setup wizard
				GlobalDB = nil
			} else {
				logger.Info("数据库初始化成功", zap.String("type", string(dbConfig.Type)))
			}
		}
	} else {
		logger.Info("未配置数据库类型，等待用户通过设置向导配置")
	}

	// Set Gin mode based on environment
	if os.Getenv("NODE_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create initial router instance
	GlobalRouter = gin.New()

	// Register routes with global database instance (may be nil) and route reload function
	// Using global GlobalZTClient, initially nil, will be set during InitZeroTierClient
	routes.SetupRoutesWithReload(GlobalRouter, GlobalZTClient, cfg.Security.JWTSecret, GlobalDB, ReloadRoutes)

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器启动在", zap.String("address", serverAddr))
	if err := GlobalRouter.Run(serverAddr); err != nil {
		logger.Fatal("启动服务器失败", zap.Error(err))
	}
}
