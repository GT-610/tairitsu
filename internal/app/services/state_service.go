package services

import (
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/zerotier"
)

type SetupStatus struct {
	Initialized bool             `json:"initialized"`
	HasDatabase bool             `json:"hasDatabase,omitempty"`
	HasAdmin    bool             `json:"hasAdmin,omitempty"`
	ZTStatus    *zerotier.Status `json:"ztStatus,omitempty"`
}

type StateService struct{}

func NewStateService() *StateService {
	return &StateService{}
}

func (s *StateService) Config() *config.Config {
	return config.AppConfig
}

func (s *StateService) IsInitialized() bool {
	cfg := s.Config()
	return cfg != nil && cfg.Initialized
}

func (s *StateService) DatabaseConfigured() bool {
	cfg := s.Config()
	return cfg != nil && cfg.Database.Type != ""
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
