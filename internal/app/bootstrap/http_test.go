package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type requestIdentity struct {
	IP     string `json:"ip"`
	Scheme string `json:"scheme"`
	Secure bool   `json:"secure"`
}

func addIdentityRoute(app *fiber.App) {
	app.Get("/identity", func(c fiber.Ctx) error {
		return c.JSON(requestIdentity{
			IP:     c.IP(),
			Scheme: c.Scheme(),
			Secure: middleware.IsSecureRequest(c),
		})
	})
}

// newTrustedTestProxyApp trusts every peer, which is what lets the Test request
// itself stand in for a reverse proxy.
func newTrustedTestProxyApp() *fiber.App {
	appConfig := httpAppConfig()
	appConfig.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: []string{"0.0.0.0/0", "::/0"}}
	return fiber.New(appConfig)
}

func identityOf(t *testing.T, app *fiber.App, headers map[string]string) requestIdentity {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var identity requestIdentity
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&identity))
	return identity
}

func TestHTTPAppConfigTrustsOnlyLoopbackProxiesByDefault(t *testing.T) {
	appConfig := httpAppConfig()

	assert.True(t, appConfig.TrustProxy)
	assert.Equal(t, "X-Forwarded-For", appConfig.ProxyHeader)
	assert.True(t, appConfig.EnableIPValidation)

	// The trust set is the whole security decision here: every flag that widens
	// it lets a peer forge the client address, and with it the per-IP rate limit
	// and the address recorded on a session. Broadening any of them
	// unintentionally is the mistake this guards, so each is asserted rather
	// than only the one that looks dangerous today.
	assert.True(t, appConfig.TrustProxyConfig.Loopback)
	assert.Empty(t, appConfig.TrustProxyConfig.Proxies)
	assert.False(t, appConfig.TrustProxyConfig.Private)
	assert.False(t, appConfig.TrustProxyConfig.LinkLocal)
}

func TestHTTPAppConfigReadsTrustedProxiesFromTheEnvironment(t *testing.T) {
	t.Setenv(TrustedProxiesEnv, "172.17.0.1, 10.0.0.0/8 not-an-address")
	t.Setenv(ProxyHeaderEnv, "X-Real-IP")

	appConfig := httpAppConfig()

	assert.Equal(t, "X-Real-IP", appConfig.ProxyHeader)
	assert.Equal(t, []string{"172.17.0.1", "10.0.0.0/8"}, appConfig.TrustProxyConfig.Proxies)
	assert.True(t, appConfig.TrustProxyConfig.Loopback)
}

func TestHTTPAppConfigCanStopTrustingLoopback(t *testing.T) {
	t.Setenv(TrustLoopbackProxyEnv, "false")

	assert.False(t, httpAppConfig().TrustProxyConfig.Loopback)
}

func TestHTTPAppConfigKeepsTheLoopbackDefaultForUnparsableValues(t *testing.T) {
	t.Setenv(TrustLoopbackProxyEnv, "yes-please")

	assert.True(t, httpAppConfig().TrustProxyConfig.Loopback)
}

func TestHTTPAppTrustsConfiguredProxyHeaders(t *testing.T) {
	app := newTrustedTestProxyApp()
	addIdentityRoute(app)

	identity := identityOf(t, app, map[string]string{
		"X-Forwarded-For":   "203.0.113.10",
		"X-Forwarded-Proto": "https",
	})

	assert.Equal(t, "203.0.113.10", identity.IP)
	assert.Equal(t, "https", identity.Scheme)
	assert.True(t, identity.Secure)
}

func TestHTTPAppIgnoresProxyHeadersFromUntrustedClients(t *testing.T) {
	app := newHTTPApp()
	addIdentityRoute(app)

	identity := identityOf(t, app, map[string]string{
		"X-Forwarded-For":   "203.0.113.99",
		"X-Forwarded-Proto": "https",
	})

	assert.Equal(t, "0.0.0.0", identity.IP)
	assert.Equal(t, "http", identity.Scheme)
	// A client that is not a trusted proxy must not be able to claim HTTPS:
	// that would silently drop the Secure flag from the session cookie.
	assert.False(t, identity.Secure)
}

// The address a trusted proxy reports is what the rate limiter keys on, so two
// clients behind that proxy must not share one budget.
func TestHTTPAppRateLimitsForwardedClientsIndependently(t *testing.T) {
	app := newTrustedTestProxyApp()
	app.Use(func(c fiber.Ctx) error {
		c.Set("X-Test-Observed-IP", c.IP())
		return c.Next()
	})
	app.Use(middleware.RateLimit())
	app.Get("/limited", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	request := func(clientIP string) *http.Response {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.Header.Set("X-Forwarded-For", clientIP)
		resp, err := app.Test(req)
		require.NoError(t, err)
		return resp
	}

	// The API budget is 100 tokens, so the 101st request from one address is the
	// one that must be refused.
	for requestNumber := 0; requestNumber < 100; requestNumber++ {
		resp := request("203.0.113.10")
		require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		resp.Body.Close()
	}

	exhausted := request("203.0.113.10")
	defer exhausted.Body.Close()
	require.Equal(t, fiber.StatusTooManyRequests, exhausted.StatusCode)
	assert.Equal(t, "203.0.113.10", exhausted.Header.Get("X-Test-Observed-IP"))

	other := request("203.0.113.11")
	defer other.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, other.StatusCode)
	assert.Equal(t, "203.0.113.11", other.Header.Get("X-Test-Observed-IP"))
}

// A client behind a proxy that is trusted but does not report a scheme must not
// be treated as HTTPS either.
func TestHTTPAppTreatsForwardedPlainHTTPAsInsecure(t *testing.T) {
	app := newTrustedTestProxyApp()
	addIdentityRoute(app)

	identity := identityOf(t, app, map[string]string{
		"X-Forwarded-For":   "203.0.113.10",
		"X-Forwarded-Proto": "http",
	})

	assert.False(t, identity.Secure)
	assert.Equal(t, "http", identity.Scheme)
}

// The session cookie's Secure flag is derived from IsSecureRequest, so the two
// must agree for every combination of proxy trust and reported scheme.
func TestHTTPAppMarksSessionCookiesSecureOnlyForTrustedHTTPS(t *testing.T) {
	newCookieApp := func(trusted bool) *fiber.App {
		appConfig := httpAppConfig()
		if trusted {
			appConfig.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: []string{"0.0.0.0/0", "::/0"}}
		}
		app := fiber.New(appConfig)
		app.Get("/login", func(c fiber.Ctx) error {
			c.Cookie(middleware.SessionCookie("credential", time.Now().Add(time.Hour), true, middleware.IsSecureRequest(c)))
			return c.SendStatus(fiber.StatusNoContent)
		})
		return app
	}

	cases := []struct {
		name       string
		trusted    bool
		scheme     string
		wantSecure bool
	}{
		{name: "trusted proxy reporting https", trusted: true, scheme: "https", wantSecure: true},
		{name: "trusted proxy reporting http", trusted: true, scheme: "http", wantSecure: false},
		{name: "untrusted client claiming https", trusted: false, scheme: "https", wantSecure: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newCookieApp(testCase.trusted)

			req := httptest.NewRequest(http.MethodGet, "/login", nil)
			if testCase.trusted {
				req.Header.Set("X-Forwarded-For", "203.0.113.10")
			}
			req.Header.Set("X-Forwarded-Proto", testCase.scheme)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			cookies := resp.Cookies()
			require.Len(t, cookies, 1)
			assert.Equal(t, middleware.SessionCookieName, cookies[0].Name)
			assert.Equal(t, testCase.wantSecure, cookies[0].Secure)
		})
	}
}

func TestHTTPAppSetsHSTSForTrustedHTTPSProxy(t *testing.T) {
	app := newTrustedTestProxyApp()
	app.Use(middleware.SecurityHeaders())
	app.Get("/headers", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/headers", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header.Get("Strict-Transport-Security"))
}

func TestHTTPAppOmitsHSTSForUntrustedHTTPSClaim(t *testing.T) {
	app := newHTTPApp()
	app.Use(middleware.SecurityHeaders())
	app.Get("/headers", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/headers", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Empty(t, resp.Header.Get("Strict-Transport-Security"))
}
