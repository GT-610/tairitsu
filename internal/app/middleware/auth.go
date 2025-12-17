package middleware

import (
	"strings"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtService *services.JWTService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 从请求头获取令牌
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:   "Unauthorized",
				Message: "缺少认证令牌",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:   "Unauthorized",
				Message: "认证格式无效",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// 验证令牌
		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:   "Unauthorized",
				Message: "无效的认证令牌",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// 将用户信息存储到上下文
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, exists := c.Locals("role").(string)
		if !exists {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "Forbidden",
				Message: "需要认证",
				Code:    fiber.StatusForbidden,
			})
		}

		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "Forbidden",
				Message: "需要管理员权限",
				Code:    fiber.StatusForbidden,
			})
		}

		return c.Next()
	}
}
