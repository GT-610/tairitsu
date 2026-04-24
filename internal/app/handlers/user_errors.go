package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func writeUserServiceError(c fiber.Ctx, err error) error {
	switch {
	case services.IsUserDBUnavailable(err):
		return writeErrorResponse(c, fiber.StatusServiceUnavailable, err.Error())
	case services.IsUsernameExists(err), services.IsInvalidUsername(err):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case services.IsInvalidCredentials(err):
		return writeErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	case services.IsPublicRegistrationDisabled(err):
		return writeErrorResponse(c, fiber.StatusForbidden, err.Error())
	case services.IsSessionRevoked(err), services.IsSessionExpired(err):
		return writeErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	case services.IsUserNotFound(err):
		return writeErrorResponse(c, fiber.StatusNotFound, err.Error())
	case services.IsSessionNotFound(err):
		return writeErrorResponse(c, fiber.StatusNotFound, err.Error())
	case services.IsOldPasswordIncorrect(err), services.IsInvalidUserRole(err), services.IsAdminTransferSelf(err), services.IsAdminResetSelf(err), services.IsAdminDeleteSelf(err), services.IsAdminDeleteBlocked(err), services.IsTransferTargetAdmin(err):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case services.IsAdminAccessDenied(err), services.IsSessionAccessDenied(err):
		return writeErrorResponse(c, fiber.StatusForbidden, err.Error())
	default:
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
}
