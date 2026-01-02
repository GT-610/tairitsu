package routes

import (
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// SetupRoutes configures application routes
func SetupRoutes(router *fiber.App, ztClient *zerotier.Client, jwtSecret string, db database.DBInterface) {
	SetupRoutesWithReload(router, ztClient, jwtSecret, db, nil)
}

// SetupRoutesWithReload configures routes with reload capability
func SetupRoutesWithReload(router *fiber.App, ztClient *zerotier.Client, jwtSecret string, db database.DBInterface, reloadRoutesFunc func()) {
	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(cors.New())
	router.Use(middleware.RateLimit())
	router.Use(middleware.ErrorHandler())

	// Root path handler for HTML browsers
	router.Get("/", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(`<html>
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
		api.Get("/health", func(c fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
		})

		// System status check (no authentication required)
		api.Get("/system/status", systemHandler.GetSystemStatus)

		// Database configuration (no authentication required, available during initial setup only)
		api.Post("/system/database", systemHandler.ConfigureDatabase)

		// ZeroTier connection test (no authentication required, available during initial setup only)
		api.Get("/system/zerotier/test", systemHandler.TestZeroTierConnection)

		// Save ZeroTier configuration (no authentication required, available during initial setup only)
		api.Post("/system/zerotier/config", systemHandler.SaveZeroTierConfig)

		// ZeroTier client initialization (no authentication required, available during initial setup only)
		api.Post("/system/zerotier/init", systemHandler.InitZeroTierClient)

		// Set system initialization status (no authentication required, available during initial setup only)
		api.Post("/system/initialized", systemHandler.SetInitialized)

		// Initialize admin account creation step (no authentication required, available during initial setup only)
		api.Post("/system/admin/init", systemHandler.InitializeAdminCreation)

		// Reload routes (no authentication required, available during initial setup only)
		api.Post("/system/reload", systemHandler.ReloadRoutes)

		// Authentication routes (no authentication required)
		auth := api.Group("/auth")
		{
			// Only enable registration and login if database is configured
			if db != nil {
				auth.Post("/register", authHandler.Register) // User registration
				auth.Post("/login", authHandler.Login)       // User login
			}
		}

		// Authenticated routes
		authenticated := api.Group("/")
		authenticated.Use(authMiddleware)
		{
			// Only enable database-dependent features if database is configured
			if db != nil {
				// User information
				authenticated.Get("/profile", authHandler.GetProfile)                   // Get current user info
				authenticated.Post("/auth/update-password", authHandler.ChangePassword) // Update user password (deprecated)
				authenticated.Put("/profile/password", authHandler.ChangePassword)      // Update user password

				// ZeroTier status
				authenticated.Get("/status", networkHandler.GetStatus)

				// Network management
				networks := authenticated.Group("/networks")
				{
					networks.Get("", networkHandler.GetNetworks)          // Get all networks
					networks.Post("", networkHandler.CreateNetwork)       // Create network
					networks.Get("/:id", networkHandler.GetNetwork)       // Get single network
					networks.Put("/:id", networkHandler.UpdateNetwork)    // Update network
					networks.Delete("/:id", networkHandler.DeleteNetwork) // Delete network

					// Member management (nested within network routes)
					networks.Get("/:id/members", memberHandler.GetMembers)                // Get member list
					networks.Get("/:id/members/:memberId", memberHandler.GetMember)       // Get single member
					networks.Put("/:id/members/:memberId", memberHandler.UpdateMember)    // Update member
					networks.Delete("/:id/members/:memberId", memberHandler.DeleteMember) // Delete member
				}
			}
		}

		// Admin-only routes
		admin := api.Group("/")
		admin.Use(authMiddleware)
		admin.Use(middleware.AdminRequired())
		{
			// System statistics
			admin.Get("/system/stats", systemHandler.GetSystemStats) // Get system resource statistics

			// User management
			admin.Get("/users", userHandler.GetAllUsers)                 // Get all users
			admin.Put("/users/:userId/role", userHandler.UpdateUserRole) // Update user role

			// Network import (admin only)
			admin.Get("/networks/importable", networkHandler.GetImportableNetworks) // Get list of importable networks
			admin.Post("/networks/import", networkHandler.ImportNetworks)           // Import specified networks
		}
	}
}
