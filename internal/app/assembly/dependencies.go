package assembly

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
)

type Services struct {
	Network    *services.NetworkService
	User       *services.UserService
	Session    *services.SessionService
	State      *services.StateService
	Runtime    *services.RuntimeService
	Setup      *services.SetupService
	System     *services.SystemService
	SetupToken *services.SetupToken
}

type Handlers struct {
	Network *handlers.NetworkHandler
	Member  *handlers.MemberHandler
	Auth    *handlers.AuthHandler
	User    *handlers.UserHandler
	System  *handlers.SystemHandler
}

type Middleware struct {
	Auth               fiber.Handler
	SetupOnly          fiber.Handler
	SetupToken         fiber.Handler
	SetupTokenOptional fiber.Handler
	RuntimeOnly        fiber.Handler
	AdminOnly          fiber.Handler
}

type Dependencies struct {
	Config     *config.Config
	Database   database.DBInterface
	ZTClient   *zerotier.Client
	Services   Services
	Handlers   Handlers
	Middleware Middleware
}

func NewDependencies(cfg *config.Config, db database.DBInterface, ztClient *zerotier.Client) (*Dependencies, error) {
	// The setup endpoints have to work before any account exists, so while the
	// instance is uninitialized they are gated by a token that is generated here
	// and announced at startup. Once setup is done, SetupOnly closes them and
	// there is nothing left to issue.
	var setupToken *services.SetupToken
	if cfg == nil || !cfg.Initialized {
		issued, err := services.NewSetupToken()
		if err != nil {
			return nil, fmt.Errorf("failed to issue the setup token: %w", err)
		}
		setupToken = issued
	}

	networkService := services.NewNetworkService(ztClient, db)

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)

	stateService := services.NewStateServiceWithConfig(cfg)
	runtimeService := services.NewRuntimeService(userService, sessionService, networkService, stateService)
	setupService := services.NewSetupService(runtimeService, stateService, userService, networkService)
	systemService := services.NewSystemService()

	authHandler := handlers.NewAuthHandler(userService, sessionService, runtimeService, stateService)

	return &Dependencies{
		Config:   cfg,
		Database: db,
		ZTClient: ztClient,
		Services: Services{
			Network:    networkService,
			User:       userService,
			Session:    sessionService,
			State:      stateService,
			Runtime:    runtimeService,
			Setup:      setupService,
			System:     systemService,
			SetupToken: setupToken,
		},
		Handlers: Handlers{
			Network: handlers.NewNetworkHandler(networkService),
			Member:  handlers.NewMemberHandler(networkService),
			Auth:    authHandler,
			User:    handlers.NewUserHandler(userService),
			System:  handlers.NewSystemHandler(setupService, systemService),
		},
		Middleware: Middleware{
			Auth:               middleware.AuthMiddleware(sessionService, userService),
			SetupOnly:          middleware.SetupOnlyWithState(stateService),
			SetupToken:         middleware.SetupTokenRequired(setupToken),
			SetupTokenOptional: middleware.SetupTokenOptional(setupToken),
			RuntimeOnly:        middleware.InitializedOnlyWithState(stateService),
			AdminOnly:          middleware.AdminRequiredWithUserService(userService),
		},
	}, nil
}
