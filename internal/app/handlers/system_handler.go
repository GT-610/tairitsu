package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/models"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	networkService   *services.NetworkService
	userService      *services.UserService
	reloadRoutesFunc func() // 添加重新加载路由的函数
	// 数据库配置将存储在配置文件中
}

// NewSystemHandler 创建新的系统处理器
func NewSystemHandler(networkService *services.NetworkService, userService *services.UserService, reloadRoutesFunc func()) *SystemHandler {
	return &SystemHandler{
		networkService:   networkService,
		userService:      userService,
		reloadRoutesFunc: reloadRoutesFunc,
	}
}

// GetSystemStatus 获取系统状态
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	// 检查是否有管理员用户
	users := h.userService.GetAllUsers()
	hasAdmin := false
	for _, user := range users {
		if user.Role == "admin" {
			hasAdmin = true
			break
		}
	}

	// 检查ZeroTier连接状态
	ztStatus, err := h.networkService.GetStatus()
	if err != nil {
		logger.Debug("[系统状态] ZeroTier状态检查失败", zap.Error(err))
		// 即使ZeroTier连接失败，也返回系统状态
		ztStatus = &zerotier.Status{
			Version: "unknown",
			Address: "",
			Online:  false,
		}
	}

	response := map[string]interface{}{
		"hasAdmin": hasAdmin,
		"ztStatus": ztStatus,
	}

	c.JSON(http.StatusOK, response)
}

// ConfigureDatabase 配置数据库
func (h *SystemHandler) ConfigureDatabase(c *gin.Context) {
	var dbConfig models.DatabaseConfig
	if err := c.ShouldBindJSON(&dbConfig); err != nil {
		logger.Error("数据库配置参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("开始配置数据库", zap.String("type", dbConfig.Type))

	// 验证数据库配置
	dbCfg := database.Config{
		Type: database.DatabaseType(dbConfig.Type),
		Path: dbConfig.Path,
		Host: dbConfig.Host,
		Port: dbConfig.Port,
		User: dbConfig.User,
		Pass: dbConfig.Pass,
		Name: dbConfig.Name,
	}

	// 尝试连接数据库
	db, err := database.NewDatabase(dbCfg)
	if err != nil {
		logger.Error("数据库连接失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接失败: " + err.Error()})
		return
	}

	// 初始化数据库
	if err := db.Init(); err != nil {
		logger.Error("数据库初始化失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库初始化失败: " + err.Error()})
		return
	}

	// 关闭数据库连接
	if err := db.Close(); err != nil {
		logger.Warn("关闭数据库连接时出现警告", zap.Error(err))
	}

	// 对于SQLite，确保路径被正确保存到配置中
	// NewDatabase函数可能会在Path为空时设置默认路径
	if dbCfg.Type == database.SQLite {
		// 从factory.go中我们知道，如果Path为空，会使用默认值"tairitsu.db"
		if dbConfig.Path == "" {
			dbCfg.Path = "tairitsu.db"
		}
		logger.Info("SQLite数据库路径已设置", zap.String("path", dbCfg.Path))
	}

	// 保存数据库配置到JSON文件
	// 注意：os.WriteFile默认会覆盖文件，满足强制覆盖的需求
	if err := database.SaveConfigToJSON(dbCfg); err != nil {
		logger.Error("保存数据库配置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存数据库配置失败: " + err.Error()})
		return
	}

	// 重新加载路由以注册认证端点
	if h.reloadRoutesFunc != nil {
		h.reloadRoutesFunc()
	}

	logger.Info("数据库配置成功", zap.String("type", dbConfig.Type))
	c.JSON(http.StatusOK, gin.H{
		"message": "数据库配置成功",
		"config":  dbConfig,
	})
}

// TestZeroTierConnection 测试ZeroTier连接
func (h *SystemHandler) TestZeroTierConnection(c *gin.Context) {
	// 动态创建ZeroTier客户端
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

// InitZeroTierClient 初始化ZeroTier客户端
func (h *SystemHandler) InitZeroTierClient(c *gin.Context) {
	logger.Info("[系统初始化] 开始初始化ZeroTier客户端")

	// 动态创建ZeroTier客户端
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("[ZeroTier] 创建客户端失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建ZeroTier客户端失败: " + err.Error()})
		return
	}

	// 将客户端设置到网络服务中
	h.networkService.SetZTClient(ztClient)

	// 验证客户端是否正常工作
	_, err = h.networkService.GetStatus()
	if err != nil {
		logger.Error("[ZeroTier] 客户端验证失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ZeroTier客户端初始化后验证失败: " + err.Error()})
		return
	}

	logger.Info("[ZeroTier] 客户端初始化并验证成功")
	c.JSON(http.StatusOK, gin.H{"message": "ZeroTier客户端初始化成功"})
}

// 数据库配置相关函数已迁移到JSON文件存储方式
