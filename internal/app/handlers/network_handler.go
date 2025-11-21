package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
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

// GetStatus 获取ZeroTier状态
func (h *NetworkHandler) GetStatus(c *gin.Context) {
	status, err := h.networkService.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetNetworks 获取所有网络
func (h *NetworkHandler) GetNetworks(c *gin.Context) {
	networks, err := h.networkService.GetAllNetworks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, networks)
}

// GetNetwork 获取单个网络
func (h *NetworkHandler) GetNetwork(c *gin.Context) {
	networkID := c.Param("id")
	network, err := h.networkService.GetNetworkByID(networkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, network)
}

// CreateNetwork 创建网络
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var network zerotier.Network
	if err := c.ShouldBindJSON(\u0026network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdNetwork, err := h.networkService.CreateNetwork(\u0026network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdNetwork)
}

// UpdateNetwork 更新网络
func (h *NetworkHandler) UpdateNetwork(c *gin.Context) {
	networkID := c.Param("id")
	var network zerotier.Network
	if err := c.ShouldBindJSON(\u0026network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedNetwork, err := h.networkService.UpdateNetwork(networkID, \u0026network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedNetwork)
}

// DeleteNetwork 删除网络
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	networkID := c.Param("id")
	if err := h.networkService.DeleteNetwork(networkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}