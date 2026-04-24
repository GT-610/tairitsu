package handlers

import "github.com/gofiber/fiber/v3"

func writeErrorResponse(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func requiredUserID(c fiber.Ctx) (string, error) {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return "", writeErrorResponse(c, fiber.StatusUnauthorized, "未授权访问")
	}
	return userID, nil
}
