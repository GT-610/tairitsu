package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/httpcode"
	"github.com/gofiber/fiber/v3"
)

func writeErrorResponse(c fiber.Ctx, status int, message string) error {
	return writeErrorResponseWithCode(c, status, httpcode.DefaultErrorCode(status), message)
}

func writeErrorResponseWithCode(c fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error":      message,
		"message":    message,
		"error_code": code,
		"code":       status,
	})
}

func writeMessageResponse(c fiber.Ctx, status int, code string, message string, extra fiber.Map) error {
	body := fiber.Map{
		"message":      message,
		"message_code": code,
	}
	for key, value := range extra {
		if key == "message" || key == "message_code" {
			continue
		}
		body[key] = value
	}
	return c.Status(status).JSON(body)
}

func requiredUserID(c fiber.Ctx) (string, error) {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return "", writeErrorResponseWithCode(c, fiber.StatusUnauthorized, "auth.unauthorized", "Unauthorized access")
	}
	return userID, nil
}
