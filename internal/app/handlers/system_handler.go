package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// SystemHandler handles system-related API endpoints and operations
type SystemHandler struct {
	networkService   *services.NetworkService
	userService      *services.UserService
	reloadRoutesFunc func() // Function to reload application routes
	// Database configuration is stored in config file
}

// NewSystemHandler creates a new system handler instance
func NewSystemHandler(networkService *services.NetworkService, userService *services.UserService, reloadRoutesFunc func()) *SystemHandler {
	return &SystemHandler{
		networkService:   networkService,
		userService:      userService,
		reloadRoutesFunc: reloadRoutesFunc,
	}
}

// GetSystemStatus retrieves the current system status
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	// Get system initialization status from config module
	sysConfig := config.AppConfig
	initialized := false

	if sysConfig != nil {
		// Check initialized field from config.json first
		initialized = sysConfig.Initialized

		// If not initialized, return uninitialized status directly
		if !initialized {
			c.JSON(http.StatusOK, map[string]interface{}{
				"initialized": false,
			})
			return
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

	c.JSON(http.StatusOK, response)
}

// ConfigureDatabase configures the database connection settings
func (h *SystemHandler) ConfigureDatabase(c *gin.Context) {
	var dbConfig models.DatabaseConfig
	if err := c.ShouldBindJSON(&dbConfig); err != nil {
		logger.Error("数据库配置参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("开始配置数据库", zap.String("type", dbConfig.Type))

	// Validate database configuration
	dbCfg := database.Config{
		Type: database.DatabaseType(dbConfig.Type),
		Path: dbConfig.Path,
		Host: dbConfig.Host,
		Port: dbConfig.Port,
		User: dbConfig.User,
		Pass: dbConfig.Pass,
		Name: dbConfig.Name,
	}

	// Try connecting to database
	db, err := database.NewDatabase(dbCfg)
	if err != nil {
		logger.Error("数据库连接失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接失败: " + err.Error()})
		return
	}

	// Initialize database schema
	if err := db.Init(); err != nil {
		logger.Error("数据库初始化失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库初始化失败: " + err.Error()})
		return
	}

	// Close database connection
	if err := db.Close(); err != nil {
		logger.Warn("关闭数据库连接时出现警告", zap.Error(err))
	}

	// For SQLite, ensure path is properly saved to config
	// NewDatabase function might set default path if Path is empty
	if dbCfg.Type == database.SQLite {
		// From factory.go we know if Path is empty, default value "tairitsu.db" will be used
		if dbConfig.Path == "" {
			dbCfg.Path = "tairitsu.db"
		}
		logger.Info("SQLite数据库路径已设置", zap.String("path", dbCfg.Path))
	}

	// Save database configuration to unified config management module
	if err := database.SaveConfig(dbCfg); err != nil {
		logger.Error("保存数据库配置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存数据库配置失败: " + err.Error()})
		return
	}

	logger.Info("数据库配置成功", zap.String("type", dbConfig.Type))
	c.JSON(http.StatusOK, gin.H{
		"message": "数据库配置成功",
		"config":  dbConfig,
	})
}

// TestZeroTierConnection tests connectivity to the ZeroTier controller
func (h *SystemHandler) TestZeroTierConnection(c *gin.Context) {
	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("[ZeroTier] 连接测试失败: 创建客户端失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建ZeroTier客户端失败: " + err.Error()})
		return
	}

	// 获取ZeroTier控制器状态
	ztStatus, err := ztClient.GetStatus()
	if err != nil {
		logger.Error("[ZeroTier] 连接测试失败: 获取状态失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接到ZeroTier控制器: " + err.Error()})
		return
	}

	logger.Info("[ZeroTier] 连接测试成功")
	c.JSON(http.StatusOK, ztStatus)
}

// InitZeroTierClient initializes the ZeroTier client for the application
func (h *SystemHandler) InitZeroTierClient(c *gin.Context) {

	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("创建ZeroTier客户端失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建ZeroTier客户端失败: " + err.Error()})
		return
	}

	// Set client in network service
	h.networkService.SetZTClient(ztClient)

	// 验证客户端是否正常工作
	status, err := h.networkService.GetStatus()
	if err != nil {
		logger.Error("ZeroTier客户端验证失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ZeroTier客户端初始化后验证失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ZeroTier客户端初始化成功", "status": status})
}

// SaveZeroTierConfig saves ZeroTier configuration and initializes connection
func (h *SystemHandler) SaveZeroTierConfig(c *gin.Context) {
	var req struct {
		ControllerURL string `json:"controllerUrl" binding:"required"`
		TokenPath     string `json:"tokenPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("ZeroTier配置参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("保存ZeroTier配置", zap.String("controllerUrl", req.ControllerURL), zap.String("tokenPath", req.TokenPath))

	// Call config module to save ZeroTier configuration
	if err := config.SetZTConfig(req.ControllerURL, req.TokenPath); err != nil {
		logger.Error("保存ZeroTier配置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存ZeroTier配置失败: " + err.Error()})
		return
	}

	// Try creating and validating ZeroTier client with new config
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("创建ZeroTier客户端失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建ZeroTier客户端失败: " + err.Error()})
		return
	}

	// Validate client works and get status information
	status, err := ztClient.GetStatus()
	if err != nil {
		logger.Error("ZeroTier客户端验证失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ZeroTier客户端验证失败: " + err.Error()})
		return
	}

	// Set client in network service
	h.networkService.SetZTClient(ztClient)

	logger.Info("ZeroTier配置保存并验证成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "ZeroTier配置保存成功",
		"config":  req,
		"status":  status,
	})
}

// InitializeAdminCreation prepares the system for admin account creation
// This function is called when user enters the admin creation step to ensure correct database state
func (h *SystemHandler) InitializeAdminCreation(c *gin.Context) {
	logger.Info("初始化管理员账户创建步骤")

	// Check if reset operation has already been performed
	resetDoneKey := "admin_creation_reset_done"
	resetDone := config.GetTempSetting(resetDoneKey)

	// Skip reset if already performed
	if resetDone == "true" {
		logger.Info("数据库重置已在之前执行，跳过重置操作")
		c.JSON(http.StatusOK, gin.H{
			"message":   "管理员账户创建步骤已初始化",
			"resetDone": true,
		})
		return
	}

	// Get current database configuration
	dbConfig := database.LoadConfig()
	logger.Info("获取数据库配置", zap.String("type", string(dbConfig.Type)))

	// Only perform reset for SQLite databases
	if dbConfig.Type == database.SQLite {
		logger.Info("准备重置SQLite数据库")

		// Execute database reset
		if err := database.ResetDatabase(dbConfig); err != nil {
			logger.Error("数据库重置失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化管理员账户创建步骤失败: " + err.Error()})
			return
		}

		logger.Info("SQLite数据库重置成功")
	} else {
		logger.Info("当前数据库类型不是SQLite，跳过数据库重置", zap.String("type", string(dbConfig.Type)))
	}

	// Set flag indicating reset operation is complete
	config.SetTempSetting(resetDoneKey, "true")
	logger.Info("设置重置操作完成标志")

	c.JSON(http.StatusOK, gin.H{
		"message":      "管理员账户创建步骤初始化成功",
		"resetDone":    true,
		"databaseType": string(dbConfig.Type),
	})
}

// SetInitialized updates the system initialization status
func (h *SystemHandler) SetInitialized(c *gin.Context) {
	var req struct {
		Initialized bool `json:"initialized" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("设置初始化状态参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("设置系统初始化状态", zap.Bool("initialized", req.Initialized))

	// Call config module to set initialization status
	if err := config.SetInitialized(req.Initialized); err != nil {
		logger.Error("设置初始化状态失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置初始化状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "初始化状态更新成功"})
}

// ReloadRoutes handles API request to reload application routes
func (h *SystemHandler) ReloadRoutes(c *gin.Context) {

	// Call internal reload function
	if h.reloadRoutesFunc != nil {
		h.reloadRoutesFunc()
		c.JSON(http.StatusOK, gin.H{"message": "路由重新加载成功"})
	} else {
		logger.Warn("重新加载路由函数未定义")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新加载路由功能不可用"})
	}
}
