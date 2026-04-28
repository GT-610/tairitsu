package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

type SetupService struct {
	runtimeService *RuntimeService
	stateService   *StateService
	userService    *UserService
	networkService *NetworkService
}

func NewSetupService(runtimeService *RuntimeService, stateService *StateService, userService *UserService, networkService *NetworkService) *SetupService {
	return &SetupService{
		runtimeService: runtimeService,
		stateService:   stateService,
		userService:    userService,
		networkService: networkService,
	}
}

func (s *SetupService) ConfigureDatabase(dbConfig models.DatabaseConfig) (database.Config, error) {
	if dbConfig.Type != string(database.SQLite) {
		return database.Config{}, fmt.Errorf("only SQLite is currently supported; use SQLite to complete initialization")
	}

	dbCfg := database.Config{
		Type: database.DatabaseType(dbConfig.Type),
		Path: dbConfig.Path,
		Host: dbConfig.Host,
		Port: dbConfig.Port,
		User: dbConfig.User,
		Pass: dbConfig.Pass,
		Name: dbConfig.Name,
	}

	db, err := database.NewDatabase(dbCfg)
	if err != nil {
		return database.Config{}, fmt.Errorf("database connection failed: %w", err)
	}

	if err := db.Init(); err != nil {
		db.Close()
		return database.Config{}, fmt.Errorf("database initialization failed: %w", err)
	}

	if dbCfg.Type == database.SQLite && dbCfg.Path == "" {
		dbCfg.Path = "data/tairitsu.db"
	}

	if err := s.stateService.SaveDatabaseConfig(dbCfg); err != nil {
		db.Close()
		return database.Config{}, fmt.Errorf("failed to save database configuration: %w", err)
	}

	s.runtimeService.BindDatabase(db)
	return dbCfg, nil
}

func (s *SetupService) SaveZeroTierConfig(controllerURL, tokenPath string) (*zerotier.Status, error) {
	if err := s.stateService.SaveZeroTierConfig(controllerURL, tokenPath); err != nil {
		return nil, fmt.Errorf("failed to save ZeroTier configuration: %w", err)
	}

	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create ZeroTier client: %w", err)
	}

	status, err := ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("ZeroTier client validation failed: %w", err)
	}

	s.runtimeService.BindZTClient(ztClient)
	return status, nil
}

func (s *SetupService) TestZeroTierConnection() (*zerotier.Status, error) {
	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create ZeroTier client: %w", err)
	}

	status, err := ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ZeroTier controller: %w", err)
	}

	return status, nil
}

func (s *SetupService) InitZTClientFromConfig() (*zerotier.Status, error) {
	return s.runtimeService.InitZTClientFromConfig()
}

func (s *SetupService) GetSetupStatus() SetupStatus {
	return s.stateService.GetSetupStatus(s.userService, s.networkService)
}

func (s *SetupService) GetRuntimeSettings() RuntimeSettings {
	return s.stateService.RuntimeSettings()
}

func (s *SetupService) UpdateRuntimeSettings(settings RuntimeSettings) error {
	return s.stateService.SaveRuntimeSettings(settings)
}

func (s *SetupService) InitializeAdminCreation() (string, error) {
	dbConfig := s.stateService.DatabaseConfig()
	if dbConfig.Type == "" {
		return "", fmt.Errorf("database configuration is incomplete; configure SQLite first")
	}
	if dbConfig.Type != database.SQLite {
		return "", fmt.Errorf("only SQLite is currently supported; %s initialization is not supported", dbConfig.Type)
	}

	if s.stateService.IsInitialized() {
		return "", fmt.Errorf("system is already initialized; cannot prepare first administrator creation again")
	}

	resetDoneKey := "admin_creation_reset_done"
	if config.GetTempSetting(resetDoneKey) == "true" {
		return string(dbConfig.Type), nil
	}

	hasAdmin, err := s.userService.HasAdminUser()
	if err == nil && hasAdmin {
		config.SetTempSetting(resetDoneKey, "true")
		return string(dbConfig.Type), nil
	}

	s.runtimeService.CloseCurrentDatabase()

	if err := database.ResetDatabase(dbConfig); err != nil {
		return "", fmt.Errorf("failed to initialize administrator account creation step: %w", err)
	}

	if err := s.runtimeService.ReopenConfiguredDatabase(); err != nil {
		return "", fmt.Errorf("failed to reinitialize database after reset: %w", err)
	}

	config.SetTempSetting(resetDoneKey, "true")
	return string(dbConfig.Type), nil
}

func (s *SetupService) SetInitialized(initialized bool) error {
	if initialized {
		if err := s.validateInitializationReady(); err != nil {
			return err
		}

		cfg := s.stateService.Config()
		if cfg.Security.JWTSecret == "" {
			cfg.Security.JWTSecret = generateRandomSecret(32)
			logger.Info("generated new JWT secret")
		}

		if cfg.Security.SessionSecret == "" {
			cfg.Security.SessionSecret = generateRandomSecret(32)
			logger.Info("generated new session secret")
		}

		if err := s.stateService.SaveConfig(); err != nil {
			return fmt.Errorf("failed to generate secure key: %w", err)
		}
	}

	if err := s.stateService.SetInitialized(initialized); err != nil {
		return fmt.Errorf("failed to set initialization state: %w", err)
	}

	return nil
}

func (s *SetupService) validateInitializationReady() error {
	if !s.stateService.DatabaseConfigured() {
		return fmt.Errorf("database configuration is incomplete; configure SQLite first")
	}

	if s.runtimeService.CurrentDatabase() == nil {
		if err := s.runtimeService.ReopenConfiguredDatabase(); err != nil {
			return fmt.Errorf("database is not ready: %w", err)
		}
	}

	hasAdmin, err := s.userService.HasAdminUser()
	if err != nil {
		return fmt.Errorf("failed to confirm administrator state: %w", err)
	}
	if !hasAdmin {
		return fmt.Errorf("create the first administrator account first")
	}

	if _, err := s.stateService.CreateZTClient(); err != nil {
		return fmt.Errorf("ZeroTier configuration is incomplete: %w", err)
	}

	status, err := s.runtimeService.InitZTClientFromConfig()
	if err != nil {
		return fmt.Errorf("ZeroTier connection validation failed: %w", err)
	}
	if status == nil || !status.Online {
		return fmt.Errorf("ZeroTier controller is currently unavailable")
	}

	return nil
}

func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		logger.Warn("failed to generate random key with crypto/rand; falling back to math/rand", zap.Error(err))

		r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		for i := range bytes {
			bytes[i] = byte(r.Intn(256))
		}
	}

	return base64.URLEncoding.EncodeToString(bytes)
}
