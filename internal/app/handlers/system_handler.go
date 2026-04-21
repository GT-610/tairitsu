package handlers

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
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SystemHandler handles system-related API endpoints and operations
type SystemHandler struct {
	networkService *services.NetworkService
	userService    *services.UserService
	runtimeService *services.RuntimeService
	systemService  *services.SystemService
	// Database configuration is stored in config file
}

// NewSystemHandler creates a new system handler instance
func NewSystemHandler(networkService *services.NetworkService, userService *services.UserService) *SystemHandler {
	return &SystemHandler{
		networkService: networkService,
		userService:    userService,
		runtimeService: services.NewRuntimeService(userService, networkService),
		systemService:  services.NewSystemService(),
	}
}

// GetSystemStatus retrieves the current system status
func (h *SystemHandler) GetSystemStatus(c fiber.Ctx) error {
	// Get system initialization status from config module
	sysConfig := config.AppConfig
	initialized := false

	if sysConfig != nil {
		// Check initialized field from config.json first
		initialized = sysConfig.Initialized

		// If not initialized, return uninitialized status directly
		if !initialized {
			return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
				"initialized": false,
			})
		}
	}

	// If initialized, retrieve additional status information
	hasDatabase := false
	if sysConfig != nil && initialized {
		// Check if database is configured
		hasDatabase = sysConfig.Database.Type != ""
	}

	// Check for admin user only if database is configured
	hasAdmin := false
	if hasDatabase && h.userService != nil {
		users := h.userService.GetAllUsers()
		for _, user := range users {
			if user.Role == "admin" {
				hasAdmin = true
				break
			}
		}
	}

	// Check ZeroTier connection status
	ztStatus, err := h.networkService.GetStatus()
	if err != nil {
		logger.Debug("[系统状态] ZeroTier状态检查失败", zap.Error(err))
		// Return system status even if ZeroTier connection fails
		ztStatus = &zerotier.Status{
			Version: "unknown",
			Address: "",
			Online:  false,
		}
	}

	// 当已初始化时，返回完整的状态信息
	response := map[string]interface{}{
		"initialized": true,
		"hasDatabase": hasDatabase,
		"hasAdmin":    hasAdmin,
		"ztStatus":    ztStatus,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ConfigureDatabase configures the database connection settings
func (h *SystemHandler) ConfigureDatabase(c fiber.Ctx) error {
	var dbConfig models.DatabaseConfig
	if err := c.Bind().Body(&dbConfig); err != nil {
		logger.Error("数据库配置参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("开始配置数据库", zap.String("type", dbConfig.Type))

	if dbConfig.Type != string(database.SQLite) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "当前一期版本仅正式支持 SQLite，请使用 SQLite 完成初始化",
		})
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
		logger.Error("数据库连接失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "数据库连接失败: " + err.Error()})
	}

	if err := db.Init(); err != nil {
		logger.Error("数据库初始化失败", zap.Error(err))
		db.Close()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "数据库初始化失败: " + err.Error()})
	}

	if dbCfg.Type == database.SQLite {
		if dbConfig.Path == "" {
			dbCfg.Path = "data/tairitsu.db"
		}
		logger.Info("SQLite数据库路径已设置", zap.String("path", dbCfg.Path))
	}

	if err := database.SaveConfig(dbCfg); err != nil {
		logger.Error("保存数据库配置失败", zap.Error(err))
		db.Close()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "保存数据库配置失败: " + err.Error()})
	}

	h.runtimeService.BindDatabase(db)

	logger.Info("数据库配置成功", zap.String("type", dbConfig.Type))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "数据库配置成功",
		"config":  dbConfig,
	})
}

// TestZeroTierConnection tests connectivity to the ZeroTier controller
func (h *SystemHandler) TestZeroTierConnection(c fiber.Ctx) error {
	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("[ZeroTier] 连接测试失败: 创建客户端失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "创建ZeroTier客户端失败: " + err.Error()})
	}

	// 获取ZeroTier控制器状态
	ztStatus, err := ztClient.GetStatus()
	if err != nil {
		logger.Error("[ZeroTier] 连接测试失败: 获取状态失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "无法连接到ZeroTier控制器: " + err.Error()})
	}

	logger.Info("[ZeroTier] 连接测试成功")
	return c.Status(fiber.StatusOK).JSON(ztStatus)
}

// InitZeroTierClient initializes the ZeroTier client for the application
func (h *SystemHandler) InitZeroTierClient(c fiber.Ctx) error {

	status, err := h.runtimeService.InitZTClientFromConfig()
	if err != nil {
		logger.Error("ZeroTier客户端初始化失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("保存ZeroTier配置", zap.String("controllerUrl", req.ControllerURL), zap.String("tokenPath", req.TokenPath))

	// Call config module to save ZeroTier configuration
	if err := config.SetZTConfig(req.ControllerURL, req.TokenPath); err != nil {
		logger.Error("保存ZeroTier配置失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "保存ZeroTier配置失败: " + err.Error()})
	}

	// Try creating and validating ZeroTier client with new config
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("创建ZeroTier客户端失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "创建ZeroTier客户端失败: " + err.Error()})
	}

	// Validate client works and get status information
	status, err := ztClient.GetStatus()
	if err != nil {
		logger.Error("ZeroTier客户端验证失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ZeroTier客户端验证失败: " + err.Error()})
	}

	h.runtimeService.BindZTClient(ztClient)

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

	// Check if reset operation has already been performed
	resetDoneKey := "admin_creation_reset_done"
	resetDone := config.GetTempSetting(resetDoneKey)

	// Skip reset if already performed
	if resetDone == "true" {
		logger.Info("数据库重置已在之前执行，跳过重置操作")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":   "管理员账户创建步骤已初始化",
			"resetDone": true,
		})
	}

	// Get current database configuration
	dbConfig := database.LoadConfig()
	logger.Info("获取数据库配置", zap.String("type", string(dbConfig.Type)))

	if dbConfig.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "尚未完成数据库配置，请先配置 SQLite 数据库",
		})
	}

	if dbConfig.Type != database.SQLite {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("当前一期版本仅正式支持 SQLite，%s 初始化暂不支持", dbConfig.Type),
		})
	}

	logger.Info("准备重置SQLite数据库")

	h.runtimeService.CloseCurrentDatabase()

	// Execute database reset
	if err := database.ResetDatabase(dbConfig); err != nil {
		logger.Error("数据库重置失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "初始化管理员账户创建步骤失败: " + err.Error()})
	}

	if err := h.runtimeService.ReopenConfiguredDatabase(); err != nil {
		logger.Error("重置后重新打开数据库失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "数据库重置后重新初始化失败: " + err.Error()})
	}

	logger.Info("SQLite数据库重置成功")

	// Set flag indicating reset operation is complete
	config.SetTempSetting(resetDoneKey, "true")
	logger.Info("设置重置操作完成标志")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "管理员账户创建步骤初始化成功",
		"resetDone":    true,
		"databaseType": string(dbConfig.Type),
	})
}

// SetInitialized updates the system initialization status
func (h *SystemHandler) SetInitialized(c fiber.Ctx) error {
	var req struct {
		Initialized bool `json:"initialized"`
	}

	if err := c.Bind().Body(&req); err != nil {
		logger.Error("设置初始化状态参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("设置系统初始化状态", zap.Bool("initialized", req.Initialized))

	// Generate JWT and session secrets if not already set
	if req.Initialized {
		logger.Info("开始生成安全密钥")

		// Generate random JWT secret if not already set
		if config.AppConfig.Security.JWTSecret == "" {
			// Generate 32-byte random secret
			jwtSecret := generateRandomSecret(32)
			config.AppConfig.Security.JWTSecret = jwtSecret
			logger.Info("生成新的JWT密钥")
		}

		// Generate random session secret if not already set
		if config.AppConfig.Security.SessionSecret == "" {
			// Generate 32-byte random secret
			sessionSecret := generateRandomSecret(32)
			config.AppConfig.Security.SessionSecret = sessionSecret
			logger.Info("生成新的会话密钥")
		}

		// Save updated configuration with new secrets
		if err := config.SaveConfig(config.AppConfig); err != nil {
			logger.Error("保存密钥配置失败", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "生成安全密钥失败: " + err.Error()})
		}
	}

	// Call config module to set initialization status
	if err := config.SetInitialized(req.Initialized); err != nil {
		logger.Error("设置初始化状态失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "设置初始化状态失败: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "初始化状态更新成功"})
}

// generateRandomSecret generates a random secret string of specified length
func generateRandomSecret(length int) string {
	// Use crypto/rand to generate secure random bytes
	// This is a simple implementation, in production you might want to use a more robust method
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a less secure method if crypto/rand fails
		logger.Warn("使用crypto/rand生成随机密钥失败，将使用math/rand作为备选方案", zap.Error(err))

		// Use math/rand as fallback
		r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		for i := range bytes {
			bytes[i] = byte(r.Intn(256))
		}
	}

	// Encode to base64 for a safe string representation
	return base64.URLEncoding.EncodeToString(bytes)
}

// ReloadRoutes handles API request to reload application routes
func (h *SystemHandler) ReloadRoutes(c fiber.Ctx) error {
	logger.Info("重新加载路由接口已弃用")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "重新加载路由接口已弃用；当前版本会在现有服务实例上直接刷新依赖",
	})
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
