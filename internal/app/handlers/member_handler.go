package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// MemberHandler 成员处理器
type MemberHandler struct {
	networkService *services.NetworkService
}

// NewMemberHandler 创建成员处理器实例
func NewMemberHandler(networkService *services.NetworkService) *MemberHandler {
	return &MemberHandler{
		networkService: networkService,
	}
}

// GetMembers 获取网络中的所有成员
func (h *MemberHandler) GetMembers(c *gin.Context) {
	networkID := c.Param("networkId")
	logger.Info("开始获取网络中的所有成员", zap.String("network_id", networkID))
	
	members, err := h.networkService.GetNetworkMembers(networkID)
	if err != nil {
		logger.Error("获取网络成员列表失败", zap.String("network_id", networkID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功获取网络成员列表", zap.String("network_id", networkID), zap.Int("member_count", len(members)))

	c.JSON(http.StatusOK, members)
}

// GetMember 获取网络中的特定成员
func (h *MemberHandler) GetMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")
	logger.Info("开始获取网络中的特定成员", zap.String("network_id", networkID), zap.String("member_id", memberID))
	
	member, err := h.networkService.GetNetworkMember(networkID, memberID)
	if err != nil {
		logger.Error("获取网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if member == nil {
		logger.Warn("网络成员不存在", zap.String("network_id", networkID), zap.String("member_id", memberID))
		c.JSON(http.StatusNotFound, gin.H{"error": "成员不存在"})
		return
	}
	
	logger.Info("成功获取网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	c.JSON(http.StatusOK, member)
}

// UpdateMember 更新网络成员
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")
	
	var req zerotier.Member
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("更新网络成员请求参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("开始更新网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	member, err := h.networkService.UpdateNetworkMember(networkID, memberID, &req)
	if err != nil {
		logger.Error("更新网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功更新网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	c.JSON(http.StatusOK, member)
}

// DeleteMember 删除网络成员
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")
	logger.Info("开始删除网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))
	
	err := h.networkService.RemoveNetworkMember(networkID, memberID)
	if err != nil {
		logger.Error("删除网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("成功删除网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	c.JSON(http.StatusOK, gin.H{"message": "成员删除成功"})
}