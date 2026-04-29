package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// MemberHandler 成员处理器
type MemberHandler struct {
	networkService *services.NetworkService
}

func memberRouteNetworkID(c fiber.Ctx) string {
	networkID := c.Params("networkId")
	if networkID == "" {
		networkID = c.Params("id")
	}
	return networkID
}

// NewMemberHandler 创建成员处理器实例
func NewMemberHandler(networkService *services.NetworkService) *MemberHandler {
	return &MemberHandler{
		networkService: networkService,
	}
}

// GetMembers 获取网络中的所有成员
func (h *MemberHandler) GetMembers(c fiber.Ctx) error {
	networkID := memberRouteNetworkID(c)

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	members, err := h.networkService.GetNetworkMembers(networkID, userID)
	if err != nil {
		logger.Error("Failed to get network members", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network member access denied")
	}

	return c.Status(fiber.StatusOK).JSON(members)
}

// GetMember 获取网络中的特定成员
func (h *MemberHandler) GetMember(c fiber.Ctx) error {
	networkID := memberRouteNetworkID(c)
	memberID := c.Params("memberId")

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	member, err := h.networkService.GetNetworkMember(networkID, memberID, userID)
	if err != nil {
		logger.Error("Failed to get network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network member access denied")
	}

	if member == nil {
		logger.Warn("Network member not found", zap.String("network_id", networkID), zap.String("member_id", memberID))
		return writeErrorResponseWithCode(c, fiber.StatusNotFound, "member.not_found", "Member not found")
	}

	return c.Status(fiber.StatusOK).JSON(member)
}

// UpdateMember 更新网络成员
func (h *MemberHandler) UpdateMember(c fiber.Ctx) error {
	networkID := memberRouteNetworkID(c)
	memberID := c.Params("memberId")

	var req zerotier.MemberUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind request", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	member, err := h.networkService.UpdateNetworkMember(networkID, memberID, &req, userID)
	if err != nil {
		logger.Error("Failed to update network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network member update access denied")
	}

	return c.Status(fiber.StatusOK).JSON(member)
}

// DeleteMember 删除网络成员
func (h *MemberHandler) DeleteMember(c fiber.Ctx) error {
	networkID := memberRouteNetworkID(c)
	memberID := c.Params("memberId")

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	err := h.networkService.RemoveNetworkMember(networkID, memberID, userID)
	if err != nil {
		logger.Error("Failed to delete network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network member delete access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "member.delete_success", "Member deleted successfully", nil)
}
