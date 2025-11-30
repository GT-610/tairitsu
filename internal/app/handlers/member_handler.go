package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// MemberHandler Member handler
type MemberHandler struct {
	networkService *services.NetworkService
}

// NewMemberHandler Create member handler instance
func NewMemberHandler(networkService *services.NetworkService) *MemberHandler {
	return &MemberHandler{
		networkService: networkService,
	}
}

// GetMembers Get all members in network
func (h *MemberHandler) GetMembers(c *gin.Context) {
	networkID := c.Param("networkId")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("Failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	members, err := h.networkService.GetNetworkMembers(networkID, userID.(string))
	if err != nil {
		logger.Error("Failed to get network members list", zap.String("network_id", networkID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}

// GetMember Get specific member in network
func (h *MemberHandler) GetMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("Failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	member, err := h.networkService.GetNetworkMember(networkID, memberID, userID.(string))
	if err != nil {
		logger.Error("Failed to get network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get member: " + err.Error()})
		return
	}

	if member == nil {
		logger.Warn("Network member does not exist", zap.String("network_id", networkID), zap.String("member_id", memberID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Member does not exist"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// UpdateMember Update network member
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")

	var req zerotier.Member
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request parameters", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("Failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	member, err := h.networkService.UpdateNetworkMember(networkID, memberID, &req, userID.(string))
	if err != nil {
		logger.Error("Failed to update network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

// DeleteMember Delete network member
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("Failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	err := h.networkService.RemoveNetworkMember(networkID, memberID, userID.(string))
	if err != nil {
		logger.Error("Failed to delete network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete member: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member deleted successfully"})
}
