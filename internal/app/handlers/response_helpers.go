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
		"detail":     detail,
	})
}

// writeMessageResponse returns a success body. It carries a human-readable
// message and nothing else: the success codes that used to accompany it were
// never read by the client, which renders its own copy.
func writeMessageResponse(c fiber.Ctx, status int, message string, extra fiber.Map) error {
	body := fiber.Map{"message": message}
	for key, value := range extra {
		if key == "message" {
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
