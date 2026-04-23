package bootstrap

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/routes"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type App struct {
	Config       *config.Config
	Database     database.DBInterface
	ZTClient     *zerotier.Client
	Dependencies *assembly.Dependencies
	Router       *fiber.App
}

func Build() (*App, error) {
	logger.InitLogger("info")
	logger.Info("开始应用启动装配")

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	app := &App{Config: cfg}

	if err := app.initializeDatabase(); err != nil {
		if cfg.Initialized {
			return nil, fmt.Errorf("系统已初始化，但数据库初始化失败: %w", err)
		}
		logger.Warn("数据库初始化失败，将以未初始化模式继续运行", zap.Error(err))
	}

	if err := app.initializeZeroTierClient(); err != nil {
		if cfg.Initialized {
			return nil, fmt.Errorf("系统已初始化，但ZeroTier客户端初始化失败: %w", err)
		}
		logger.Warn("ZeroTier客户端初始化失败，将以未初始化模式继续运行", zap.Error(err))
	}

	app.Dependencies = assembly.NewDependencies(app.Config, app.Database, app.ZTClient)
	app.Router = fiber.New()
	routes.SetupRoutes(app.Router, app.Dependencies)

	logger.Info("应用启动装配完成")
	return app, nil
}

func (a *App) Listen() error {
	if a.Router == nil {
		return fmt.Errorf("路由器未初始化")
	}

	serverAddr := config.ServerAddressFrom(a.Config)
	logger.Info("启动HTTP服务器", zap.String("address", serverAddr))

	if err := a.Router.Listen(serverAddr); err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	return nil
}

func (a *App) initializeDatabase() error {
	dbConfig := database.LoadConfigFromApp(a.Config)
	if dbConfig.Type == "" {
		logger.Info("未配置数据库类型，跳过数据库初始化")
		return nil
	}

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("创建数据库实例失败: %w", err)
	}

	if err := db.Init(); err != nil {
		db.Close()
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	a.Database = db
	return nil
}

func (a *App) initializeZeroTierClient() error {
	if a.Config == nil || !a.Config.Initialized {
		logger.Info("系统未初始化，跳过ZeroTier客户端自动初始化")
		return nil
	}

	ztClient, err := zerotier.NewClientWithConfig(a.Config)
	if err != nil {
		return fmt.Errorf("创建ZeroTier客户端失败: %w", err)
	}

	if _, err := ztClient.GetStatus(); err != nil {
		return fmt.Errorf("ZeroTier客户端验证失败: %w", err)
	}

	a.ZTClient = ztClient
	return nil
}
