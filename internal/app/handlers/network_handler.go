package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// NetworkHandler 网络处理器
type NetworkHandler struct {
	networkService *services.NetworkService
}

// NewNetworkHandler 创建网络处理器实例
func NewNetworkHandler(networkService *services.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// GetStatus 获取ZeroTier网络状态
func (h *NetworkHandler) GetStatus(c *gin.Context) {
	logger.Info("开始获取ZeroTier网络状态")
	
	status, err := h.networkService.GetStatus()
	if err != nil {
		logger.Error("获取ZeroTier网络状态失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功获取ZeroTier网络状态")

	c.JSON(http.StatusOK, status)
}

// GetNetworks 获取所有网络
func (h *NetworkHandler) GetNetworks(c *gin.Context) {
	logger.Info("开始获取所有网络列表")
	
	networks, err := h.networkService.GetAllNetworks()
	if err != nil {
		logger.Error("获取网络列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功获取网络列表", zap.Int("network_count", len(networks)))

	c.JSON(http.StatusOK, networks)
}

// GetNetwork 获取特定网络
func (h *NetworkHandler) GetNetwork(c *gin.Context) {
	id := c.Param("id")
	logger.Info("开始获取特定网络", zap.String("network_id", id))
	
	network, err := h.networkService.GetNetworkByID(id)
	if err != nil {
		logger.Error("获取网络失败", zap.String("network_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if network == nil {
		logger.Warn("网络不存在", zap.String("network_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": "网络不存在"})
		return
	}
	
	logger.Info("成功获取网络", zap.String("network_id", id))

	c.JSON(http.StatusOK, network)
}

// CreateNetwork 创建新网络
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var req zerotier.Network
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("创建网络请求参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("开始创建新网络", zap.String("network_name", req.Name))

	network, err := h.networkService.CreateNetwork(&req)
	if err != nil {
		logger.Error("创建网络失败", zap.String("network_name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功创建网络", zap.String("network_id", network.ID), zap.String("network_name", network.Name))

	c.JSON(http.StatusCreated, network)
}

// UpdateNetwork 更新网络
func (h *NetworkHandler) UpdateNetwork(c *gin.Context) {
	id := c.Param("id")
	
	var req zerotier.Network
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("更新网络请求参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("开始更新网络", zap.String("network_id", id))

	network, err := h.networkService.UpdateNetwork(id, &req)
	if err != nil {
		logger.Error("更新网络失败", zap.String("network_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功更新网络", zap.String("network_id", network.ID))

	c.JSON(http.StatusOK, network)
}

// DeleteNetwork 删除网络
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	id := c.Param("id")
	logger.Info("开始删除网络", zap.String("network_id", id))
	
	err := h.networkService.DeleteNetwork(id)
	if err != nil {
		logger.Error("删除网络失败", zap.String("network_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功删除网络", zap.String("network_id", id))

	c.JSON(http.StatusOK, gin.H{"message": "网络删除成功"})
}