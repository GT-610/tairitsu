package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// Logger Custom logging middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		startTime := time.Now()

		// Process request
		c.Next()

		// End time
		endTime := time.Now()
		// Execution time
		latency := endTime.Sub(startTime)

		// Request method
		method := c.Request.Method
		// Request route
		path := c.Request.URL.Path
		// Status code
		statusCode := c.Writer.Status()
		// Client IP
		clientIP := c.ClientIP()

		// Log format
		logger.Info("GIN Request",
			zap.String("method", method),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("clientIP", clientIP),
			zap.String("path", path),
		)
	}
}