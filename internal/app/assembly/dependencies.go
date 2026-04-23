package assembly

import (
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
)

type Services struct {
	Network *services.NetworkService
	User    *services.UserService
	JWT     *services.JWTService
	State   *services.StateService
	Runtime *services.RuntimeService
	Setup   *services.SetupService
	System  *services.SystemService
}

type Handlers struct {
	Network *handlers.NetworkHandler
	Member  *handlers.MemberHandler
	Auth    *handlers.AuthHandler
	User    *handlers.UserHandler
	System  *handlers.SystemHandler
}

type Middleware struct {
	Auth        fiber.Handler
	SetupOnly   fiber.Handler
	RuntimeOnly fiber.Handler
	AdminOnly   fiber.Handler
}

type Dependencies struct {
	Config     *config.Config
	Database   database.DBInterface
	ZTClient   *zerotier.Client
	Services   Services
	Handlers   Handlers
	Middleware Middleware
}

func NewDependencies(cfg *config.Config, db database.DBInterface, ztClient *zerotier.Client) *Dependencies {
	networkService := services.NewNetworkService(ztClient, db)

	var userService *services.UserService
	if db != nil {
		userService = services.NewUserServiceWithDB(db)
	} else {
		userService = services.NewUserServiceWithoutDB()
	}

	stateService := services.NewStateServiceWithConfig(cfg)
	runtimeService := services.NewRuntimeService(userService, networkService, stateService)
	setupService := services.NewSetupService(runtimeService, stateService, userService, networkService)
	systemService := services.NewSystemService()
	jwtSecret := ""
	if cfg != nil {
		jwtSecret = cfg.Security.JWTSecret
	}
	jwtService := services.NewJWTService(jwtSecret)

	authHandler := handlers.NewAuthHandler(userService, jwtService, runtimeService)

	return &Dependencies{
		Config:   cfg,
		Database: db,
		ZTClient: ztClient,
		Services: Services{
			Network: networkService,
			User:    userService,
			JWT:     jwtService,
			State:   stateService,
			Runtime: runtimeService,
			Setup:   setupService,
			System:  systemService,
		},
		Handlers: Handlers{
			Network: handlers.NewNetworkHandler(networkService),
			Member:  handlers.NewMemberHandler(networkService),
			Auth:    authHandler,
			User:    handlers.NewUserHandler(userService),
			System:  handlers.NewSystemHandler(setupService, systemService),
		},
		Middleware: Middleware{
			Auth:        middleware.AuthMiddleware(jwtService),
			SetupOnly:   middleware.SetupOnlyWithState(stateService),
			RuntimeOnly: middleware.InitializedOnlyWithState(stateService),
			AdminOnly:   middleware.AdminRequired(),
		},
	}
}
