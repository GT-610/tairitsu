package handlers

import (
	"math"
	"strconv"

	"github.com/GT-610/tempest-zt/internal/app/logger"
	"github.com/GT-610/tempest-zt/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func writeUserServiceError(c fiber.Ctx, err error) error {
	switch {
	case services.IsUserDBUnavailable(err):
		logger.Error("User service database is unavailable", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusServiceUnavailable, "user.db_unavailable", "User service is unavailable")
	case services.IsUsernameExists(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.username_exists", err.Error())
	case services.IsInvalidUsername(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.invalid_username", err.Error())
	case services.IsUsernameTooLong(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.username_too_long", err.Error())
	case services.IsPasswordTooShort(err), services.IsPasswordTooLong(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.invalid_password", err.Error())
	case services.IsInvalidCredentials(err):
		return writeErrorResponseWithCode(c, fiber.StatusUnauthorized, "user.invalid_credentials", err.Error())
	case services.IsLoginThrottled(err):
		// Tell the caller how long to wait rather than leaving them to guess, and
		// keep the wording free of anything that hints whether the account exists.
		if retryAfter := services.LoginRetryAfter(err); retryAfter > 0 {
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
		}
		return writeErrorResponseWithCode(c, fiber.StatusTooManyRequests, "auth.too_many_attempts", "Too many sign-in attempts. Try again later.")
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
	case services.IsAdminTransferSelf(err), services.IsAdminResetSelf(err), services.IsAdminDeleteSelf(err), services.IsAdminDeleteBlocked(err), services.IsTransferTargetAdmin(err):
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.invalid_admin_operation", err.Error())
	case services.IsAdminAlreadyExists(err):
		return writeErrorResponseWithCode(c, fiber.StatusConflict, "user.admin_already_exists", err.Error())
	case services.IsAdminAccessDenied(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "user.admin_access_denied", err.Error())
	case services.IsSessionAccessDenied(err):
		return writeErrorResponseWithCode(c, fiber.StatusForbidden, "session.access_denied", err.Error())
	default:
		logger.Error("Unhandled user service error", zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusInternalServerError, "system.internal_error", "Internal Server Error")
	}
}
