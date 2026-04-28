package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

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
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := h.setupService.UpdateRuntimeSettings(req); err != nil {
		logger.Error("Failed to update instance settings", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Instance settings updated successfully",
		"message_code": "system.settings_updated",
		"settings":     req,
	})
}

// ConfigureDatabase configures the database connection settings
func (h *SystemHandler) ConfigureDatabase(c fiber.Ctx) error {
	var dbConfig models.DatabaseConfig
	if err := c.Bind().Body(&dbConfig); err != nil {
		logger.Error("Failed to bind database configuration request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Configuring database", zap.String("type", dbConfig.Type))

	dbCfg, err := h.setupService.ConfigureDatabase(dbConfig)
	if err != nil {
		logger.Error("Database configuration failed", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("SQLite database path set", zap.String("path", dbCfg.Path))

	logger.Info("Database configured successfully", zap.String("type", dbConfig.Type))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Database configured successfully",
		"message_code": "system.database_configured",
		"config": fiber.Map{
			"type": dbCfg.Type,
			"path": dbCfg.Path,
		},
	})
}

// TestZeroTierConnection tests connectivity to the ZeroTier controller
func (h *SystemHandler) TestZeroTierConnection(c fiber.Ctx) error {
	ztStatus, err := h.setupService.TestZeroTierConnection()
	if err != nil {
		logger.Error("[ZeroTier] connection test failed", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("[ZeroTier] connection test succeeded")
	return c.Status(fiber.StatusOK).JSON(ztStatus)
}

// InitZeroTierClient initializes the ZeroTier client for the application
func (h *SystemHandler) InitZeroTierClient(c fiber.Ctx) error {

	status, err := h.setupService.InitZTClientFromConfig()
	if err != nil {
		logger.Error("ZeroTier client initialization failed", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
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
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Saving ZeroTier configuration", zap.String("controllerUrl", req.ControllerURL), zap.String("tokenPath", req.TokenPath))

	status, err := h.setupService.SaveZeroTierConfig(req.ControllerURL, req.TokenPath)
	if err != nil {
		logger.Error("Failed to save ZeroTier configuration", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("ZeroTier configuration saved and validated")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "ZeroTier configuration saved successfully",
		"message_code": "system.zerotier_configured",
		"config":       req,
		"status":       status,
	})
}

// InitializeAdminCreation prepares the system for admin account creation
// This function is called when user enters the admin creation step to ensure correct database state
func (h *SystemHandler) InitializeAdminCreation(c fiber.Ctx) error {
	logger.Info("Initializing administrator account creation step")

	databaseType, err := h.setupService.InitializeAdminCreation()
	if err != nil {
		logger.Error("Failed to initialize administrator account creation step", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("SQLite database reset successfully")

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
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Setting system initialization state", zap.Bool("initialized", req.Initialized))

	if err := h.setupService.SetInitialized(req.Initialized); err != nil {
		logger.Error("Failed to set initialization state", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":      "Unable to retrieve system resource statistics",
			"message":    "Unable to retrieve system resource statistics",
			"error_code": "system.stats_unavailable",
			"details":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}
