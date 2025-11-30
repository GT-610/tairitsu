package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// NetworkHandler Network handler
type NetworkHandler struct {
	networkService *services.NetworkService
}

// NewNetworkHandler Create network handler instance
func NewNetworkHandler(networkService *services.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// GetStatus Get ZeroTier network status
func (h *NetworkHandler) GetStatus(c *gin.Context) {
	logger.Info("Starting to get ZeroTier network status")
	
	status, err := h.networkService.GetStatus()
	if err != nil {
		logger.Error("Failed to get ZeroTier network status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Successfully got ZeroTier network status")

	c.JSON(http.StatusOK, status)
}

// GetNetworks Get all networks for current user
func (h *NetworkHandler) GetNetworks(c *gin.Context) {
	logger.Info("Starting to get all networks list for current user")
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("Failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}
	
	networks, err := h.networkService.GetAllNetworks(userID.(string))
	if err != nil {
		logger.Error("Failed to get networks list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Successfully got networks list", zap.Int("network_count", len(networks)))

	c.JSON(http.StatusOK, networks)
}

// GetNetwork Get specific network
func (h *NetworkHandler) GetNetwork(c *gin.Context) {
	id := c.Param("id")
	logger.Info("Starting to get specific network", zap.String("network_id", id))
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("获取用户ID失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}
	
	network, err := h.networkService.GetNetworkByID(id, userID.(string))
	if err != nil {
		logger.Error("Failed to get network", zap.String("network_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if network == nil {
		logger.Warn("Network does not exist or no permission to access", zap.String("network_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": "Network does not exist or no permission to access"})
		return
	}
	
	logger.Info("Successfully got network", zap.String("network_id", id))

	c.JSON(http.StatusOK, network)
}

// CreateNetwork Create new network
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var req zerotier.Network
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind create network request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Starting to create new network", zap.String("network_name", req.Name))

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("获取用户ID失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	network, err := h.networkService.CreateNetwork(&req, userID.(string))
	if err != nil {
		logger.Error("Failed to create network", zap.String("network_name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Successfully created network", zap.String("network_id", network.ID), zap.String("network_name", network.Name))

	c.JSON(http.StatusCreated, network)
}

// UpdateNetwork Update network
func (h *NetworkHandler) UpdateNetwork(c *gin.Context) {
	id := c.Param("id")
	
	var req zerotier.Network
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind update network request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Starting to update network", zap.String("network_id", id))

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("获取用户ID失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	network, err := h.networkService.UpdateNetwork(id, &req, userID.(string))
	if err != nil {
		logger.Error("Failed to update network", zap.String("network_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Successfully updated network", zap.String("network_id", network.ID))

	c.JSON(http.StatusOK, network)
}

// DeleteNetwork Delete network
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	id := c.Param("id")
	logger.Info("Starting to delete network", zap.String("network_id", id))
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("获取用户ID失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}
	
	err := h.networkService.DeleteNetwork(id, userID.(string))
	if err != nil {
		logger.Error("Failed to delete network", zap.String("network_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("Successfully deleted network", zap.String("network_id", id))

	c.JSON(http.StatusOK, gin.H{"message": "Network deleted successfully"})
}