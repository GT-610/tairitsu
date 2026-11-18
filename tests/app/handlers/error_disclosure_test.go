package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apphandlers "github.com/GT-610/tempest-zt/internal/app/handlers"
	"github.com/GT-610/tempest-zt/internal/app/middleware"
	"github.com/GT-610/tempest-zt/internal/app/models"
	"github.com/GT-610/tempest-zt/internal/app/services"
	"github.com/GT-610/tempest-zt/internal/zerotier"
	"github.com/GT-610/tempest-zt/tests/testutil"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newNetworkHandlerApp wires the networks surface behind a session and returns
// the credential a browser would hold, so a test exercises the handler the way a
// signed-in caller reaches it.
func newNetworkHandlerApp(t *testing.T, ztClient *zerotier.Client) (*fiber.App, string) {
	t.Helper()

	db := testutil.NewSQLiteDB(t)
	require.NoError(t, db.CreateUser(&models.User{
		ID:        "user-1",
		Username:  "alice",
		Password:  "hashed-password",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))

	networkService := services.NewNetworkService(ztClient, db)
	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	networkHandler := apphandlers.NewNetworkHandler(networkService)

	_, credential := newTestSession(t, sessionService, "user-1", "test-browser", "127.0.0.1")

	app := fiber.New()
	app.Use(middleware.ErrorHandler())
	app.Use(middleware.AuthMiddleware(sessionService, userService))
	app.Get("/networks", networkHandler.GetNetworks)
	app.Post("/networks", networkHandler.CreateNetwork)

	return app, credential
}

func decodeBody(t *testing.T, resp *http.Response) (string, map[string]any) {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload), "body: %s", string(body))

	return string(body), payload
}

func postCreateNetwork(t *testing.T, app *fiber.App, credential, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/networks", bytes.NewBufferString(body))
	authenticateRequest(req, credential)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, testRequestTimeout)
	require.NoError(t, err)
	return resp
}

// TestNetworkHandler_BindErrorDoesNotNameGoTypes covers the decoder's habit of
// naming the type and field it was reading. That describes the server, not the
// caller's mistake.
func TestNetworkHandler_BindErrorDoesNotNameGoTypes(t *testing.T) {
	app, credential := newNetworkHandlerApp(t, nil)

	resp := postCreateNetwork(t, app, credential, `{"name":123}`)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	raw, payload := decodeBody(t, resp)
	assert.Equal(t, "system.invalid_request", payload["error_code"])
	assert.Equal(t, "Invalid request body", payload["message"])

	// The leak this guards is the type and field name in the decoder's message.
	assert.NotContains(t, raw, "Go struct field")
	assert.NotContains(t, raw, "zerotier.Network")
	assert.NotContains(t, raw, "cannot unmarshal")
}

// TestNetworkHandler_ControllerFailureDoesNotLeakTheControllerAddress is the
// regression test for an internal-topology disclosure: a failed controller call
// used to be echoed to the caller verbatim, handing out the address, the port
// and the dial error.
func TestNetworkHandler_ControllerFailureDoesNotLeakTheControllerAddress(t *testing.T) {
	// A closed listener gives a deterministic connection refusal without a
	// hard-coded port that might be in use.
	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	ztClient := &zerotier.Client{
		BaseURL:    deadURL,
		Token:      "test-token",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	app, credential := newNetworkHandlerApp(t, ztClient)

	resp := postCreateNetwork(t, app, credential, `{"name":"alpha"}`)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	raw, payload := decodeBody(t, resp)
	assert.Equal(t, "system.internal_error", payload["error_code"])
	assert.Equal(t, "Internal Server Error", payload["message"])

	assert.NotContains(t, raw, "127.0.0.1")
	assert.NotContains(t, raw, "dial ")
	assert.NotContains(t, raw, "connection refused")
	assert.NotContains(t, raw, "failed to send request")
}

// TestNetworkHandler_ControllerFailureDoesNotEchoTheResponseBody covers the
// other half: the client used to put the controller's whole response body into
// the error, up to its 10 MiB read limit.
func TestNetworkHandler_ControllerFailureDoesNotEchoTheResponseBody(t *testing.T) {
	const marker = "controller-internal-detail-marker"

	controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, marker, http.StatusInternalServerError)
	}))
	t.Cleanup(controller.Close)

	ztClient := &zerotier.Client{
		BaseURL:    controller.URL,
		Token:      "test-token",
		HTTPClient: controller.Client(),
	}

	app, credential := newNetworkHandlerApp(t, ztClient)

	resp := postCreateNetwork(t, app, credential, `{"name":"alpha"}`)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	raw, payload := decodeBody(t, resp)
	assert.Equal(t, "system.internal_error", payload["error_code"])
	assert.NotContains(t, raw, marker)
}
