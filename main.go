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

// GlobalDB 全局数据库实例，在用户完成设置向导后初始化
var GlobalDB database.DBInterface

// GlobalZTClient 全局ZeroTier客户端实例，在初始化后保持
var GlobalZTClient *zerotier.Client

// GlobalRouter 全局路由器实例，用于重新加载路由
var GlobalRouter *gin.Engine
var routerMutex sync.RWMutex

// ReloadRoutes 重新加载路由
func ReloadRoutes() {
	logger.Info("开始重新加载路由")
	
	routerMutex.Lock()
	defer routerMutex.Unlock()

	// 清除现有路由
	GlobalRouter = gin.New()
	logger.Info("已创建新的路由器实例")

	// 重新加载数据库配置
	dbConfig := database.LoadConfig()
	logger.Info("已加载数据库配置", zap.Any("config", dbConfig))
	logger.Info("数据库配置详情", zap.String("type", string(dbConfig.Type)), zap.String("path", dbConfig.Path))

	// 如果数据库配置有效，则创建数据库实例
	if dbConfig.Type != "" {
		db, err := database.NewDatabase(dbConfig)
		if err != nil {
			logger.Error("重新加载路由时创建数据库实例失败", zap.Error(err))
			GlobalDB = nil
		} else {
			if err := db.Init(); err != nil {
				logger.Error("重新加载路由时数据库初始化失败", zap.Error(err))
				GlobalDB = nil
			} else {
				GlobalDB = db
				logger.Info("重新加载路由时数据库初始化成功", zap.String("type", string(dbConfig.Type)))
				logger.Info("数据库实例已设置，认证路由应可用")
			}
		}
	} else {
		GlobalDB = nil
		logger.Info("数据库配置为空，跳过数据库初始化")
	}

	// 添加更多调试日志以诊断问题
	logger.Info("准备重新注册路由")

	// 重新注册路由，使用全局的ZeroTier客户端
	cfg, _ := config.LoadConfig()
	if cfg != nil {
		logger.Info("使用配置文件中的JWT密钥重新注册路由")
		routes.SetupRoutesWithReload(GlobalRouter, GlobalZTClient, cfg.Security.JWTSecret, GlobalDB, ReloadRoutes)
	} else {
		// 使用默认配置
		logger.Info("使用默认JWT密钥重新注册路由")
		routes.SetupRoutesWithReload(GlobalRouter, GlobalZTClient, "default-secret-key", GlobalDB, ReloadRoutes)
	}

	logger.Info("路由重新加载完成")
}

func main() {
	fmt.Println("Tairitsu - ZeroTier Controller UI")

	// 初始化日志
	logger.InitLogger("info")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("加载配置失败", zap.Error(err))
	}

	// 加载数据库配置
	dbConfig := database.LoadConfig()
	
	// 只有在配置了数据库类型时才创建数据库实例
	// 否则，GlobalDB将保持为nil，直到用户通过设置向导配置数据库
	if dbConfig.Type != "" {
		// 创建数据库实例
		GlobalDB, err = database.NewDatabase(dbConfig)
		if err != nil {
			logger.Error("创建数据库实例失败", zap.Error(err))
			// 不终止程序，允许通过设置向导配置数据库
		} else {
			// 初始化数据库
			if err := GlobalDB.Init(); err != nil {
				logger.Error("初始化数据库失败", zap.Error(err))
				// 不终止程序，允许通过设置向导重新配置数据库
				GlobalDB = nil
			} else {
				logger.Info("数据库初始化成功", zap.String("type", string(dbConfig.Type)))
			}
		}
	} else {
		logger.Info("未配置数据库类型，等待用户通过设置向导配置")
	}

	// 设置Gin模式
	if os.Getenv("NODE_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	GlobalRouter = gin.New()

	// 注册路由，传递全局数据库实例（可能为nil）和重新加载路由的函数
	// 使用全局的GlobalZTClient，初始为nil，将在InitZeroTierClient时设置
	routes.SetupRoutesWithReload(GlobalRouter, GlobalZTClient, cfg.Security.JWTSecret, GlobalDB, ReloadRoutes)

	// 启动服务器
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器启动在", zap.String("address", serverAddr))
	if err := GlobalRouter.Run(serverAddr); err != nil {
		logger.Fatal("启动服务器失败", zap.Error(err))
	}
}