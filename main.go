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
	
	// 创建数据库实例
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		logger.Fatal("创建数据库实例失败", zap.Error(err))
	}

	// 初始化数据库
	if err := db.Init(); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
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

	// 注册路由
	routes.SetupRoutes(router, ztClient, cfg.Security.JWTSecret, db)

	// 启动服务器
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器启动在", zap.String("address", serverAddr))
	if err := router.Run(serverAddr); err != nil {
		logger.Fatal("启动服务器失败", zap.Error(err))
	}
}