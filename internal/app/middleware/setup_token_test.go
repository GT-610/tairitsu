package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTokenAndApp(t *testing.T) (*services.SetupToken, *fiber.App) {
	t.Helper()

	token, err := services.NewSetupToken()
	require.NoError(t, err)

	app := fiber.New()
	app.Use(SetupTokenRequired(token))
	app.Get("/setup", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"authorized": SetupAuthorized(c)})
	})

	return token, app
}

func TestSetupTokenRequiredAcceptsTheIssuedToken(t *testing.T) {
	token, app := newTokenAndApp(t)

	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	req.Header.Set(SetupTokenHeader, token.Value())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetupTokenRequiredToleratesSurroundingWhitespace(t *testing.T) {
	token, app := newTokenAndApp(t)

	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	req.Header.Set(SetupTokenHeader, "  "+token.Value()+"\n")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetupTokenRequiredRejectsMissingOrWrongTokens(t *testing.T) {
	token, app := newTokenAndApp(t)

	for _, candidate := range []string{"", "wrong", token.Value() + "x", token.Value()[:8]} {
		name := candidate
		if name == "" {
			name = "empty"
		}

		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/setup", nil)
			if candidate != "" {
				req.Header.Set(SetupTokenHeader, candidate)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// 403 and not 401: a 401 would make the client clear its session and
			// send the operator to the login page.
			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}
}

func TestSetupTokenOptionalNeverBlocksRequests(t *testing.T) {
	token, err := services.NewSetupToken()
	require.NoError(t, err)

	app := fiber.New()
	app.Use(SetupTokenOptional(token))
	app.Get("/status", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"authorized": SetupAuthorized(c)})
	})

	anonymous := httptest.NewRequest(http.MethodGet, "/status", nil)
	anonymousResp, err := app.Test(anonymous)
	require.NoError(t, err)
	defer anonymousResp.Body.Close()
	assert.Equal(t, fiber.StatusOK, anonymousResp.StatusCode)
	assert.Contains(t, readBody(t, anonymousResp), `"authorized":false`)

	authorized := httptest.NewRequest(http.MethodGet, "/status", nil)
	authorized.Header.Set(SetupTokenHeader, token.Value())
	authorizedResp, err := app.Test(authorized)
	require.NoError(t, err)
	defer authorizedResp.Body.Close()
	assert.Contains(t, readBody(t, authorizedResp), `"authorized":true`)
}

// A process without a token, which is what an initialized instance looks like,
// must fail closed rather than treating every request as authorized.
func TestSetupTokenMiddlewareFailsClosedWithoutAnIssuedToken(t *testing.T) {
	app := fiber.New()
	app.Use(SetupTokenRequired(nil))
	app.Get("/setup", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	buffer := make([]byte, 256)
	count, _ := resp.Body.Read(buffer)
	return string(buffer[:count])
}
