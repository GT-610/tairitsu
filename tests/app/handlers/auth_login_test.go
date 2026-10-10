package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/database"
	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_LoginRecordsForwardedClientIP(t *testing.T) {
	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: filepath.Join(t.TempDir(), "tairitsu.db"),
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	_, err = userService.Register(&models.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	}, "user")
	require.NoError(t, err)

	authHandler := apphandlers.NewAuthHandler(userService, sessionService, nil, nil)
	app := fiber.New(fiber.Config{
		TrustProxy:         true,
		ProxyHeader:        "X-Real-IP",
		TrustProxyConfig:   fiber.TrustProxyConfig{Proxies: []string{"0.0.0.0/0", "::/0"}},
		EnableIPValidation: true,
	})
	app.Post("/auth/login", authHandler.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"username":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "203.0.113.10")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body struct {
		Session models.SessionResponse `json:"session"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "203.0.113.10", body.Session.IPAddress)
}

func TestAuthHandler_LoginIssuesHttpOnlySessionCookie(t *testing.T) {
	db := newTestDatabase(t)

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	_, err := userService.Register(&models.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	}, "user")
	require.NoError(t, err)

	authHandler := apphandlers.NewAuthHandler(userService, sessionService, nil, nil)
	app := fiber.New()
	app.Post("/auth/login", authHandler.Login)
	app.Get("/profile", middleware.AuthMiddleware(sessionService, userService), authHandler.GetProfile)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"username":"alice","password":"secret123","remember_me":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
		}
	}
	require.NotNil(t, sessionCookie, "login must set the session cookie")
	assert.True(t, sessionCookie.HttpOnly, "the session cookie must not be readable by scripts")
	assert.NotEmpty(t, sessionCookie.Value)

	// The issued cookie must authenticate a following request.
	profileReq := httptest.NewRequest(http.MethodGet, "/profile", nil)
	profileReq.AddCookie(sessionCookie)
	profileResp, err := app.Test(profileReq)
	require.NoError(t, err)
	defer profileResp.Body.Close()
	assert.Equal(t, fiber.StatusOK, profileResp.StatusCode)
}

func TestAuthHandler_LoginResponseOmitsCredential(t *testing.T) {
	db := newTestDatabase(t)

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	_, err := userService.Register(&models.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	}, "user")
	require.NoError(t, err)

	authHandler := apphandlers.NewAuthHandler(userService, sessionService, nil, nil)
	app := fiber.New()
	app.Post("/auth/login", authHandler.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"username":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// The credential must live only in the cookie, never in the payload that
	// frontend JavaScript can read.
	var rawBody map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rawBody))
	assert.NotContains(t, rawBody, "token")
	assert.Contains(t, rawBody, "user")
	assert.Contains(t, rawBody, "session")
}

func TestAuthHandler_LogoutClearsSessionCookie(t *testing.T) {
	db := newTestDatabase(t)

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	user, err := userService.Register(&models.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	}, "user")
	require.NoError(t, err)

	_, credential := newTestSession(t, sessionService, user.ID, "test-agent", "127.0.0.1")

	authHandler := apphandlers.NewAuthHandler(userService, sessionService, nil, nil)
	app := fiber.New()
	app.Post("/auth/logout", middleware.AuthMiddleware(sessionService, userService), authHandler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	authenticateRequest(req, credential)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	cleared := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == middleware.SessionCookieName && cookie.Value == "" {
			cleared = true
		}
	}
	assert.True(t, cleared, "logout must clear the session cookie")

	// The revoked credential must no longer authenticate.
	retryReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	authenticateRequest(retryReq, credential)
	retryResp, err := app.Test(retryReq)
	require.NoError(t, err)
	defer retryResp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, retryResp.StatusCode)
}
