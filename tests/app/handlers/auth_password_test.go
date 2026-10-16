package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_ChangePasswordRevokesOtherSessionsWhenRequested(t *testing.T) {
	db, err := database.NewDatabase(database.Config{Type: database.SQLite, Path: filepath.Join(t.TempDir(), "tairitsu.db")})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	authHandler := apphandlers.NewAuthHandler(userService, sessionService, nil, nil)

	user, err := userService.Register(&models.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	}, "user")
	require.NoError(t, err)

	_, credential := newTestSession(t, sessionService, user.ID, "current-browser", "127.0.0.1")
	otherSession, _ := newTestSession(t, sessionService, user.ID, "other-browser", "127.0.0.2")

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(sessionService, userService))
	app.Put("/profile/password", authHandler.ChangePassword)

	body, err := json.Marshal(map[string]any{
		"current_password":      "secret123",
		"new_password":          "updated456",
		"confirm_password":      "updated456",
		"logout_other_sessions": true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/profile/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	authenticateRequest(req, credential)
	resp, err := app.Test(req, fiber.TestConfig{
		Timeout:       10 * time.Second,
		FailOnTimeout: true,
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var responseBody struct {
		Message              string `json:"message"`
		RevokedOtherSessions int    `json:"revoked_other_sessions"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
	assert.Equal(t, "Password updated successfully", responseBody.Message)
	assert.Equal(t, 1, responseBody.RevokedOtherSessions)

	_, err = sessionService.ResolveSession(credential)
	require.NoError(t, err)

	revoked, err := sessionService.GetSessionByID(otherSession.ID)
	require.NoError(t, err)
	assert.NotNil(t, revoked.RevokedAt, "the other session must be revoked")

	updatedUser, err := userService.Login(&models.LoginRequest{
		Username: "alice",
		Password: "updated456",
	})
	require.NoError(t, err)
	assert.Equal(t, user.ID, updatedUser.ID)
}
