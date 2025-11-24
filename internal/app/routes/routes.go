package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/config"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/handlers"
	"github.com/tairitsu/tairitsu/internal/app/middleware"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, ztClient *zerotier.Client, jwtSecret string, db database.DBInterface) {
	authService := services.NewAuthService(jwtSecret)
	// 创建重新加载函数 - 使用简单实现避免循环引用
	reloadFunc := func() { SetupRoutes(router, ztClient, jwtSecret, db) }
	SetupRoutesWithReload(router, ztClient, authService, db, reloadFunc)
}

// 简单的认证中间件函数
func authMiddleware(authService *services.AuthService, c *gin.Context, isSetupWizard bool) {
	// 从请求头获取令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(401, gin.H{"error": "未提供认证令牌"})
		c.Abort()
		return
	}

	// 提取令牌部分
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(401, gin.H{"error": "认证令牌格式无效"})
		c.Abort()
		return
	}

	tokenString := authHeader[len(bearerPrefix):]

	// 验证令牌
	claims, err := authService.ValidateToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{"error": "认证令牌无效: " + err.Error()})
		c.Abort()
		return
	}

	// 将用户信息存储到上下文中
	c.Set("user", claims)
	c.Next()
}

// SetupRoutesWithReload 创建并配置路由，支持路由重载
func SetupRoutesWithReload(router *gin.Engine, ztClient *zerotier.Client, authService *services.AuthService, db database.DBInterface, reloadRoutesFunc func()) {
	// 应用中间件
	router.Use(middleware.CORS())

	// 创建服务实例
	networkService := services.NewNetworkService(ztClient)
	userService := services.NewUserService()

	// 创建系统处理器，提供所有必要参数
	systemHandler := handlers.NewSystemHandler(networkService, userService, reloadRoutesFunc, authService)

	// 创建认证中间件，使用新的中间件创建方法
	setupWizardMiddleware := func(c *gin.Context) {
		authMiddleware(authService, c, true)
	}
	normalAuthMiddleware := func(c *gin.Context) {
		authMiddleware(authService, c, false)
	}

	// 获取系统初始化状态
	isInitialized := config.IsInitialized()

	// API 路由组
	api := router.Group("/api")
	{
		// 始终可用的路由（无需认证）
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// 系统状态检测（无需认证）
		api.GET("/system/status", systemHandler.GetSystemStatus)

		// 设置向导专用路由（仅在未初始化时可用）
		if !isInitialized {
			// 初始化设置向导专用路由组
			setup := api.Group("/setup")
			{
				// 获取设置向导临时令牌
				setup.POST("/token", func(c *gin.Context) {
					token, err := authService.GenerateTempToken()
					if err != nil {
						c.JSON(500, gin.H{"error": "生成临时令牌失败"})
						return
					}
					c.JSON(200, gin.H{"token": token})
				})
			}

			// 初始化相关路由（使用设置向导认证）
			systemSetup := api.Group("/system")
			systemSetup.Use(setupWizardMiddleware)
			{
				// 数据库配置
				systemSetup.POST("/database", systemHandler.ConfigureDatabase)

				// 初始化管理员账户创建步骤
				systemSetup.POST("/admin/init", systemHandler.InitializeAdminCreation)

				// 设置系统初始化状态（完成初始化）
				systemSetup.POST("/initialized", systemHandler.SetInitialized)

				// 完成设置向导
				systemSetup.POST("/setup/complete", systemHandler.CompleteSetupWizard)
			}
		}

		// 正常使用的路由（仅在已初始化时可用）
		if isInitialized {
			// 需要认证的路由
			authenticated := api.Group("/")
			authenticated.Use(normalAuthMiddleware)
			{
				// 系统相关路由 - 使用已存在的方法
				authenticated.GET("/system/status", systemHandler.GetSystemStatus)
			}
		}
	}
}
