package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/logger"
	appmiddleware "github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRecoveringApp assembles the middleware pair the router uses: the error
// handler outside, the recoverer inside it. A panic therefore becomes an error
// that travels back out through the error handler.
func newRecoveringApp() *fiber.App {
	logger.InitLogger("error")

	app := fiber.New()
	app.Use(appmiddleware.ErrorHandler())
	app.Use(appmiddleware.Recover())

	app.Get("/panics", func(c fiber.Ctx) error {
		panic("boom")
	})
	app.Get("/panics-with-fiber-error", func(c fiber.Ctx) error {
		panic(fiber.NewError(fiber.StatusTeapot, "leaked detail"))
	})
	app.Get("/survives", func(c fiber.Ctx) error {
		return c.SendString("still here")
	})

	return app
}

// TestRecover_TurnsPanicIntoGenericInternalError is the regression test for a
// panicking handler taking the whole listener down.
func TestRecover_TurnsPanicIntoGenericInternalError(t *testing.T) {
	app := newRecoveringApp()

	req := httptest.NewRequest(http.MethodGet, "/panics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.Equal(t, "system.internal_error", payload["error_code"])
	assert.Equal(t, "Internal Server Error", payload["message"])
	assert.NotContains(t, string(body), "boom", "the panic value must not reach the caller")
}

// TestRecover_HidesPanicPayloadEvenWhenItLooksLikeAFiberError covers the case
// where a handler panics with something the error handler would otherwise echo:
// a *fiber.Error carries a message intended for the caller, and a panic is not a
// considered response.
func TestRecover_HidesPanicPayloadEvenWhenItLooksLikeAFiberError(t *testing.T) {
	app := newRecoveringApp()

	req := httptest.NewRequest(http.MethodGet, "/panics-with-fiber-error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "leaked detail")
}

// TestRecover_KeepsServingAfterAPanic proves the process is still usable, which
// is the actual harm the missing recoverer caused.
func TestRecover_KeepsServingAfterAPanic(t *testing.T) {
	app := newRecoveringApp()

	panicReq := httptest.NewRequest(http.MethodGet, "/panics", nil)
	panicResp, err := app.Test(panicReq)
	require.NoError(t, err)
	require.NoError(t, panicResp.Body.Close())

	nextReq := httptest.NewRequest(http.MethodGet, "/survives", nil)
	nextResp, err := app.Test(nextReq)
	require.NoError(t, err)
	defer nextResp.Body.Close()

	assert.Equal(t, fiber.StatusOK, nextResp.StatusCode)

	body, err := io.ReadAll(nextResp.Body)
	require.NoError(t, err)
	assert.Equal(t, "still here", string(body))
}
