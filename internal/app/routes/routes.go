package routes

import (
	"github.com/GT-610/tairitsu/internal/app/assembly"
	"github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// SetupRoutes configures application routes
func SetupRoutes(router *fiber.App, dependencies *assembly.Dependencies) {
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

	networkHandler := dependencies.Handlers.Network
	memberHandler := dependencies.Handlers.Member
	authHandler := dependencies.Handlers.Auth
	userHandler := dependencies.Handlers.User
	systemHandler := dependencies.Handlers.System

	authMiddleware := dependencies.Middleware.Auth
	setupOnly := dependencies.Middleware.SetupOnly
	runtimeOnly := dependencies.Middleware.RuntimeOnly
	adminOnly := dependencies.Middleware.AdminOnly

	// API routes group
	api := router.Group("/api")
	{
		// Health check
		api.Get("/health", func(c fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
		})

		// System status check (no authentication required)
		api.Get("/system/status", systemHandler.GetSystemStatus)

		auth := api.Group("/auth")
		{
			auth.Post("/register", setupOnly, authHandler.Register)
			auth.Post("/login", runtimeOnly, authHandler.Login)
		}

		api.Post("/system/database", setupOnly, systemHandler.ConfigureDatabase)
		api.Get("/system/zerotier/test", setupOnly, systemHandler.TestZeroTierConnection)
		api.Post("/system/zerotier/config", setupOnly, systemHandler.SaveZeroTierConfig)
		api.Post("/system/zerotier/init", setupOnly, systemHandler.InitZeroTierClient)
		api.Post("/system/initialized", setupOnly, systemHandler.SetInitialized)
		api.Post("/system/admin/init", setupOnly, systemHandler.InitializeAdminCreation)
		api.Post("/system/reload", setupOnly, systemHandler.ReloadRoutes)

		api.Get("/profile", runtimeOnly, authMiddleware, authHandler.GetProfile)
		api.Post("/auth/update-password", runtimeOnly, authMiddleware, authHandler.ChangePassword)
		api.Put("/profile/password", runtimeOnly, authMiddleware, authHandler.ChangePassword)

		api.Get("/status", runtimeOnly, authMiddleware, networkHandler.GetStatus)

		api.Get("/networks", runtimeOnly, authMiddleware, networkHandler.GetNetworks)
		api.Post("/networks", runtimeOnly, authMiddleware, networkHandler.CreateNetwork)
		api.Get("/networks/:id", runtimeOnly, authMiddleware, networkHandler.GetNetwork)
		api.Put("/networks/:id", runtimeOnly, authMiddleware, networkHandler.UpdateNetwork)
		api.Put("/networks/:id/metadata", runtimeOnly, authMiddleware, networkHandler.UpdateNetworkMetadata)
		api.Delete("/networks/:id", runtimeOnly, authMiddleware, networkHandler.DeleteNetwork)

		api.Get("/networks/:id/members", runtimeOnly, authMiddleware, memberHandler.GetMembers)
		api.Get("/networks/:id/members/:memberId", runtimeOnly, authMiddleware, memberHandler.GetMember)
		api.Put("/networks/:id/members/:memberId", runtimeOnly, authMiddleware, memberHandler.UpdateMember)
		api.Delete("/networks/:id/members/:memberId", runtimeOnly, authMiddleware, memberHandler.DeleteMember)

		// Admin-only routes
		api.Get("/system/stats", runtimeOnly, authMiddleware, adminOnly, systemHandler.GetSystemStats)
		api.Get("/users", runtimeOnly, authMiddleware, adminOnly, userHandler.GetAllUsers)
		api.Put("/users/:userId/role", runtimeOnly, authMiddleware, adminOnly, userHandler.UpdateUserRole)
		api.Get("/admin/networks/importable", runtimeOnly, authMiddleware, adminOnly, networkHandler.GetImportableNetworks)
		api.Post("/admin/networks/import", runtimeOnly, authMiddleware, adminOnly, networkHandler.ImportNetworks)
		api.Get("/admin/planet/identity", runtimeOnly, authMiddleware, adminOnly, handlers.GetIdentityHandler)
		api.Post("/admin/planet/generate", runtimeOnly, authMiddleware, adminOnly, handlers.GeneratePlanetHandler)
		api.Post("/admin/planet/keys", runtimeOnly, authMiddleware, adminOnly, handlers.GenerateSigningKeysHandler)
	}
}
