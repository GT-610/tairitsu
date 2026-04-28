package handlers

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func writeNetworkServiceError(c fiber.Ctx, err error, notFoundMessage string, forbiddenMessage string) error {
	switch {
	case services.IsNetworkNotFound(err):
		return writeErrorResponseWithCode(c, fiber.StatusNotFound, "network.not_found", notFoundMessage)
	case services.IsNetworkAccessDenied(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "network.access_denied", forbiddenMessage)
	case errors.Is(err, services.ErrImportAccessDenied):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "network.import_access_denied", err.Error())
	case errors.Is(err, services.ErrImportOwnerRequired):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.import_owner_required", err.Error())
	case errors.Is(err, services.ErrImportOwnerNotFound):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.import_owner_not_found", err.Error())
	case errors.Is(err, services.ErrUserNotFound):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.not_found", err.Error())
	case errors.Is(err, services.ErrViewerTargetInvalid):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.viewer_target_invalid", err.Error())
	default:
		return writeErrorResponseWithCode(c, fiber.StatusInternalServerError, "system.internal_error", err.Error())
	}
}
