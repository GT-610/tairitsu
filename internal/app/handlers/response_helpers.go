package handlers

import (
	"strings"

	"github.com/GT-610/tairitsu/internal/app/httpcode"
	"github.com/gofiber/fiber/v3"
)

func writeErrorResponse(c fiber.Ctx, status int, message string) error {
	return writeErrorResponseWithCode(c, status, httpcode.DefaultErrorCode(status), message)
}

func writeErrorResponseWithCode(c fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"message":    message,
		"error_code": code,
		"code":       status,
	})
}

// writeErrorResponseWithDetail returns an error response that includes a
// client-safe detail string. Use this instead of writeErrorResponseWithCode
// when callers need to surface a sanitized explanation (e.g. a human-readable
// setup error code). Never pass raw err.Error()—callers must strip internal
// paths, stack traces, and wrapped causes before calling this function.
func writeErrorResponseWithDetail(c fiber.Ctx, status int, code string, message string, detail string) error {
	detail = strings.TrimSpace(detail)
	if len(detail) > 256 {
		detail = detail[:256]
	}
	return c.Status(status).JSON(fiber.Map{
		"message":    message,
		"error_code": code,
		"code":       status,
		"detail":     detail,
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
