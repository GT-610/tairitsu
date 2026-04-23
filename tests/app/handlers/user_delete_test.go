package handlers

import (
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

func TestUserHandler_DeleteUser_AdminOnlyAndResponseShape(t *testing.T) {
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	userService := services.NewUserServiceWithDB(db)
	jwtService := services.NewJWTService("test-secret")
	sessionService := services.NewSessionServiceWithDB(db)
	userHandler := apphandlers.NewUserHandler(userService)

	admin, err := userService.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	target, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "net-1",
		Name:        "alice-network",
		Description: "test network",
		OwnerID:     target.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}))
	require.NoError(t, db.CreateSession(&models.Session{
		ID:         "session-1",
		UserID:     target.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}))

	adminSession, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     admin.ID,
		UserAgent:  "admin-browser",
		IPAddress:  "127.0.0.2",
		RememberMe: true,
		ExpiresAt:  now.Add(time.Hour),
	})
	require.NoError(t, err)

	adminToken, err := jwtService.GenerateToken(admin, adminSession.ID)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(jwtService, sessionService))
	app.Delete("/users/:userId", middleware.AdminRequiredWithUserService(userService), userHandler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+target.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var responseBody struct {
		Message             string              `json:"message"`
		User                models.UserResponse `json:"user"`
		TransferredNetworks int                 `json:"transferred_networks"`
		RevokedSessions     int                 `json:"revoked_sessions"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&responseBody))
	assert.Equal(t, target.ID, responseBody.User.ID)
	assert.Equal(t, 1, responseBody.TransferredNetworks)
	assert.Equal(t, 1, responseBody.RevokedSessions)
	assert.NotEmpty(t, responseBody.Message)
}
