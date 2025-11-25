package initializer

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/config"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// AppContext 应用上下文，包含所有全局依赖
type AppContext struct {
	Config    *config.Config
	Database  database.DBInterface
	ZTClient  *zerotier.Client
	Router    *gin.Engine
	JWTSecret string
}

// AppInitializer 应用初始化器
type AppInitializer struct {
	context *AppContext
}

// NewAppInitializer 创建新的应用初始化器
func NewAppInitializer() *AppInitializer {
	return &AppInitializer{
		context: &AppContext{},
	}
}

// Initialize 执行完整的应用初始化流程
func (ai *AppInitializer) Initialize() (*AppContext, error) {
	// 1. 初始化日志系统（首先初始化，其他步骤可能需要使用日志）
	if err := ai.initializeLogger(); err != nil {
		return nil, fmt.Errorf("初始化日志系统失败: %w", err)
	}

	logger.Info("开始应用初始化流程")

	// 2. 加载配置
	if err := ai.loadConfiguration(); err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 3. 初始化数据库
	if err := ai.initializeDatabase(); err != nil {
		logger.Warn("数据库初始化失败，将继续运行", zap.Error(err))
		// 不返回错误，允许用户通过设置向导配置数据库
	}

	// 4. 初始化ZeroTier客户端
	if err := ai.initializeZeroTierClient(); err != nil {
		logger.Warn("ZeroTier客户端初始化失败，将继续运行", zap.Error(err))
		// 不返回错误，允许用户通过设置向导配置ZeroTier
	}

	// 5. 初始化HTTP服务器
	if err := ai.initializeHTTPServer(); err != nil {
		return nil, fmt.Errorf("初始化HTTP服务器失败: %w", err)
	}

	logger.Info("应用初始化完成")
	return ai.context, nil
}

// initializeLogger 初始化日志系统
func (ai *AppInitializer) initializeLogger() error {
	logger.InitLogger("info")
	logger.Info("日志系统初始化完成")
	return nil
}

// loadConfiguration 加载应用配置
func (ai *AppInitializer) loadConfiguration() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	ai.context.Config = cfg
	ai.context.JWTSecret = cfg.Security.JWTSecret

	// 设置Gin模式
	if os.Getenv("NODE_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger.Info("配置加载完成")
	return nil
}

// initializeDatabase 初始化数据库
func (ai *AppInitializer) initializeDatabase() error {
	// 加载数据库配置
	dbConfig := database.LoadConfig()

	// 如果未配置数据库类型，跳过初始化
	if dbConfig.Type == "" {
		logger.Info("未配置数据库类型，跳过数据库初始化")
		return nil
	}

	// 创建数据库实例
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("创建数据库实例失败: %w", err)
	}

	// 初始化数据库架构
	if err := db.Init(); err != nil {
		db.Close()
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	ai.context.Database = db
	logger.Info("数据库初始化完成", zap.String("type", string(dbConfig.Type)))
	return nil
}

// initializeZeroTierClient 初始化ZeroTier客户端
func (ai *AppInitializer) initializeZeroTierClient() error {
	// 只有在系统已初始化的情况下才自动初始化ZeroTier客户端
	if !ai.context.Config.Initialized {
		logger.Info("系统未初始化，跳过ZeroTier客户端自动初始化")
		return nil
	}

	logger.Info("系统已初始化，开始初始化ZeroTier客户端")

	// 动态创建ZeroTier客户端
	ztClient, err := zerotier.NewClient()
	if err != nil {
		return fmt.Errorf("创建ZeroTier客户端失败: %w", err)
	}

	// 验证客户端是否正常工作
	_, err = ztClient.GetStatus()
	if err != nil {
		return fmt.Errorf("ZeroTier客户端验证失败: %w", err)
	}

	ai.context.ZTClient = ztClient
	logger.Info("ZeroTier客户端初始化完成")
	return nil
}

// initializeHTTPServer 初始化HTTP服务器
func (ai *AppInitializer) initializeHTTPServer() error {
	// 创建路由器实例
	router := gin.New()
	ai.context.Router = router

	logger.Info("HTTP服务器初始化完成")
	return nil
}

// GetContext 获取应用上下文
func (ai *AppInitializer) GetContext() *AppContext {
	return ai.context
}
