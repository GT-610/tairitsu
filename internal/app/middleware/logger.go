package middleware

import (
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Logger 自定义日志中间件
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		err := c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间
		latency := endTime.Sub(startTime)

		// 请求方法
		method := c.Method()
		// 请求路由
		path := c.Path()
		// 状态码
		statusCode := c.Response().StatusCode()
		// 客户端IP
		clientIP := c.IP()

		// 日志格式
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
