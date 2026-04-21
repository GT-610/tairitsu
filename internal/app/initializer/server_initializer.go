package initializer

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ServerInitializer 服务器初始化器
type ServerInitializer struct {
	router    *fiber.App
	jwtSecret string
	db        database.DBInterface
	ztClient  *zerotier.Client
}

// NewServerInitializer 创建新的服务器初始化器
func NewServerInitializer() *ServerInitializer {
	return &ServerInitializer{}
}

// Initialize 初始化HTTP服务器
func (si *ServerInitializer) Initialize(db database.DBInterface, ztClient *zerotier.Client, jwtSecret string) (*fiber.App, error) {
	si.db = db
	si.ztClient = ztClient
	si.jwtSecret = jwtSecret

	// 创建路由器实例
	router := fiber.New()
	si.router = router

	// 注册路由
	if err := si.setupRoutes(); err != nil {
		return nil, fmt.Errorf("设置路由失败: %w", err)
	}

	logger.Info("HTTP服务器初始化完成")
	return router, nil
}

// setupRoutes 设置应用路由
func (si *ServerInitializer) setupRoutes() error {
	routes.SetupRoutes(si.router, si.ztClient, si.jwtSecret, si.db)
	return nil
}

// GetRouter 获取路由器实例
func (si *ServerInitializer) GetRouter() *fiber.App {
	return si.router
}

// StartServer 启动HTTP服务器
func (si *ServerInitializer) StartServer() error {
	if si.router == nil {
		return fmt.Errorf("路由器未初始化")
	}

	cfg := config.AppConfig
	if cfg == nil {
		return fmt.Errorf("配置未加载")
	}

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("启动HTTP服务器", zap.String("address", serverAddr))

	if err := si.router.Listen(serverAddr); err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	return nil
}
