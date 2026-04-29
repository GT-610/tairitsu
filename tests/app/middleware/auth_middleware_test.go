package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)

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
	assert.Contains(t, string(body), "auth.missing_token")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)

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
	assert.Contains(t, string(body), "auth.invalid_format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)

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
	assert.Contains(t, string(body), "auth.invalid_token")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)

	// Create a test user and generate a valid token
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}
	token, _ := jwtService.GenerateToken(user, "session-1")

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
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)
	adminMiddleware := middleware.AdminRequired()

	// Create a test user with non-admin role and generate a valid token
	user := &models.User{
		ID:       "test-user-id",
		Username: "test-user",
		Role:     "user",
	}
	token, _ := jwtService.GenerateToken(user, "session-1")

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
	assert.Contains(t, string(body), "auth.admin_required")
}

func TestAdminRequired_AdminUser(t *testing.T) {
	// Arrange
	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)
	adminMiddleware := middleware.AdminRequired()

	// Create a test user with admin role and generate a valid token
	user := &models.User{
		ID:       "test-admin-id",
		Username: "test-admin",
		Role:     "admin",
	}
	token, _ := jwtService.GenerateToken(user, "session-1")

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

func TestAdminRequiredWithUserService_DeniesStaleAdminTokenAfterTransfer(t *testing.T) {
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	currentAdmin := &models.User{
		ID:        "admin-1",
		Username:  "admin",
		Password:  "hashed-password",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.CreateUser(currentAdmin))

	jwtService := services.NewJWTService("test-secret-key")
	authMiddleware := middleware.AuthMiddleware(jwtService, nil)
	adminMiddleware := middleware.AdminRequiredWithUserService(services.NewUserService(db))

	token, err := jwtService.GenerateToken(&models.User{
		ID:       currentAdmin.ID,
		Username: currentAdmin.Username,
		Role:     "admin",
	}, "session-1")
	require.NoError(t, err)

	router := fiber.New()
	router.Use(authMiddleware)
	router.Use(adminMiddleware)
	router.Get("/admin", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := router.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAuthMiddleware_RejectsRevokedSession(t *testing.T) {
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	user := &models.User{
		ID:        "user-1",
		Username:  "alice",
		Password:  "hashed-password",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.CreateUser(user))

	session := &models.Session{
		ID:         "session-1",
		UserID:     user.ID,
		UserAgent:  "test-agent",
		IPAddress:  "127.0.0.1",
		LastSeenAt: time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, db.CreateSession(session))

	sessionService := services.NewSessionService(db)
	jwtService := services.NewJWTService("test-secret-key")
	token, err := jwtService.GenerateToken(user, session.ID)
	require.NoError(t, err)
	require.NoError(t, sessionService.RevokeSession(user.ID, session.ID))

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(jwtService, sessionService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := router.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
