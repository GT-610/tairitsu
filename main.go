package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/config"
	"github.com/tairitsu/tairitsu/internal/app/routes"
	"github.com/tairitsu/tairitsu/internal/zerotier"
)

func main() {
	fmt.Println("Tairitsu - ZeroTier Controller UI")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建ZeroTier客户端
	ztToken, err := config.GetZTToken(cfg.ZeroTier.TokenPath)
	if err != nil {
		log.Fatalf("获取ZeroTier Token失败: %v", err)
	}

	ztClient := zerotier.NewClient(cfg.ZeroTier.URL, ztToken)

	// 设置Gin模式
	if os.Getenv("NODE_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 注册路由
	routes.SetupRoutes(router, ztClient, cfg.Security.JWTSecret)

	// 启动服务器
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("服务器启动在 %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}