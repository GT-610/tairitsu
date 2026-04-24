package services

import (
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/zerotier"
)

type SetupStatus struct {
	Initialized             bool             `json:"initialized"`
	HasDatabase             bool             `json:"hasDatabase"`
	DatabaseConfigured      bool             `json:"databaseConfigured"`
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
	cfg          *config.Config
	globalBacked bool
}

func NewStateService() *StateService {
	return &StateService{globalBacked: true}
}

func NewStateServiceWithConfig(cfg *config.Config) *StateService {
	return &StateService{cfg: cfg}
}

func (s *StateService) Config() *config.Config {
	if s.cfg != nil {
		return s.cfg
	}
	if s.globalBacked {
		return config.AppConfig
	}
	return nil
}

func (s *StateService) SaveConfig() error {
	cfg := s.ensureConfig()
	if err := config.SaveConfig(cfg); err != nil {
		return err
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

func (s *StateService) GetSetupStatus(userService *UserService, networkService *NetworkService) SetupStatus {
	cfg := s.Config()
	databaseConfigured := s.DatabaseConfigured()
	zeroTierConfigured := cfg != nil && cfg.ZeroTier.URL != "" && cfg.ZeroTier.TokenPath != ""

	status := SetupStatus{
		Initialized:             s.IsInitialized(),
		HasDatabase:             databaseConfigured,
		DatabaseConfigured:      databaseConfigured,
		ZeroTierConfigured:      zeroTierConfigured,
		AdminCreationPrepared:   config.GetTempSetting("admin_creation_reset_done") == "true",
		AllowPublicRegistration: config.AllowPublicRegistration(s.Config()),
	}

	if databaseConfigured && cfg != nil {
		status.DatabaseConfig = &SetupDatabase{
			Type: string(cfg.Database.Type),
			Path: cfg.Database.Path,
		}
	}

	if zeroTierConfigured && cfg != nil {
		status.ZeroTierConfig = &SetupZeroTier{
			ControllerURL: cfg.ZeroTier.URL,
			TokenPath:     cfg.ZeroTier.TokenPath,
		}
	}

	if databaseConfigured && userService != nil {
		users := userService.GetAllUsers()
		for _, user := range users {
			if user.Role == "admin" {
				status.HasAdmin = true
				status.AdminUsername = user.Username
				break
			}
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
	if s.globalBacked {
		config.AppConfig = cfg
	}
	return cfg
}
