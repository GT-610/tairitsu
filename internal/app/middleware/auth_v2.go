package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/services"
)

// AuthMiddlewareV2 新版认证中间件
func AuthMiddlewareV2(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "缺少认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "认证格式无效",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "无效的认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 将认证信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("key_type", claims.KeyType)

		c.Next()
	}
}

// SetupWizardAuthMiddleware 仅用于设置向导的认证中间件
func SetupWizardAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "缺少认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "认证格式无效",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "无效的认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证密钥类型是否为设置向导类型
		if claims.KeyType != services.KeyTypeSetupWizard {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Forbidden",
				Message: "需要设置向导认证",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// 将认证信息存储到上下文
		c.Set("key_type", claims.KeyType)

		c.Next()
	}
}

// NormalAuthMiddleware 仅用于普通认证的中间件
func NormalAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "缺少认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "认证格式无效",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "无效的认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证密钥类型是否为普通类型
		if claims.KeyType != services.KeyTypeNormal {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Forbidden",
				Message: "需要普通认证",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("key_type", claims.KeyType)

		c.Next()
	}
}