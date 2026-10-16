package services

import (
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

type SetupStatus struct {
	Initialized             bool             `json:"initialized"`
	DatabaseConfigured      bool             `json:"databaseConfigured"`
	SetupAuthorized         bool             `json:"setupAuthorized"`
	HasAdmin                bool             `json:"hasAdmin"`
	ZeroTierConfigured      bool             `json:"zerotierConfigured"`
	AdminCreationPrepared   bool             `json:"adminCreationPrepared"`
	AdminUsername           string           `json:"adminUsername,omitempty"`
	DatabaseConfig          *SetupDatabase   `json:"databaseConfig,omitempty"`
	ZeroTierConfig          *SetupZeroTier   `json:"zeroTierConfig,omitempty"`
	AllowPublicRegistration bool             `json:"allowPublicRegistration"`
	ZTStatus                *zerotier.Status `json:"ztStatus,omitempty"`
}

type SetupDatabase struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

type SetupZeroTier struct {
	ControllerURL string `json:"controllerUrl"`
	TokenPath     string `json:"tokenPath"`
}

type RuntimeSettings struct {
	AllowPublicRegistration bool `json:"allow_public_registration"`
}

type StateService struct {
	cfg *config.Config
}

func NewStateServiceWithConfig(cfg *config.Config) *StateService {
	return &StateService{cfg: cfg}
}

func (s *StateService) Config() *config.Config {
	if s.cfg != nil {
		return s.cfg
	}
	return nil
}

func (s *StateService) SetInitialized(initialized bool) error {
	cfg := s.ensureConfig()
	cfg.Initialized = initialized
	return config.SaveConfig(cfg)
}

func (s *StateService) IsInitialized() bool {
	cfg := s.Config()
	return cfg != nil && cfg.Initialized
}

func (s *StateService) DatabaseConfigured() bool {
	cfg := s.Config()
	return cfg != nil && cfg.Database.Type != ""
}

func (s *StateService) DatabaseConfig() database.Config {
	return database.LoadConfigFromApp(s.Config())
}

func (s *StateService) SaveDatabaseConfig(dbCfg database.Config) error {
	cfg := s.ensureConfig()
	if err := database.SaveConfigToApp(cfg, dbCfg); err != nil {
		return err
	}
	return nil
}

func (s *StateService) SaveZeroTierConfig(url, tokenPath string) error {
	cfg := s.ensureConfig()
	if err := config.SetZTConfigOn(cfg, url, tokenPath); err != nil {
		return err
	}
	return nil
}

func (s *StateService) RuntimeSettings() RuntimeSettings {
	return RuntimeSettings{
		AllowPublicRegistration: config.AllowPublicRegistration(s.Config()),
	}
}

func (s *StateService) SaveRuntimeSettings(settings RuntimeSettings) error {
	cfg := s.ensureConfig()
	return config.SetAllowPublicRegistrationOn(cfg, settings.AllowPublicRegistration)
}

func (s *StateService) CreateZTClient() (*zerotier.Client, error) {
	return zerotier.NewClientWithConfig(s.Config())
}

// GetSetupStatus reports how far the first-run wizard has progressed.
//
// setupCaller reports whether the request presented the setup token. The
// endpoint that serves this is public, because the client has to know whether
// to show the wizard before anyone can sign in, so that token is what decides
// whether the reply carries the deployment layout: the administrator name, the
// SQLite path, and the controller URL and token path.
//
// The details are withheld once the instance is initialized even from a caller
// holding the token. The token exists to close the first-run window, and this
// process keeps the one it issued until it restarts, so honouring it afterwards
// would leave the deployment layout readable from a stale log line forever.
func (s *StateService) GetSetupStatus(userService *UserService, networkService *NetworkService, setupCaller bool) SetupStatus {
	cfg := s.Config()
	initialized := s.IsInitialized()
	includeSensitive := setupCaller && !initialized
	databaseConfigured := s.DatabaseConfigured()
	zeroTierConfigured := cfg != nil && cfg.ZeroTier.URL != "" && cfg.ZeroTier.TokenPath != ""

	status := SetupStatus{
		Initialized:             initialized,
		DatabaseConfigured:      databaseConfigured,
		SetupAuthorized:         includeSensitive,
		ZeroTierConfigured:      zeroTierConfigured,
		AdminCreationPrepared:   config.GetTempSetting("admin_creation_reset_done") == "true",
		AllowPublicRegistration: config.AllowPublicRegistration(cfg),
	}

	if includeSensitive && databaseConfigured && cfg != nil {
		status.DatabaseConfig = &SetupDatabase{
			Type: string(cfg.Database.Type),
			Path: cfg.Database.Path,
		}
	}

	if includeSensitive && zeroTierConfigured && cfg != nil {
		status.ZeroTierConfig = &SetupZeroTier{
			ControllerURL: cfg.ZeroTier.URL,
			TokenPath:     cfg.ZeroTier.TokenPath,
		}
	}

	if databaseConfigured && userService != nil {
		if includeSensitive {
			// The wizard shows which account was created, so the name is needed.
			users, err := userService.GetAllUsers()
			if err != nil {
				logger.Warn("GetSetupStatus: GetAllUsers failed", zap.Error(err))
			}
			for _, user := range users {
				if user.Role == "admin" {
					status.HasAdmin = true
					status.AdminUsername = user.Username
					break
				}
			}
		} else {
			// The public answer only needs to know whether anyone can sign in,
			// which is a count rather than a full table read.
			hasAdmin, err := userService.HasAdminUser()
			if err != nil {
				logger.Warn("GetSetupStatus: HasAdminUser failed", zap.Error(err))
			}
			status.HasAdmin = hasAdmin
		}
	}

	if zeroTierConfigured {
		if networkService != nil {
			if ztStatus, err := networkService.GetStatus(); err == nil {
				status.ZTStatus = ztStatus
			}
		}
		if status.ZTStatus == nil {
			if ztClient, err := s.CreateZTClient(); err == nil {
				if ztStatus, err := ztClient.GetStatus(); err == nil {
					status.ZTStatus = ztStatus
				}
			}
		}
	}

	if status.Initialized && status.ZTStatus == nil && networkService != nil {
		if ztStatus, err := networkService.GetStatus(); err == nil {
			status.ZTStatus = ztStatus
		} else {
			status.ZTStatus = &zerotier.Status{
				Version: "unknown",
				Address: "",
				Online:  false,
			}
		}
	}

	return status
}

func (s *StateService) ensureConfig() *config.Config {
	cfg := s.Config()
	if cfg == nil {
		cfg = &config.Config{}
	}
	s.cfg = cfg
	return cfg
}
