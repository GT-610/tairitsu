package middleware

import (
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Logger is the custom logging middleware
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Start time
		startTime := time.Now()

		// Process the request
		err := c.Next()

		// End time
		endTime := time.Now()
		// Latency
		latency := endTime.Sub(startTime)

		// Request method
		method := c.Method()
		// Request path
		path := c.Path()
		// Status code
		statusCode := c.Response().StatusCode()
		// Client IP
		clientIP := c.IP()

		// Log format
		logger.Info("FIBER Request",
			zap.String("method", method),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("clientIP", clientIP),
			zap.String("path", path),
		)

		return err
	}
}
