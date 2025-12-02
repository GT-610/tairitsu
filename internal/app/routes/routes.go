package routes

import (
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures application routes
func SetupRoutes(router *gin.Engine, ztClient *zerotier.Client, jwtSecret string, db database.DBInterface) {
	SetupRoutesWithReload(router, ztClient, jwtSecret, db, nil)
}

// SetupRoutesWithReload configures routes with reload capability
func SetupRoutesWithReload(router *gin.Engine, ztClient *zerotier.Client, jwtSecret string, db database.DBInterface, reloadRoutesFunc func()) {
	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())
	router.Use(middleware.ErrorHandler())

	// Root path handler for HTML browsers
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<html>
<head>
	<title>Tairitsu Backend</title>
	<style>
		body {
			margin: 20px;
			color: #333;
		}
		h1 {
			color: #4a6fa5;
		}
	</style>
</head>
<body>
	<h1>Tairitsu</h1>
	<p>This is Tairitsu backend service.</p>
	<p>If you can see this page, it means the backend is running.</p>
</body>
</html>`)
	})

	// Create service instances
	networkService := services.NewNetworkService(ztClient, db)

	// Create user service instance, may use nil database
	var userService *services.UserService
	if db != nil {
		userService = services.NewUserServiceWithDB(db) // Use provided database instance
	} else {
		userService = services.NewUserServiceWithoutDB() // Create service instance without database
	}

	jwtService := services.NewJWTService(jwtSecret)

	// Create handler instances
	networkHandler := handlers.NewNetworkHandler(networkService)
	memberHandler := handlers.NewMemberHandler(networkService)
	authHandler := handlers.NewAuthHandler(userService, jwtService)
	userHandler := handlers.NewUserHandler(userService)

	// Use empty function if no reload function provided
	if reloadRoutesFunc == nil {
		reloadRoutesFunc = func() {}
	}

	systemHandler := handlers.NewSystemHandler(networkService, userService, reloadRoutesFunc)

	// Create authentication middleware
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// API routes group
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// System status check (no authentication required)
		api.GET("/system/status", systemHandler.GetSystemStatus)

		// Database configuration (no authentication required, available during initial setup only)
		api.POST("/system/database", systemHandler.ConfigureDatabase)

		// ZeroTier connection test (no authentication required, available during initial setup only)
		api.GET("/system/zerotier/test", systemHandler.TestZeroTierConnection)

		// Save ZeroTier configuration (no authentication required, available during initial setup only)
		api.POST("/system/zerotier/config", systemHandler.SaveZeroTierConfig)

		// ZeroTier client initialization (no authentication required, available during initial setup only)
		api.POST("/system/zerotier/init", systemHandler.InitZeroTierClient)

		// Set system initialization status (no authentication required, available during initial setup only)
		api.POST("/system/initialized", systemHandler.SetInitialized)

		// Initialize admin account creation step (no authentication required, available during initial setup only)
		api.POST("/system/admin/init", systemHandler.InitializeAdminCreation)

		// Reload routes (no authentication required, available during initial setup only)
		api.POST("/system/reload", systemHandler.ReloadRoutes)

		// Authentication routes (no authentication required)
		auth := api.Group("/auth")
		{
			// Only enable registration and login if database is configured
			if db != nil {
				auth.POST("/register", authHandler.Register) // User registration
				auth.POST("/login", authHandler.Login)       // User login
			}
		}

		// Authenticated routes
		authenticated := api.Group("/")
		authenticated.Use(authMiddleware)
		{
			// Only enable database-dependent features if database is configured
			if db != nil {
				// User information
			authenticated.GET("/profile", authHandler.GetProfile) // Get current user info
			authenticated.POST("/auth/update-password", authHandler.ChangePassword) // Update user password (deprecated)
			authenticated.PUT("/profile/password", authHandler.ChangePassword) // Update user password

				// ZeroTier status
				authenticated.GET("/status", networkHandler.GetStatus)

				// Network management
				networks := authenticated.Group("/networks")
				{
					networks.GET("", networkHandler.GetNetworks)          // Get all networks
					networks.POST("", networkHandler.CreateNetwork)       // Create network
					networks.GET("/:id", networkHandler.GetNetwork)       // Get single network
					networks.PUT("/:id", networkHandler.UpdateNetwork)    // Update network
					networks.DELETE("/:id", networkHandler.DeleteNetwork) // Delete network

					// Member management (nested within network routes)
					networks.GET("/:id/members", memberHandler.GetMembers)                // Get member list
					networks.GET("/:id/members/:memberId", memberHandler.GetMember)       // Get single member
					networks.PUT("/:id/members/:memberId", memberHandler.UpdateMember)    // Update member
					networks.DELETE("/:id/members/:memberId", memberHandler.DeleteMember) // Delete member
				}
			}
		}

		// Admin-only routes
		admin := api.Group("/")
		admin.Use(authMiddleware)
		admin.Use(middleware.AdminRequired())
		{
			// System statistics
			admin.GET("/system/stats", systemHandler.GetSystemStats) // Get system resource statistics
			
			// User management
			admin.GET("/users", userHandler.GetAllUsers) // Get all users
			admin.PUT("/users/:userId/role", userHandler.UpdateUserRole) // Update user role
		}
	}
}
