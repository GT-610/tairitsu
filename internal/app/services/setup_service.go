package services

import (
	"errors"
	"fmt"

	"github.com/GT-610/tempest-zt/internal/app/config"
	"github.com/GT-610/tempest-zt/internal/app/database"
	"github.com/GT-610/tempest-zt/internal/app/logger"
	"github.com/GT-610/tempest-zt/internal/app/models"
	"github.com/GT-610/tempest-zt/internal/zerotier"
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
	ErrSetupDatabasePathRejected       = errors.New("setup.database_path_rejected")
	ErrSetupDatabaseConnectionFailed   = errors.New("setup.database_connection_failed")
	ErrSetupDatabaseInitialization     = errors.New("setup.database_initialization_failed")
	ErrSetupDatabaseConfigSaveFailed   = errors.New("setup.database_config_save_failed")
	ErrSetupZeroTierConfigSaveFailed   = errors.New("setup.zerotier_config_save_failed")
	ErrSetupInvalidControllerURL       = errors.New("setup.invalid_controller_url")
	ErrSetupZeroTierClientCreateFailed = errors.New("setup.zerotier_client_create_failed")
	ErrSetupZeroTierValidationFailed   = errors.New("setup.zerotier_validation_failed")
	ErrSetupAlreadyInitialized         = errors.New("setup.already_initialized")
	ErrSetupAdminStateCheckFailed      = errors.New("setup.admin_state_check_failed")
	ErrSetupAdminCreationInitFailed    = errors.New("setup.admin_creation_init_failed")
	ErrSetupDatabaseReopenFailed       = errors.New("setup.database_reopen_failed")
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

func (s *SetupService) ConfigureDatabase(dbConfig config.DatabaseConfig) (database.Config, error) {
	if dbConfig.Type != string(database.SQLite) {
		return database.Config{}, ErrSetupUnsupportedDatabase
	}

	// The caller picks the SQLite path and a later reset deletes it, so it is
	// confined to the mounted data directory before anything is created.
	safePath, err := config.ResolveDataPath(dbConfig.Path)
	if err != nil {
		return database.Config{}, fmt.Errorf("%w: %v", ErrSetupDatabasePathRejected, err)
	}

	dbCfg := database.Config{
		Type: database.DatabaseType(dbConfig.Type),
		Path: safePath,
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
	// Reject a malformed address as the caller's mistake rather than letting it
	// reach the client, which would report it as a connection failure.
	normalizedURL, err := config.NormalizeControllerURL(controllerURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetupInvalidControllerURL, err)
	}

	if err := s.stateService.SaveZeroTierConfig(normalizedURL, tokenPath); err != nil {
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

func (s *SetupService) GetSetupStatus(setupCaller bool) SetupStatus {
	return s.stateService.GetSetupStatus(s.userService, s.networkService, setupCaller)
}

// CreateInitialAdministrator creates the first account and grants it the
// administrator role. It is the only path that can produce an administrator
// without an existing one vouching for it, so the route serving it is gated by
// the setup token rather than by a session.
//
// The "already claimed" decision is made by PromoteFirstAdministrator inside the
// insert transaction, not by a check here: asking first and writing later would
// let two concurrent callers both observe "no administrator yet".
func (s *SetupService) CreateInitialAdministrator(username, password string) (*models.User, error) {
	if s.stateService.IsInitialized() {
		return nil, ErrSetupAlreadyInitialized
	}
	if s.stateService.DatabaseConfig().Type == "" {
		return nil, ErrSetupInvalidConfig
	}

	user, err := s.userService.Register(&models.RegisterRequest{
		Username: username,
		Password: password,
	}, "admin")
	if err != nil {
		return nil, err
	}

	// The controller client is created here rather than on registration: this is
	// the one account creation that belongs to the setup flow, and proving the
	// controller is reachable is part of finishing setup. A failure is not fatal
	// because the wizard validates the connection again before it marks the
	// instance initialized.
	if s.runtimeService != nil {
		if _, initErr := s.runtimeService.InitZTClientFromConfig(); initErr != nil {
			logger.Warn("ZeroTier client initialization failed after administrator creation; it will be retried later", zap.Error(initErr))
		}
	}

	return user, nil
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

	status, err := s.runtimeService.InitZTClientFromConfig()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSetupZeroTierValidationFailed, err)
	}
	if status == nil || !status.Online {
		return ErrSetupZeroTierUnavailable
	}

	return nil
}
