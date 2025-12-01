package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/services"
)

// AuthMiddleware Authentication middleware
func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from request header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Missing authentication token",
				Code:    "UNAUTHORIZED",
				Status:  http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid authentication format",
				Code:    "UNAUTHORIZED",
				Status:  http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid authentication token",
				Code:    "UNAUTHORIZED",
				Status:  http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminRequired Admin permission middleware
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Forbidden",
				Message: "Authentication required",
				Code:    "FORBIDDEN",
				Status:  http.StatusForbidden,
			})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Forbidden",
				Message: "Admin permission required",
				Code:    "FORBIDDEN",
				Status:  http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}