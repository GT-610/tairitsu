package handlers

import (
	"errors"
	"net/url"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func setupErrorResponse(c fiber.Ctx, err error) error {
	code := "system.internal_error"
	message := "Internal server error"
	status := fiber.StatusInternalServerError

	// Determine HTTP status based on error type
	switch {
	case errors.Is(err, services.ErrSetupUnsupportedDatabase),
		errors.Is(err, services.ErrSetupInvalidConfig):
		status = fiber.StatusBadRequest
	case errors.Is(err, services.ErrSetupDatabaseConnectionFailed),
		errors.Is(err, services.ErrSetupDatabaseInitialization),
		errors.Is(err, services.ErrSetupDatabaseConfigSaveFailed),
		errors.Is(err, services.ErrSetupConfigSaveFailed),
		errors.Is(err, services.ErrSetupZeroTierConfigSaveFailed),
		errors.Is(err, services.ErrSetupAdminStateCheckFailed),
		errors.Is(err, services.ErrSetupAdminCreationInitFailed),
		errors.Is(err, services.ErrSetupDatabaseReopenFailed),
		errors.Is(err, services.ErrSetupSecretGenerationFailed),
		errors.Is(err, services.ErrSetupInitializationStateFailed):
		status = fiber.StatusInternalServerError
	case errors.Is(err, services.ErrSetupAdminRequired),
		errors.Is(err, services.ErrSetupAlreadyInitialized):
		status = fiber.StatusConflict
	case errors.Is(err, services.ErrSetupZeroTierUnavailable),
		errors.Is(err, services.ErrSetupZeroTierValidationFailed),
		errors.Is(err, services.ErrSetupZeroTierClientCreateFailed):
		status = fiber.StatusServiceUnavailable
	}

	// Map error to code and message
	switch {
	case errors.Is(err, services.ErrSetupUnsupportedDatabase):
		code = "setup.unsupported_database"
		message = "Only SQLite is currently supported"
	case errors.Is(err, services.ErrSetupInvalidConfig):
		code = "setup.invalid_config"
		message = "Setup configuration is incomplete"
	case errors.Is(err, services.ErrSetupDatabaseConnectionFailed):
		code = "setup.database_connection_failed"
		message = "Database connection failed"
	case errors.Is(err, services.ErrSetupDatabaseInitialization):
		code = "setup.database_initialization_failed"
		message = "Database initialization failed"
	case errors.Is(err, services.ErrSetupDatabaseConfigSaveFailed):
		code = "setup.database_config_save_failed"
		message = "Failed to save database configuration"
	case errors.Is(err, services.ErrSetupConfigSaveFailed):
		code = "setup.config_save_failed"
		message = "Failed to save setup configuration"
	case errors.Is(err, services.ErrSetupZeroTierConfigSaveFailed):
		code = "setup.zerotier_config_save_failed"
		message = "Failed to save ZeroTier configuration"
	case errors.Is(err, services.ErrSetupZeroTierClientCreateFailed):
		code = "setup.zerotier_client_create_failed"
		message = "Failed to create ZeroTier client"
	case errors.Is(err, services.ErrSetupZeroTierValidationFailed):
		code = "setup.zerotier_validation_failed"
		message = "ZeroTier controller validation failed"
	case errors.Is(err, services.ErrSetupAlreadyInitialized):
		code = "setup.already_initialized"
		message = "System is already initialized"
	case errors.Is(err, services.ErrSetupAdminStateCheckFailed):
		code = "setup.admin_state_check_failed"
		message = "Failed to confirm administrator state"
	case errors.Is(err, services.ErrSetupAdminCreationInitFailed):
		code = "setup.admin_creation_init_failed"
		message = "Failed to initialize administrator account creation"
	case errors.Is(err, services.ErrSetupDatabaseReopenFailed):
		code = "setup.database_reopen_failed"
		message = "Failed to reopen configured database"
	case errors.Is(err, services.ErrSetupSecretGenerationFailed):
		code = "setup.secret_generation_failed"
		message = "Failed to generate secure secret"
	case errors.Is(err, services.ErrSetupInitializationStateFailed):
		code = "setup.initialization_state_failed"
		message = "Failed to update initialization state"
	case errors.Is(err, services.ErrSetupAdminRequired):
		code = "setup.admin_required"
		message = "Create the first administrator account first"
	case errors.Is(err, services.ErrSetupZeroTierUnavailable):
		code = "setup.zerotier_unavailable"
		message = "ZeroTier controller is currently unavailable"
	}
	return writeErrorResponseWithCode(c, status, code, message)
}

// SystemHandler handles system-related API endpoints and operations
type SystemHandler struct {
	setupService  *services.SetupService
	systemService *services.SystemService
	// Database configuration is stored in config file
}

// NewSystemHandler creates a new system handler instance
func NewSystemHandler(
	setupService *services.SetupService,
	systemService *services.SystemService,
) *SystemHandler {
	return &SystemHandler{
		setupService:  setupService,
		systemService: systemService,
	}
}

// GetSystemStatus retrieves the current system status
func (h *SystemHandler) GetSystemStatus(c fiber.Ctx) error {
	status := h.setupService.GetSetupStatus()
	if status.Initialized && status.ZTStatus != nil && !status.ZTStatus.Online {
		logger.Debug("[system status] ZeroTier status check failed or is offline")
	}
	return c.Status(fiber.StatusOK).JSON(status)
}

func (h *SystemHandler) GetRuntimeSettings(c fiber.Ctx) error {
	settings := h.setupService.GetRuntimeSettings()
	return c.Status(fiber.StatusOK).JSON(settings)
}

func (h *SystemHandler) UpdateRuntimeSettings(c fiber.Ctx) error {
	var req services.RuntimeSettings
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind instance settings request", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "system.invalid_request", "Invalid request body")
	}

	if err := h.setupService.UpdateRuntimeSettings(req); err != nil {
		logger.Error("Failed to update instance settings", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	return writeMessageResponse(c, fiber.StatusOK, "system.settings_updated", "Instance settings updated successfully", fiber.Map{"settings": req})
}

// ConfigureDatabase configures the database connection settings
func (h *SystemHandler) ConfigureDatabase(c fiber.Ctx) error {
	var dbConfig models.DatabaseConfig
	if err := c.Bind().Body(&dbConfig); err != nil {
		logger.Error("Failed to bind database configuration request", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "system.invalid_request", "Invalid request body")
	}

	logger.Info("Configuring database", zap.String("type", string(dbConfig.Type)))

	dbCfg, err := h.setupService.ConfigureDatabase(dbConfig)
	if err != nil {
		logger.Error("Database configuration failed", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	logger.Info("Database configured successfully", zap.String("type", string(dbCfg.Type)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Database configured successfully",
		"message_code": "system.database_configured",
		"config": fiber.Map{
			"type": dbCfg.Type,
		},
	})
}

// TestZeroTierConnection tests connectivity to the ZeroTier controller
func (h *SystemHandler) TestZeroTierConnection(c fiber.Ctx) error {
	ztStatus, err := h.setupService.TestZeroTierConnection()
	if err != nil {
		logger.Error("[ZeroTier] connection test failed", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	logger.Info("[ZeroTier] connection test succeeded")
	return c.Status(fiber.StatusOK).JSON(ztStatus)
}

// InitZeroTierClient initializes the ZeroTier client for the application
func (h *SystemHandler) InitZeroTierClient(c fiber.Ctx) error {

	status, err := h.setupService.InitZTClientFromConfig()
	if err != nil {
		logger.Error("ZeroTier client initialization failed", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	return writeMessageResponse(c, fiber.StatusOK, "system.zerotier_initialized", "ZeroTier client initialized successfully", fiber.Map{"status": status})
}

// SaveZeroTierConfig saves ZeroTier configuration and initializes connection
func (h *SystemHandler) SaveZeroTierConfig(c fiber.Ctx) error {
	var req struct {
		ControllerURL string `json:"controllerUrl"`
		TokenPath     string `json:"tokenPath"`
	}

	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind ZeroTier configuration request", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "system.invalid_request", "Invalid request body")
	}

	// Sanitize: log only hostname from URL and whether token path is present
	controllerHost := "invalid"
	if u, err := url.Parse(req.ControllerURL); err == nil && u.Host != "" {
		controllerHost = u.Host
	}
	tokenPresent := req.TokenPath != ""
	logger.Info("Saving ZeroTier configuration", zap.String("controllerHost", controllerHost), zap.Bool("tokenPathPresent", tokenPresent))

	status, err := h.setupService.SaveZeroTierConfig(req.ControllerURL, req.TokenPath)
	if err != nil {
		logger.Error("Failed to save ZeroTier configuration", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	logger.Info("ZeroTier configuration saved and validated")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "ZeroTier configuration saved successfully",
		"message_code": "system.zerotier_configured",
		"config": fiber.Map{
			"controllerUrl": req.ControllerURL,
		},
		"status": status,
	})
}

// InitializeAdminCreation prepares the system for admin account creation
// This function is called when user enters the admin creation step to ensure correct database state
func (h *SystemHandler) InitializeAdminCreation(c fiber.Ctx) error {
	logger.Info("Initializing administrator account creation step")

	databaseType, err := h.setupService.InitializeAdminCreation()
	if err != nil {
		logger.Error("Failed to initialize administrator account creation step", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	logger.Info("Database reset successfully", zap.String("type", databaseType))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Administrator account creation step initialized successfully",
		"message_code": "system.admin_creation_initialized",
		"resetDone":    true,
		"databaseType": databaseType,
	})
}

// SetInitialized updates the system initialization status
func (h *SystemHandler) SetInitialized(c fiber.Ctx) error {
	var req struct {
		Initialized bool `json:"initialized"`
	}

	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind initialization state request", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "system.invalid_request", "Invalid request body")
	}

	logger.Info("Setting system initialization state", zap.Bool("initialized", req.Initialized))

	if err := h.setupService.SetInitialized(req.Initialized); err != nil {
		logger.Error("Failed to set initialization state", zap.Error(err))
		return setupErrorResponse(c, err)
	}

	return writeMessageResponse(c, fiber.StatusOK, "system.initialized_updated", "Initialization state updated successfully", nil)
}

// GetSystemStats retrieves system resource usage statistics
// This endpoint is only accessible to admin users
func (h *SystemHandler) GetSystemStats(c fiber.Ctx) error {
	// Get system stats from service
	stats, err := h.systemService.GetSystemStats()
	if err != nil {
		logger.Error("Failed to get system stats", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusInternalServerError, "system.stats_unavailable", "Unable to retrieve system resource statistics")
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}
