package handlers

import (
	"fmt"
	"strings"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// AccessHandler manages account roles on one network. Every endpoint delegates
// the permission decision to the network service; handlers only validate shape.
type AccessHandler struct {
	networkService *services.NetworkService
}

func NewAccessHandler(networkService *services.NetworkService) *AccessHandler {
	return &AccessHandler{networkService: networkService}
}

// GetNetworkAccess lists the owner and delegated roles for a network.
func (h *AccessHandler) GetNetworkAccess(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	access, err := h.networkService.GetNetworkAccess(networkID, userID)
	if err != nil {
		logger.Error("Failed to get network access", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return c.Status(fiber.StatusOK).JSON(access)
}

// GetNetworkAccessCandidates searches accounts that are not on the network yet.
func (h *AccessHandler) GetNetworkAccessCandidates(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	candidates, err := h.networkService.GetNetworkAccessCandidates(networkID, userID, c.Query("q"))
	if err != nil {
		logger.Error("Failed to get network access candidates", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return c.Status(fiber.StatusOK).JSON(candidates)
}

// GrantNetworkAccess adds one account to a network with the given role.
func (h *AccessHandler) GrantNetworkAccess(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	var request struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateAccessTarget(request.UserID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateAccessRole(request.Role); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := h.networkService.GrantNetworkAccess(networkID, userID, request.UserID, request.Role); err != nil {
		logger.Error("Failed to grant network access", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "Network access granted", nil)
}

// UpdateNetworkAccess changes one account's role.
func (h *AccessHandler) UpdateNetworkAccess(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	targetUserID := c.Params("userId")
	if err := validateAccessTarget(targetUserID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	var request struct {
		Role string `json:"role"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateAccessRole(request.Role); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := h.networkService.UpdateNetworkAccess(networkID, userID, targetUserID, request.Role); err != nil {
		logger.Error("Failed to update network access", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "Network access updated", nil)
}

// RevokeNetworkAccess removes one account from a network.
func (h *AccessHandler) RevokeNetworkAccess(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	targetUserID := c.Params("userId")
	if err := validateAccessTarget(targetUserID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	if err := h.networkService.RevokeNetworkAccess(networkID, userID, targetUserID); err != nil {
		logger.Error("Failed to revoke network access", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "Network access revoked", nil)
}

// TransferNetworkOwnership moves the network to another member.
func (h *AccessHandler) TransferNetworkOwnership(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	var request struct {
		UserID            string `json:"user_id"`
		PreviousOwnerRole string `json:"previous_owner_role"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateAccessTarget(request.UserID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := h.networkService.TransferNetworkOwnership(networkID, userID, request.UserID, request.PreviousOwnerRole); err != nil {
		logger.Error("Failed to transfer network ownership", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "Network ownership transferred", nil)
}

// LeaveNetwork removes the caller's own grant.
func (h *AccessHandler) LeaveNetwork(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	if err := h.networkService.LeaveNetwork(networkID, userID); err != nil {
		logger.Error("Failed to leave network", zap.String("network_id", networkID), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "Left network", nil)
}

func validateAccessTarget(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("user is required")
	}
	if len(value) > 128 {
		return fmt.Errorf("user ID is too long")
	}
	return nil
}

func validateAccessRole(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("role is required")
	}
	return nil
}
