package middleware

import (
	"errors"
	"net/http"

	"github.com/GT-610/tairitsu/internal/app/httpcode"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Code      int    `json:"code,omitempty"`
}

// ErrorHandler 全局错误处理中间件
func ErrorHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			logger.Error("API error", zap.Error(err))

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return c.Status(fiberErr.Code).JSON(ErrorResponse{
					Error:     http.StatusText(fiberErr.Code),
					Message:   fiberErr.Message,
					ErrorCode: httpcode.DefaultErrorCode(fiberErr.Code),
					Code:      fiberErr.Code,
				})
			}

			// 响应错误
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error:     "Internal Server Error",
				Message:   "Internal Server Error",
				ErrorCode: "system.internal_error",
				Code:      fiber.StatusInternalServerError,
			})
		}
		return nil
	}
}
