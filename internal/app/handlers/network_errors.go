package handlers

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func writeNetworkServiceError(c fiber.Ctx, err error, notFoundMessage string, forbiddenMessage string) error {
	switch {
	case services.IsNetworkNotFound(err):
		return writeErrorResponseWithCode(c, fiber.StatusNotFound, "network.not_found", notFoundMessage)
	case errors.Is(err, services.ErrNetworkPermissionDenied):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "network.permission_denied", "You do not have permission to perform this action on this network")
	case services.IsNetworkAccessDenied(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "network.access_denied", forbiddenMessage)
	case errors.Is(err, services.ErrNetworkGrantNotFound):
		return writeErrorResponseWithCode(c, fiber.StatusNotFound, "network.grant_not_found", "Network member not found")
	case errors.Is(err, services.ErrNetworkGrantExists):
		return writeErrorResponseWithCode(c, fiber.StatusConflict, "network.grant_exists", "User already has access to this network")
	case errors.Is(err, services.ErrNetworkGrantChanged):
		return writeErrorResponseWithCode(c, fiber.StatusConflict, "network.grant_conflict", "This member's role changed while the request was in flight. Reload and try again.")
	case errors.Is(err, services.ErrNetworkGrantRoleInvalid):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.grant_role_invalid", "Invalid network role")
	case errors.Is(err, services.ErrNetworkGrantTargetInvalid):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.grant_target_invalid", "This user cannot be granted access to the network")
	case errors.Is(err, services.ErrNetworkGrantSelfChange):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.grant_self_change", "Use the leave action to remove your own access")
	case errors.Is(err, services.ErrNetworkOwnerTransferInvalid):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.owner_transfer_invalid", "Network ownership cannot be transferred to that user")
	case errors.Is(err, services.ErrNetworkOwnerCannotLeave):
		return writeErrorResponseWithCode(c, fiber.StatusConflict, "network.owner_cannot_leave", "Transfer ownership before leaving the network")
	case errors.Is(err, services.ErrImportAccessDenied):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "network.import_access_denied", err.Error())
	case errors.Is(err, services.ErrImportOwnerRequired):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.import_owner_required", err.Error())
	case errors.Is(err, services.ErrImportOwnerNotFound):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.import_owner_not_found", err.Error())
	case errors.Is(err, services.ErrUserNotFound):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.not_found", err.Error())
	default:
		logger.Error("unhandled network service error", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusInternalServerError, "system.internal_error", "Internal Server Error")
	}
}
