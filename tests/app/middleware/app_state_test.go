package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	appmiddleware "github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

type stateStub struct {
	initialized bool
}

func (s stateStub) IsInitialized() bool {
	return s.initialized
}

func TestSetupOnly_BlocksAfterInitialization(t *testing.T) {
	router := fiber.New()
	router.Use(appmiddleware.SetupOnlyWithState(stateStub{initialized: true}))
	router.Post("/setup-only", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/setup-only", nil)
	resp, err := router.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "系统已初始化")
}

func TestInitializedOnly_BlocksBeforeInitialization(t *testing.T) {
	router := fiber.New()
	router.Use(appmiddleware.InitializedOnlyWithState(stateStub{initialized: false}))
	router.Get("/runtime-only", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/runtime-only", nil)
	resp, err := router.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "系统尚未初始化")
}
