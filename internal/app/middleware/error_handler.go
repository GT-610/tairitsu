package middleware

import (
	"net/http"

	"github.com/GT-610/tairitsu/internal/app/errors"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorResponse Error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Status  int    `json:"status,omitempty"`
}

// ErrorHandler Global error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("API Error", zap.Error(err.Err))

			// Check if it's a custom AppError
			if appErr, ok := errors.IsAppError(err.Err); ok {
				// Respond with custom error
				c.JSON(appErr.StatusCode, ErrorResponse{
					Error:   appErr.Code,
					Message: appErr.Message,
					Code:    appErr.Code,
					Status:  appErr.StatusCode,
				})
				return
			}

			// Respond with default error
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "INTERNAL_SERVER_ERROR",
				Message: err.Error(),
				Code:    "INTERNAL_SERVER_ERROR",
				Status:  http.StatusInternalServerError,
			})
		}
	}
}

// CORS Cross-origin resource sharing middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
