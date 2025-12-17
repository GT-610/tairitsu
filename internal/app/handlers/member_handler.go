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

// NewMemberHandler 创建成员处理器实例
func NewMemberHandler(networkService *services.NetworkService) *MemberHandler {
	return &MemberHandler{
		networkService: networkService,
	}
}

// GetMembers 获取网络中的所有成员
func (h *MemberHandler) GetMembers(c fiber.Ctx) error {
	networkID := c.Params("networkId")

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	members, err := h.networkService.GetNetworkMembers(networkID, userID.(string))
	if err != nil {
		logger.Error("获取网络成员列表失败", zap.String("network_id", networkID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(members)
}

// GetMember 获取网络中的特定成员
func (h *MemberHandler) GetMember(c fiber.Ctx) error {
	networkID := c.Params("networkId")
	memberID := c.Params("memberId")

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	member, err := h.networkService.GetNetworkMember(networkID, memberID, userID.(string))
	if err != nil {
		logger.Error("获取网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "获取成员失败: " + err.Error()})
	}

	if member == nil {
		logger.Warn("网络成员不存在", zap.String("network_id", networkID), zap.String("member_id", memberID))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "成员不存在"})
	}

	return c.Status(fiber.StatusOK).JSON(member)
}

// UpdateMember 更新网络成员
func (h *MemberHandler) UpdateMember(c fiber.Ctx) error {
	networkID := c.Params("networkId")
	memberID := c.Params("memberId")

	var req zerotier.Member
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("请求参数绑定失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	member, err := h.networkService.UpdateNetworkMember(networkID, memberID, &req, userID.(string))
	if err != nil {
		logger.Error("更新网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "更新成员失败: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(member)
}

// DeleteMember 删除网络成员
func (h *MemberHandler) DeleteMember(c fiber.Ctx) error {
	networkID := c.Params("networkId")
	memberID := c.Params("memberId")

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	err := h.networkService.RemoveNetworkMember(networkID, memberID, userID.(string))
	if err != nil {
		logger.Error("删除网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "删除成员失败: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "成员删除成功"})
}
