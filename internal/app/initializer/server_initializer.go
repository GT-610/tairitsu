package initializer

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/config"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/routes"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// ServerInitializer 服务器初始化器
type ServerInitializer struct {
	router      *gin.Engine
	jwtSecret   string
	db          database.DBInterface
	ztClient    *zerotier.Client
	reloadFunc  func()
	routerMutex sync.RWMutex
}

// NewServerInitializer 创建新的服务器初始化器
func NewServerInitializer() *ServerInitializer {
	return &ServerInitializer{}
}

// Initialize 初始化HTTP服务器
func (si *ServerInitializer) Initialize(db database.DBInterface, ztClient *zerotier.Client, jwtSecret string) (*gin.Engine, error) {
	si.db = db
	si.ztClient = ztClient
	si.jwtSecret = jwtSecret

	// 创建路由器实例
	router := gin.New()
	si.router = router

	// 设置路由重新加载函数
	si.reloadFunc = si.ReloadRoutes

	// 注册路由
	if err := si.setupRoutes(); err != nil {
		return nil, fmt.Errorf("设置路由失败: %w", err)
	}

	logger.Info("HTTP服务器初始化完成")
	return router, nil
}

// setupRoutes 设置应用路由
func (si *ServerInitializer) setupRoutes() error {
	// 使用路由重新加载函数注册路由
	routes.SetupRoutesWithReload(si.router, si.ztClient, si.jwtSecret, si.db, si.reloadFunc)
	return nil
}

// ReloadRoutes 重新加载所有API路由
func (si *ServerInitializer) ReloadRoutes() {
	logger.Info("开始重新加载路由")

	si.routerMutex.Lock()
	defer si.routerMutex.Unlock()

	// 清除现有路由，创建新的路由器实例
	si.router = gin.New()

	// 重新注册路由
	if err := si.setupRoutes(); err != nil {
		logger.Error("重新注册路由失败", zap.Error(err))
		return
	}

	logger.Info("路由重新加载完成")
}

// GetRouter 获取路由器实例
func (si *ServerInitializer) GetRouter() *gin.Engine {
	si.routerMutex.RLock()
	defer si.routerMutex.RUnlock()
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

	if err := si.router.Run(serverAddr); err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	return nil
}
