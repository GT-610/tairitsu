package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.SecurityHeaders())
	app.Get("/headers", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/headers", nil))
	require.NoError(t, err)

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.Equal(t, "camera=(), microphone=(), geolocation=()", resp.Header.Get("Permissions-Policy"))
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "frame-ancestors 'none'")
	assert.Empty(t, resp.Header.Get("Strict-Transport-Security"))
}
