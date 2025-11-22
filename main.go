package main

import (
	"fmt"
	"os"

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
	dbConfig := database.LoadConfigFromEnv()
	
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

	// 创建ZeroTier客户端
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Fatal("创建ZeroTier客户端失败", zap.Error(err))
	}

	// 设置Gin模式
	if os.Getenv("NODE_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 注册路由，传递全局数据库实例（可能为nil）
	routes.SetupRoutes(router, ztClient, cfg.Security.JWTSecret, GlobalDB)

	// 启动服务器
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器启动在", zap.String("address", serverAddr))
	if err := router.Run(serverAddr); err != nil {
		logger.Fatal("启动服务器失败", zap.Error(err))
	}
}