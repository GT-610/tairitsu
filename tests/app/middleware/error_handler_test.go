package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appmiddleware "github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlerPreservesFiberErrorStatus(t *testing.T) {
	app := fiber.New()
	app.Use(appmiddleware.ErrorHandler())

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "Not Found", body.Error)
	assert.Equal(t, "Not Found", body.Message)
	assert.Equal(t, fiber.StatusNotFound, body.Code)
}
