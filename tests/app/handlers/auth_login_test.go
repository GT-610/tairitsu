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

	authHandler := apphandlers.NewAuthHandler(
		userService,
		sessionService,
		services.NewJWTService("test-secret"),
		nil,
		nil,
	)
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
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body struct {
		Session models.SessionResponse `json:"session"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "203.0.113.10", body.Session.IPAddress)
}
