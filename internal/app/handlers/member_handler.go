package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
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

// GetMembers 获取网络成员列表
func (h *MemberHandler) GetMembers(c *gin.Context) {
	networkID := c.Param("networkId")
	members, err := h.networkService.GetNetworkMembers(networkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

// GetMember 获取单个成员
func (h *MemberHandler) GetMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")
	member, err := h.networkService.GetNetworkMember(networkID, memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, member)
}

// UpdateMember 更新成员
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")
	var member zerotier.Member
	if err := c.ShouldBindJSON(\u0026member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedMember, err := h.networkService.UpdateNetworkMember(networkID, memberID, \u0026member)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedMember)
}

// DeleteMember 删除成员
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	networkID := c.Param("networkId")
	memberID := c.Param("memberId")
	if err := h.networkService.RemoveNetworkMember(networkID, memberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}