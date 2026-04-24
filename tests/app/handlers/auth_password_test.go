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
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	jwtService := services.NewJWTService("test-secret")
	authHandler := apphandlers.NewAuthHandler(userService, sessionService, jwtService, nil, nil)

	user, err := userService.Register(&models.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	}, "user")
	require.NoError(t, err)

	currentSession, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  "current-browser",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	otherSession, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  "other-browser",
		IPAddress:  "127.0.0.2",
		RememberMe: false,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	token, err := jwtService.GenerateToken(user, currentSession.ID)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(jwtService, sessionService))
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
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var responseBody struct {
		Message              string `json:"message"`
		RevokedOtherSessions int    `json:"revoked_other_sessions"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
	assert.Equal(t, "密码修改成功", responseBody.Message)
	assert.Equal(t, 1, responseBody.RevokedOtherSessions)

	_, err = sessionService.ValidateSession(user.ID, currentSession.ID)
	require.NoError(t, err)

	_, err = sessionService.ValidateSession(user.ID, otherSession.ID)
	assert.ErrorIs(t, err, services.ErrSessionRevoked)

	updatedUser, err := userService.Login(&models.LoginRequest{
		Username: "alice",
		Password: "updated456",
	})
	require.NoError(t, err)
	assert.Equal(t, user.ID, updatedUser.ID)
}
