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
		logger.Debug("[系统状态] ZeroTier状态检查失败或离线")
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
		logger.Error("实例设置参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := h.setupService.UpdateRuntimeSettings(req); err != nil {
		logger.Error("更新实例设置失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "实例设置更新成功",
		"settings": req,
	})
}

// ConfigureDatabase configures the database connection settings
func (h *SystemHandler) ConfigureDatabase(c fiber.Ctx) error {
	var dbConfig models.DatabaseConfig
	if err := c.Bind().Body(&dbConfig); err != nil {
		logger.Error("数据库配置参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("开始配置数据库", zap.String("type", dbConfig.Type))

	dbCfg, err := h.setupService.ConfigureDatabase(dbConfig)
	if err != nil {
		logger.Error("数据库配置失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("SQLite数据库路径已设置", zap.String("path", dbCfg.Path))

	logger.Info("数据库配置成功", zap.String("type", dbConfig.Type))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "数据库配置成功",
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
		logger.Error("[ZeroTier] 连接测试失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("[ZeroTier] 连接测试成功")
	return c.Status(fiber.StatusOK).JSON(ztStatus)
}

// InitZeroTierClient initializes the ZeroTier client for the application
func (h *SystemHandler) InitZeroTierClient(c fiber.Ctx) error {

	status, err := h.setupService.InitZTClientFromConfig()
	if err != nil {
		logger.Error("ZeroTier客户端初始化失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "ZeroTier客户端初始化成功", "status": status})
}

// SaveZeroTierConfig saves ZeroTier configuration and initializes connection
func (h *SystemHandler) SaveZeroTierConfig(c fiber.Ctx) error {
	var req struct {
		ControllerURL string `json:"controllerUrl"`
		TokenPath     string `json:"tokenPath"`
	}

	if err := c.Bind().Body(&req); err != nil {
		logger.Error("ZeroTier配置参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("保存ZeroTier配置", zap.String("controllerUrl", req.ControllerURL), zap.String("tokenPath", req.TokenPath))

	status, err := h.setupService.SaveZeroTierConfig(req.ControllerURL, req.TokenPath)
	if err != nil {
		logger.Error("保存ZeroTier配置失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("ZeroTier配置保存并验证成功")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ZeroTier配置保存成功",
		"config":  req,
		"status":  status,
	})
}

// InitializeAdminCreation prepares the system for admin account creation
// This function is called when user enters the admin creation step to ensure correct database state
func (h *SystemHandler) InitializeAdminCreation(c fiber.Ctx) error {
	logger.Info("初始化管理员账户创建步骤")

	databaseType, err := h.setupService.InitializeAdminCreation()
	if err != nil {
		logger.Error("初始化管理员账户创建步骤失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("SQLite数据库重置成功")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "管理员账户创建步骤初始化成功",
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
		logger.Error("设置初始化状态参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("设置系统初始化状态", zap.Bool("initialized", req.Initialized))

	if err := h.setupService.SetInitialized(req.Initialized); err != nil {
		logger.Error("设置初始化状态失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "初始化状态更新成功"})
}

// GetSystemStats retrieves system resource usage statistics
// This endpoint is only accessible to admin users
func (h *SystemHandler) GetSystemStats(c fiber.Ctx) error {
	// Get system stats from service
	stats, err := h.systemService.GetSystemStats()
	if err != nil {
		logger.Error("Failed to get system stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "无法获取系统资源统计信息",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}
