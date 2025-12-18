package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test router with the middleware
	router := fiber.New()
	router.Use(authMiddleware)
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := router.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "缺少认证令牌")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test router with the middleware
	router := fiber.New()
	router.Use(authMiddleware)
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	resp, err := router.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "认证格式无效")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test router with the middleware
	router := fiber.New()
	router.Use(authMiddleware)
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := router.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "无效的认证令牌")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test user and generate a valid token
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}
	token, _ := jwtService.GenerateToken(user)

	// Create a test router with the middleware
	router := fiber.New()
	router.Use(authMiddleware)
	router.Get("/test", func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
		username := c.Locals("username")
		role := c.Locals("role")
		return c.JSON(fiber.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := router.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "test-user-id")
	assert.Contains(t, string(body), "test-user")
	assert.Contains(t, string(body), "user")
}

func TestAdminRequired_NonAdminUser(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)
	adminMiddleware := middleware.AdminRequired()

	// Create a test user with non-admin role and generate a valid token
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}
	token, _ := jwtService.GenerateToken(user)

	// Create a test router with both middlewares
	router := fiber.New()
	router.Use(authMiddleware)
	router.Use(adminMiddleware)
	router.Get("/admin", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := router.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "需要管理员权限")
}

func TestAdminRequired_AdminUser(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)
	adminMiddleware := middleware.AdminRequired()

	// Create a test user with admin role and generate a valid token
	user := &models.User{
		ID:       "test-admin-id",
		Username: "test-admin",
		Role:     "admin",
	}
	token, _ := jwtService.GenerateToken(user)

	// Create a test router with both middlewares
	router := fiber.New()
	router.Use(authMiddleware)
	router.Use(adminMiddleware)
	router.Get("/admin", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := router.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "admin access granted")
}
