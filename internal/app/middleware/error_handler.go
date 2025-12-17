package middleware

import (
	"net/http"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// ErrorHandler 全局错误处理中间件
func ErrorHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			logger.Error("API错误", zap.Error(err))

			// 响应错误
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error:   "Internal Server Error",
				Message: err.Error(),
				Code:    fiber.StatusInternalServerError,
			})
		}
		return nil
	}
}

// CORS 跨域中间件
func CORS() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(http.StatusNoContent)
		}

		return c.Next()
	}
}
