package handlers

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func writeNetworkServiceError(c fiber.Ctx, err error, notFoundMessage string, forbiddenMessage string) error {
	switch {
	case services.IsNetworkNotFound(err):
		return writeErrorResponse(c, fiber.StatusNotFound, notFoundMessage)
	case services.IsNetworkAccessDenied(err):
		return writeErrorResponse(c, fiber.StatusForbidden, forbiddenMessage)
	case errors.Is(err, services.ErrImportAccessDenied):
		return writeErrorResponse(c, fiber.StatusForbidden, err.Error())
	case errors.Is(err, services.ErrImportOwnerRequired):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrImportOwnerNotFound):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrUserNotFound):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrViewerTargetInvalid):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	default:
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
}
