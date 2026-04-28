package middleware

import (
	"strings"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtService *services.JWTService, sessionService *services.SessionService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 从请求头获取令牌
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:     "Unauthorized",
				Message:   "Missing authentication token",
				ErrorCode: "auth.missing_token",
				Code:      fiber.StatusUnauthorized,
			})
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:     "Unauthorized",
				Message:   "Invalid authentication format",
				ErrorCode: "auth.invalid_format",
				Code:      fiber.StatusUnauthorized,
			})
		}

		// 验证令牌
		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:     "Unauthorized",
				Message:   "Invalid authentication token",
				ErrorCode: "auth.invalid_token",
				Code:      fiber.StatusUnauthorized,
			})
		}

		// 将用户信息存储到上下文
		if sessionService != nil {
			session, err := sessionService.ValidateSession(claims.UserID, claims.SessionID)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
					Error:     "Unauthorized",
					Message:   "Invalid authentication token",
					ErrorCode: "auth.invalid_token",
					Code:      fiber.StatusUnauthorized,
				})
			}
			_ = sessionService.TouchSession(session)
			c.Locals("session_id", claims.SessionID)
		}

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
				Error:     "Forbidden",
				Message:   "Authentication required",
				ErrorCode: "auth.required",
				Code:      fiber.StatusForbidden,
			})
		}

		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
				Code:      fiber.StatusForbidden,
			})
		}

		return c.Next()
	}
}

func AdminRequiredWithUserService(userService *services.UserService) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, exists := c.Locals("user_id").(string)
		if !exists || userID == "" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Authentication required",
				ErrorCode: "auth.required",
				Code:      fiber.StatusForbidden,
			})
		}

		if userService == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Error:     "Service Unavailable",
				Message:   "User service is unavailable",
				ErrorCode: "system.user_service_unavailable",
				Code:      fiber.StatusServiceUnavailable,
			})
		}

		user, err := userService.GetUserByID(userID)
		if err != nil {
			if services.IsUserDBUnavailable(err) {
				logger.Error("Administrator authorization failed because user database is unavailable", zap.Error(err))
				return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
					Error:     "Service Unavailable",
					Message:   "User service is unavailable",
					ErrorCode: "user.db_unavailable",
					Code:      fiber.StatusServiceUnavailable,
				})
			}
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
				Code:      fiber.StatusForbidden,
			})
		}

		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
				Code:      fiber.StatusForbidden,
			})
		}

		c.Locals("role", user.Role)
		return c.Next()
	}
}
