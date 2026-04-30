package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

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

var (
	ErrSetupUnsupportedDatabase        = errors.New("setup.unsupported_database")
	ErrSetupInvalidConfig              = errors.New("setup.invalid_config")
	ErrSetupDatabaseConnectionFailed   = errors.New("setup.database_connection_failed")
	ErrSetupDatabaseInitialization     = errors.New("setup.database_initialization_failed")
	ErrSetupDatabaseConfigSaveFailed   = errors.New("setup.database_config_save_failed")
	ErrSetupConfigSaveFailed           = errors.New("setup.config_save_failed")
	ErrSetupZeroTierConfigSaveFailed   = errors.New("setup.zerotier_config_save_failed")
	ErrSetupZeroTierClientCreateFailed = errors.New("setup.zerotier_client_create_failed")
	ErrSetupZeroTierValidationFailed   = errors.New("setup.zerotier_validation_failed")
	ErrSetupAlreadyInitialized         = errors.New("setup.already_initialized")
	ErrSetupAdminStateCheckFailed      = errors.New("setup.admin_state_check_failed")
	ErrSetupAdminCreationInitFailed    = errors.New("setup.admin_creation_init_failed")
	ErrSetupDatabaseReopenFailed       = errors.New("setup.database_reopen_failed")
	ErrSetupSecretGenerationFailed     = errors.New("setup.secret_generation_failed")
	ErrSetupInitializationStateFailed  = errors.New("setup.initialization_state_failed")
	ErrSetupAdminRequired              = errors.New("setup.admin_required")
	ErrSetupZeroTierUnavailable        = errors.New("setup.zerotier_unavailable")
)

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
		return database.Config{}, ErrSetupUnsupportedDatabase
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
		return database.Config{}, fmt.Errorf("%w: %v", ErrSetupDatabaseConnectionFailed, err)
	}

	if err := db.Init(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			logger.Warn("failed to close database after initialization error", zap.Error(closeErr))
			return database.Config{}, fmt.Errorf("%w: %v; close database: %v", ErrSetupDatabaseInitialization, err, closeErr)
		}
		return database.Config{}, fmt.Errorf("%w: %v", ErrSetupDatabaseInitialization, err)
	}

	if dbCfg.Type == database.SQLite && dbCfg.Path == "" {
		dbCfg.Path = "data/tairitsu.db"
	}

	if err := s.stateService.SaveDatabaseConfig(dbCfg); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			logger.Warn("failed to close database after save configuration error", zap.Error(closeErr))
			return database.Config{}, fmt.Errorf("%w: %v; close database: %v", ErrSetupDatabaseConfigSaveFailed, err, closeErr)
		}
		return database.Config{}, fmt.Errorf("%w: %v", ErrSetupDatabaseConfigSaveFailed, err)
	}

	s.runtimeService.BindDatabase(db)
	return dbCfg, nil
}

func (s *SetupService) SaveZeroTierConfig(controllerURL, tokenPath string) (*zerotier.Status, error) {
	if err := s.stateService.SaveZeroTierConfig(controllerURL, tokenPath); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetupZeroTierConfigSaveFailed, err)
	}

	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetupZeroTierClientCreateFailed, err)
	}

	status, err := ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetupZeroTierValidationFailed, err)
	}

	s.runtimeService.BindZTClient(ztClient)
	return status, nil
}

func (s *SetupService) TestZeroTierConnection() (*zerotier.Status, error) {
	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetupZeroTierClientCreateFailed, err)
	}

	status, err := ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetupZeroTierValidationFailed, err)
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
		return "", ErrSetupInvalidConfig
	}
	if dbConfig.Type != database.SQLite {
		return "", fmt.Errorf("%w: %s", ErrSetupUnsupportedDatabase, dbConfig.Type)
	}

	if s.stateService.IsInitialized() {
		return "", ErrSetupAlreadyInitialized
	}

	resetDoneKey := "admin_creation_reset_done"
	if config.GetTempSetting(resetDoneKey) == "true" {
		return string(dbConfig.Type), nil
	}

	hasAdmin, err := s.userService.HasAdminUser()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSetupAdminStateCheckFailed, err)
	}
	if hasAdmin {
		config.SetTempSetting(resetDoneKey, "true")
		return string(dbConfig.Type), nil
	}

	s.runtimeService.CloseCurrentDatabase()

	if err := database.ResetDatabase(dbConfig); err != nil {
		return "", fmt.Errorf("%w: %v", ErrSetupAdminCreationInitFailed, err)
	}

	if err := s.runtimeService.ReopenConfiguredDatabase(); err != nil {
		return "", fmt.Errorf("%w: %v", ErrSetupDatabaseReopenFailed, err)
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
			secret, err := generateRandomSecret(32)
			if err != nil {
				return fmt.Errorf("%w: JWT secret: %v", ErrSetupSecretGenerationFailed, err)
			}
			cfg.Security.JWTSecret = secret
			logger.Info("generated new JWT secret")
		}

		if err := s.stateService.SaveConfig(); err != nil {
			return fmt.Errorf("%w: %v", ErrSetupConfigSaveFailed, err)
		}
	}

	if err := s.stateService.SetInitialized(initialized); err != nil {
		return fmt.Errorf("%w: %v", ErrSetupInitializationStateFailed, err)
	}

	return nil
}

func (s *SetupService) validateInitializationReady() error {
	if !s.stateService.DatabaseConfigured() {
		return ErrSetupInvalidConfig
	}

	if s.runtimeService.CurrentDatabase() == nil {
		if err := s.runtimeService.ReopenConfiguredDatabase(); err != nil {
			return fmt.Errorf("%w: %v", ErrSetupDatabaseReopenFailed, err)
		}
	}

	hasAdmin, err := s.userService.HasAdminUser()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSetupAdminStateCheckFailed, err)
	}
	if !hasAdmin {
		return ErrSetupAdminRequired
	}

	if _, err := s.stateService.CreateZTClient(); err != nil {
		return fmt.Errorf("%w: %v", ErrSetupZeroTierClientCreateFailed, err)
	}

	status, err := s.runtimeService.InitZTClientFromConfig()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSetupZeroTierValidationFailed, err)
	}
	if status == nil || !status.Online {
		return ErrSetupZeroTierUnavailable
	}

	return nil
}

func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
