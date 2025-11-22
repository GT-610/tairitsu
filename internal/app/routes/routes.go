package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/handlers"
	"github.com/tairitsu/tairitsu/internal/app/middleware"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, ztClient *zerotier.Client, jwtSecret string, db database.DBInterface) {
	// 应用中间件
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	// 创建服务实例
	networkService := services.NewNetworkService(ztClient)
	
	// 创建用户服务实例，可能使用nil数据库
	var userService *services.UserService
	if db != nil {
		userService = services.NewUserServiceWithDB(db) // 使用传入的数据库实例
	} else {
		userService = services.NewUserServiceWithoutDB() // 创建不使用数据库的服务实例
	}
	
	jwtService := services.NewJWTService(jwtSecret)

	// 创建处理器实例
	networkHandler := handlers.NewNetworkHandler(networkService)
	memberHandler := handlers.NewMemberHandler(networkService)
	authHandler := handlers.NewAuthHandler(userService, jwtService)
	systemHandler := handlers.NewSystemHandler(networkService, userService)

	// 创建认证中间件
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// API 路由组
	api := router.Group("/api")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// 系统状态检测（无需认证）
		api.GET("/system/status", systemHandler.GetSystemStatus)
		
		// 数据库配置（无需认证，仅在初始设置时可用）
		api.POST("/system/database", systemHandler.ConfigureDatabase)

		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			// 只有在数据库已配置时才启用注册和登录功能
			if db != nil {
				auth.POST("/register", authHandler.Register) // 用户注册
				auth.POST("/login", authHandler.Login)       // 用户登录
			}
		}

		// 需要认证的路由
		authenticated := api.Group("/")
		authenticated.Use(authMiddleware)
		{
			// 只有在数据库已配置时才启用需要数据库的功能
			if db != nil {
				// 用户信息
				authenticated.GET("/profile", authHandler.GetProfile) // 获取当前用户信息

				// ZeroTier 状态
				authenticated.GET("/status", networkHandler.GetStatus)

				// 网络管理
				networks := authenticated.Group("/networks")
				{
					networks.GET("", networkHandler.GetNetworks)           // 获取所有网络
					networks.POST("", networkHandler.CreateNetwork)         // 创建网络
					networks.GET("/:id", networkHandler.GetNetwork)         // 获取单个网络
					networks.PUT("/:id", networkHandler.UpdateNetwork)      // 更新网络
					networks.DELETE("/:id", networkHandler.DeleteNetwork)   // 删除网络

					// 成员管理（嵌套在网络路由中）
					networks.GET("/:id/members", memberHandler.GetMembers)         // 获取成员列表
					networks.GET("/:id/members/:memberId", memberHandler.GetMember) // 获取单个成员
					networks.PUT("/:id/members/:memberId", memberHandler.UpdateMember) // 更新成员
					networks.DELETE("/:id/members/:memberId", memberHandler.DeleteMember) // 删除成员
				}
			}
		}
	}
}