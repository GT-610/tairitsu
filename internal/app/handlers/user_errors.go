package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func writeUserServiceError(c fiber.Ctx, err error) error {
	switch {
	case services.IsUserDBUnavailable(err):
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	case services.IsUsernameExists(err):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case services.IsInvalidCredentials(err):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	case services.IsPublicRegistrationDisabled(err):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case services.IsUserNotFound(err):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case services.IsOldPasswordIncorrect(err), services.IsInvalidUserRole(err), services.IsAdminTransferSelf(err), services.IsTransferTargetAdmin(err):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case services.IsAdminAccessDenied(err):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
