package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/services"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
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
		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "无效的认证令牌",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Forbidden",
				Message: "需要认证",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Forbidden",
				Message: "需要管理员权限",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}