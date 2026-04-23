package services

import (
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/zerotier"
)

type SetupStatus struct {
	Initialized bool             `json:"initialized"`
	HasDatabase bool             `json:"hasDatabase,omitempty"`
	HasAdmin    bool             `json:"hasAdmin,omitempty"`
	ZTStatus    *zerotier.Status `json:"ztStatus,omitempty"`
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

func (s *StateService) CreateZTClient() (*zerotier.Client, error) {
	return zerotier.NewClientWithConfig(s.Config())
}

func (s *StateService) GetSetupStatus(userService *UserService, networkService *NetworkService) SetupStatus {
	if !s.IsInitialized() {
		return SetupStatus{Initialized: false}
	}

	status := SetupStatus{
		Initialized: true,
		HasDatabase: s.DatabaseConfigured(),
	}

	if status.HasDatabase && userService != nil {
		users := userService.GetAllUsers()
		for _, user := range users {
			if user.Role == "admin" {
				status.HasAdmin = true
				break
			}
		}
	}

	if networkService != nil {
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
