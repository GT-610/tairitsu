package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type requestIdentity struct {
	IP     string `json:"ip"`
	Scheme string `json:"scheme"`
}

func addIdentityRoute(app *fiber.App) {
	app.Get("/identity", func(c fiber.Ctx) error {
		return c.JSON(requestIdentity{IP: c.IP(), Scheme: c.Scheme()})
	})
}

func newTrustedTestProxyApp() *fiber.App {
	config := httpAppConfig()
	config.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: []string{"0.0.0.0/0", "::/0"}}
	return fiber.New(config)
}

func TestHTTPAppConfigTrustsOnlyLoopbackProxies(t *testing.T) {
	config := httpAppConfig()

	assert.True(t, config.TrustProxy)
	assert.Equal(t, "X-Real-IP", config.ProxyHeader)
	assert.True(t, config.TrustProxyConfig.Loopback)
	assert.False(t, config.TrustProxyConfig.Private)
	assert.True(t, config.EnableIPValidation)
}

func TestHTTPAppTrustsConfiguredProxyHeaders(t *testing.T) {
	app := newTrustedTestProxyApp()
	addIdentityRoute(app)

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	req.Header.Set("X-Real-IP", "203.0.113.10")
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var identity requestIdentity
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&identity))
	assert.Equal(t, "203.0.113.10", identity.IP)
	assert.Equal(t, "https", identity.Scheme)
}

func TestHTTPAppIgnoresProxyHeadersFromUntrustedClients(t *testing.T) {
	app := newHTTPApp()
	addIdentityRoute(app)

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	req.Header.Set("X-Real-IP", "203.0.113.99")
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var identity requestIdentity
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&identity))
	assert.Equal(t, "0.0.0.0", identity.IP)
	assert.Equal(t, "http", identity.Scheme)
}

func TestHTTPAppRateLimitsForwardedClientsIndependently(t *testing.T) {
	app := newTrustedTestProxyApp()
	app.Use(func(c fiber.Ctx) error {
		c.Set("X-Test-Observed-IP", c.IP())
		return c.Next()
	})
	app.Use(middleware.RateLimitWithLimiter(middleware.NewRateLimiter(1, 0)))
	app.Get("/limited", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	request := func(clientIP string) *http.Response {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.Header.Set("X-Real-IP", clientIP)
		resp, err := app.Test(req)
		require.NoError(t, err)
		return resp
	}

	first := request("203.0.113.10")
	second := request("203.0.113.11")
	third := request("203.0.113.10")
	defer first.Body.Close()
	defer second.Body.Close()
	defer third.Body.Close()

	assert.Equal(t, "203.0.113.10", first.Header.Get("X-Test-Observed-IP"))
	assert.Equal(t, "203.0.113.11", second.Header.Get("X-Test-Observed-IP"))
	assert.Equal(t, "203.0.113.10", third.Header.Get("X-Test-Observed-IP"))
	assert.Equal(t, fiber.StatusNoContent, first.StatusCode)
	assert.Equal(t, fiber.StatusNoContent, second.StatusCode)
	assert.Equal(t, fiber.StatusTooManyRequests, third.StatusCode)
}

func TestHTTPAppSetsHSTSForTrustedHTTPSProxy(t *testing.T) {
	app := newTrustedTestProxyApp()
	app.Use(middleware.SecurityHeaders())
	app.Get("/headers", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/headers", nil)
	req.Header.Set("X-Real-IP", "203.0.113.10")
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header.Get("Strict-Transport-Security"))
}
