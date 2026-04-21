package middleware

import (
	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/gofiber/fiber/v3"
)

func currentInitializedState() bool {
	return config.AppConfig != nil && config.AppConfig.Initialized
}

// SetupOnly blocks requests once the application has completed initialization.
func SetupOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		if currentInitializedState() {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error:   "Already Initialized",
				Message: "系统已初始化，当前接口仅在首次设置期间可用",
				Code:    fiber.StatusConflict,
			})
		}

		return c.Next()
	}
}

// InitializedOnly blocks runtime routes until the application has completed setup.
func InitializedOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !currentInitializedState() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Error:   "Setup Required",
				Message: "系统尚未初始化，请先完成设置向导",
				Code:    fiber.StatusServiceUnavailable,
			})
		}

		return c.Next()
	}
}
