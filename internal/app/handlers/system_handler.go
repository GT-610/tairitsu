package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	networkService *services.NetworkService
	userService    *services.UserService
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