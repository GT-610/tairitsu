package services

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// RuntimeService coordinates mutable runtime dependencies such as the active DB and ZeroTier client.
type RuntimeService struct {
	userService    *UserService
	sessionService *SessionService
	networkService *NetworkService
	stateService   *StateService
}

func NewRuntimeService(userService *UserService, sessionService *SessionService, networkService *NetworkService, stateService *StateService) *RuntimeService {
	return &RuntimeService{
		userService:    userService,
		sessionService: sessionService,
		networkService: networkService,
		stateService:   stateService,
	}
}

func (s *RuntimeService) CurrentDatabase() database.DBInterface {
	if s.userService != nil && s.userService.GetDB() != nil {
		return s.userService.GetDB()
	}
	if s.networkService != nil && s.networkService.GetDB() != nil {
		return s.networkService.GetDB()
	}
	return nil
}

func (s *RuntimeService) BindDatabase(db database.DBInterface) {
	if current := s.CurrentDatabase(); current != nil && current != db {
		if err := current.Close(); err != nil {
			logger.Warn("failed to close previous database connection", zap.Error(err))
		}
	}

	if s.userService != nil {
		s.userService.SetDB(db)
	}
	if s.sessionService != nil {
		s.sessionService.SetDB(db)
	}
	if s.networkService != nil {
		s.networkService.SetDB(db)
	}
}

func (s *RuntimeService) CloseCurrentDatabase() {
	current := s.CurrentDatabase()
	if current == nil {
		return
	}

	if err := current.Close(); err != nil {
		logger.Warn("failed to close current database connection", zap.Error(err))
	}

	if s.userService != nil {
		s.userService.SetDB(nil)
	}
	if s.sessionService != nil {
		s.sessionService.SetDB(nil)
	}
	if s.networkService != nil {
		s.networkService.SetDB(nil)
	}
}

func (s *RuntimeService) ReopenConfiguredDatabase() error {
	dbConfig := s.stateService.DatabaseConfig()
	if dbConfig.Type == "" {
		return fmt.Errorf("database is not configured")
	}

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to reopen database: %w", err)
	}

	if err := db.Init(); err != nil {
		db.Close()
		return fmt.Errorf("failed to reinitialize database: %w", err)
	}

	s.BindDatabase(db)
	return nil
}

func (s *RuntimeService) BindZTClient(client *zerotier.Client) {
	if s.networkService == nil {
		return
	}
	s.networkService.SetZTClient(client)
}

func (s *RuntimeService) InitZTClientFromConfig() (*zerotier.Status, error) {
	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create ZeroTier client: %w", err)
	}

	s.BindZTClient(ztClient)

	status, err := s.networkService.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("ZeroTier client validation failed after initialization: %w", err)
	}

	return status, nil
}
