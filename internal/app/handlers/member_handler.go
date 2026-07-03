package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// MemberHandler handles member-related HTTP requests
type MemberHandler struct {
	networkService *services.NetworkService
}

// NewMemberHandler creates a new member handler instance
func NewMemberHandler(networkService *services.NetworkService) *MemberHandler {
	return &MemberHandler{
		networkService: networkService,
	}
}

// GetMembers retrieves all members in a network
func (h *MemberHandler) GetMembers(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

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

// GetMember retrieves a specific member in a network
func (h *MemberHandler) GetMember(c fiber.Ctx) error {
	networkID := c.Params("id")
	memberID := c.Params("memberId")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateMemberID(memberID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

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

// UpdateMember updates a network member
func (h *MemberHandler) UpdateMember(c fiber.Ctx) error {
	networkID := c.Params("id")
	memberID := c.Params("memberId")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateMemberID(memberID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	var req zerotier.MemberUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind request", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := validateMemberName(req.Name); err != nil {
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

// DeleteMember deletes a network member
func (h *MemberHandler) DeleteMember(c fiber.Ctx) error {
	networkID := c.Params("id")
	memberID := c.Params("memberId")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateMemberID(memberID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

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
