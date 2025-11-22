package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	networkService *services.NetworkService
	userService    *services.UserService
	// 数据库配置将存储在配置文件中
}

// NewSystemHandler 创建新的系统处理器
func NewSystemHandler(networkService *services.NetworkService, userService *services.UserService) *SystemHandler {
	return &SystemHandler{
		networkService: networkService,
		userService:    userService,
	}
}

// GetSystemStatus 获取系统状态
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	logger.Info("开始获取系统状态")

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
		logger.Error("获取ZeroTier状态失败", zap.Error(err))
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

	logger.Info("成功获取系统状态")
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
	
	// 保存数据库配置到环境变量文件
	if err := saveDatabaseConfig(dbConfig); err != nil {
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

// saveDatabaseConfig 保存数据库配置到环境变量文件
func saveDatabaseConfig(config models.DatabaseConfig) error {
	// 读取现有的.env文件内容
	envFile := ".env"
	data, err := os.ReadFile(envFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("读取.env文件失败: %w", err)
	}
	
	// 将内容按行分割
	lines := strings.Split(string(data), "\n")
	
	// 创建一个映射来存储现有的环境变量
	envVars := make(map[string]string)
	for _, line := range lines {
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}
	
	// 更新数据库配置
	envVars["DATABASE_TYPE"] = config.Type
	if config.Path != "" {
		envVars["DATABASE_PATH"] = config.Path
	}
	if config.Host != "" {
		envVars["DATABASE_HOST"] = config.Host
	}
	if config.Port != 0 {
		envVars["DATABASE_PORT"] = strconv.Itoa(config.Port)
	}
	if config.User != "" {
		envVars["DATABASE_USER"] = config.User
	}
	if config.Pass != "" {
		envVars["DATABASE_PASS"] = config.Pass
	}
	if config.Name != "" {
		envVars["DATABASE_NAME"] = config.Name
	}
	
	// 重新构建.env文件内容
	var newLines []string
	for key, value := range envVars {
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}
	
	// 添加一些注释来分隔配置部分
	content := "# Tairitsu Configuration\n"
	for _, line := range newLines {
		content += line + "\n"
	}
	
	// 写入文件
	return os.WriteFile(envFile, []byte(content), 0644)
}