package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// app.Test drives requests without a real connection, so Fiber reports every
// caller as the same placeholder address unless a proxy is trusted. Trusting
// one here is what makes the client address controllable, and it is also how a
// deployment behind a reverse proxy presents itself.
func newLimitedApp(t *testing.T, limiter fiber.Handler) *fiber.App {
	t.Helper()

	app := fiber.New(fiber.Config{
		TrustProxy:       true,
		ProxyHeader:      "X-Forwarded-For",
		TrustProxyConfig: fiber.TrustProxyConfig{Proxies: []string{"0.0.0.0/0", "::/0"}},
	})
	app.Use(limiter)
	app.Get("/limited", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	return app
}

// Each test spends a distinct client address so that the process-wide limiters
// keep their budgets to themselves.
var rateLimitTestAddress = 0

func nextAddress(t *testing.T) string {
	t.Helper()

	rateLimitTestAddress++
	// 198.51.100.0/24 is reserved for documentation, so it cannot collide with
	// an address a real caller would have.
	return fmt.Sprintf("198.51.100.%d", rateLimitTestAddress%250+1)
}

func requestFrom(t *testing.T, app *fiber.App, address string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.Header.Set("X-Forwarded-For", address)

	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func spendBudget(t *testing.T, app *fiber.App, address string, requests int) {
	t.Helper()

	for request := 0; request < requests; request++ {
		resp := requestFrom(t, app, address)
		require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		resp.Body.Close()
	}
}

// Two clients must not share a budget, or one of them can lock the other out.
func TestRateLimitMiddlewareKeepsClientBudgetsSeparate(t *testing.T) {
	app := newLimitedApp(t, middleware.RateLimit())

	first := nextAddress(t)
	second := nextAddress(t)

	// The API budget is 100 tokens, so the 101st request is the one refused.
	spendBudget(t, app, first, 100)

	exhausted := requestFrom(t, app, first)
	defer exhausted.Body.Close()
	require.Equal(t, fiber.StatusTooManyRequests, exhausted.StatusCode)

	other := requestFrom(t, app, second)
	defer other.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, other.StatusCode, "an unrelated client must still be served")
}

func TestRateLimitMiddlewareReportsAStableErrorCode(t *testing.T) {
	app := newLimitedApp(t, middleware.RateLimit())
	address := nextAddress(t)

	spendBudget(t, app, address, 100)

	resp := requestFrom(t, app, address)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "system.rate_limited", body["error_code"])
	assert.Equal(t, "Too many requests. Please try again later.", body["message"])
}

// The authentication surface has a much smaller budget than the API surface.
func TestAuthRateLimitMiddlewareUsesTheSmallerBudget(t *testing.T) {
	app := newLimitedApp(t, middleware.AuthRateLimit())
	address := nextAddress(t)

	spendBudget(t, app, address, 20)

	resp := requestFrom(t, app, address)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}
