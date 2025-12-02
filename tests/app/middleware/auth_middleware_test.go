package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test router with the middleware
	router := gin.Default()
	router.Use(authMiddleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test request without Authorization header
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "缺少认证令牌")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test router with the middleware
	router := gin.Default()
	router.Use(authMiddleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test request with invalid token format
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "认证格式无效")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// Create a test router with the middleware
	router := gin.Default()
	router.Use(authMiddleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test request with invalid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "无效的认证令牌")
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
	router := gin.Default()
	router.Use(authMiddleware)
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	// Create a test request with valid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-user-id")
	assert.Contains(t, w.Body.String(), "test-user")
	assert.Contains(t, w.Body.String(), "user")
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
	router := gin.Default()
	router.Use(authMiddleware)
	router.Use(adminMiddleware)
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create a test request with non-admin token
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "需要管理员权限")
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
	router := gin.Default()
	router.Use(authMiddleware)
	router.Use(adminMiddleware)
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create a test request with admin token
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin access granted")
}
