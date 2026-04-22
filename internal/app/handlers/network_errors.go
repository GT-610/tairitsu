package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func writeNetworkServiceError(c fiber.Ctx, err error, notFoundMessage string, forbiddenMessage string) error {
	switch {
	case services.IsNetworkNotFound(err):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": notFoundMessage})
	case services.IsNetworkAccessDenied(err):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": forbiddenMessage})
	case err == services.ErrImportAccessDenied:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case err == services.ErrImportOwnerRequired:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case err == services.ErrImportOwnerNotFound:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
