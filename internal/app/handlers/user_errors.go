package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func writeUserServiceError(c fiber.Ctx, err error) error {
	switch {
	case services.IsUserDBUnavailable(err):
		return writeErrorResponseWithCode(c, fiber.StatusServiceUnavailable, "user.db_unavailable", err.Error())
	case services.IsUsernameExists(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.username_exists", err.Error())
	case services.IsInvalidUsername(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.invalid_username", err.Error())
	case services.IsInvalidCredentials(err):
		return writeErrorResponseWithCode(c, fiber.StatusUnauthorized, "user.invalid_credentials", err.Error())
	case services.IsPublicRegistrationDisabled(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "user.public_registration_disabled", err.Error())
	case services.IsSessionRevoked(err):
		return writeErrorResponseWithCode(c, fiber.StatusUnauthorized, "session.revoked", err.Error())
	case services.IsSessionExpired(err):
		return writeErrorResponseWithCode(c, fiber.StatusUnauthorized, "session.expired", err.Error())
	case services.IsUserNotFound(err):
		return writeErrorResponseWithCode(c, fiber.StatusNotFound, "user.not_found", err.Error())
	case services.IsSessionNotFound(err):
		return writeErrorResponseWithCode(c, fiber.StatusNotFound, "session.not_found", err.Error())
	case services.IsOldPasswordIncorrect(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.old_password_incorrect", err.Error())
	case services.IsInvalidUserRole(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.invalid_role", err.Error())
	case services.IsAdminTransferSelf(err), services.IsAdminResetSelf(err), services.IsAdminDeleteSelf(err), services.IsAdminDeleteBlocked(err), services.IsTransferTargetAdmin(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.invalid_admin_operation", err.Error())
	case services.IsAdminAccessDenied(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "user.admin_access_denied", err.Error())
	case services.IsSessionAccessDenied(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "session.access_denied", err.Error())
	default:
		return writeErrorResponseWithCode(c, fiber.StatusInternalServerError, "system.internal_error", err.Error())
	}
}
