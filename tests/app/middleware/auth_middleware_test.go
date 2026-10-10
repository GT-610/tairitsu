package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
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

// sessionCookieName mirrors the cookie the login handler issues.
const sessionCookieName = middleware.SessionCookieName

func newAuthTestDB(t *testing.T) database.DBInterface {
	t.Helper()

	db, err := database.NewDatabase(database.Config{Type: database.SQLite, Path: t.TempDir() + "/tairitsu.db"})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func createTestUser(t *testing.T, db database.DBInterface, user *models.User) {
	t.Helper()

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}
	if user.Password == "" {
		user.Password = "hashed-password"
	}
	require.NoError(t, db.CreateUser(user))
}

// issueSession creates a session and returns the browser credential.
func issueSession(t *testing.T, sessionService *services.SessionService, userID string) string {
	t.Helper()

	_, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     userID,
		UserAgent:  "test-agent",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	return credential
}

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	db := newAuthTestDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "auth.missing_token")
}

func TestAuthMiddleware_RejectsUnknownCredential(t *testing.T) {
	db := newAuthTestDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "forged-credential"})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "auth.invalid_token")
}

func TestAuthMiddleware_AcceptsValidSessionCookie(t *testing.T) {
	db := newAuthTestDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	user := &models.User{ID: "test-user-id", Username: "test-user", Role: "user"}
	createTestUser(t, db, user)
	credential := issueSession(t, sessionService, user.ID)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":  c.Locals("user_id"),
			"username": c.Locals("username"),
			"role":     c.Locals("role"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "test-user-id")
	assert.Contains(t, string(body), "test-user")
}

func TestAuthMiddleware_RejectsCredentialAfterLogout(t *testing.T) {
	db := newAuthTestDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	user := &models.User{ID: "user-1", Username: "alice", Role: "user"}
	createTestUser(t, db, user)

	session, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:    user.ID,
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	require.NoError(t, sessionService.RevokeSession(user.ID, session.ID))

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_ClearsCookieWhenUserIsGone(t *testing.T) {
	db := newAuthTestDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	// A session whose user no longer exists must not keep working, and the
	// browser should be told to drop the credential.
	credential := issueSession(t, sessionService, "deleted-user")

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	cleared := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName && cookie.Value == "" {
			cleared = true
		}
	}
	assert.True(t, cleared, "expected the session cookie to be cleared")
}

func TestAdminRequiredWithUserService_DeniesStaleAdminSessionAfterTransfer(t *testing.T) {
	db := newAuthTestDB(t)
	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)

	currentAdmin := &models.User{ID: "admin-1", Username: "admin", Role: "user"}
	createTestUser(t, db, currentAdmin)
	credential := issueSession(t, sessionService, currentAdmin.ID)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Use(middleware.AdminRequiredWithUserService(userService))
	router.Get("/admin", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSessionCookieIsHttpOnlyAndSameSiteLax(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)

	persistent := middleware.SessionCookie("credential", expiresAt, true, false)
	assert.True(t, persistent.HTTPOnly)
	assert.Equal(t, fiber.CookieSameSiteLaxMode, persistent.SameSite)
	assert.Equal(t, "/", persistent.Path)
	// Plain HTTP must not set Secure, or the cookie would never be sent.
	assert.False(t, persistent.Secure)
	assert.Positive(t, persistent.MaxAge)

	transient := middleware.SessionCookie("credential", expiresAt, false, false)
	assert.True(t, transient.SessionOnly)
	assert.Zero(t, transient.MaxAge)

	secure := middleware.SessionCookie("credential", expiresAt, true, true)
	assert.True(t, secure.Secure)
}

func TestExpiredSessionCookieClearsTheCredential(t *testing.T) {
	cookie := middleware.ExpiredSessionCookie(false)

	assert.Equal(t, middleware.SessionCookieName, cookie.Name)
	assert.Empty(t, cookie.Value)
	assert.Negative(t, cookie.MaxAge)
	assert.True(t, cookie.HTTPOnly)
}
