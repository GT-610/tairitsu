package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkHandler_ImportConsoleAndImportResponseShape(t *testing.T) {
	db := testutil.NewSQLiteDB(t)

	require.NoError(t, db.CreateUser(&models.User{
		ID:        "admin-1",
		Username:  "admin",
		Password:  "hashed-password",
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))
	require.NoError(t, db.CreateUser(&models.User{
		ID:        "user-1",
		Username:  "alice",
		Password:  "hashed-password",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))

	ztClient := testutil.NewFakeController(t, &testutil.FakeController{
		Networks: map[string]zerotier.Network{
			"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha", Description: "alpha-desc", Status: "OK"},
		},
		Members: map[string][]zerotier.Member{"8056c2e21c000001": {}},
	})

	networkService := services.NewNetworkService(ztClient, db)
	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	networkHandler := apphandlers.NewNetworkHandler(networkService)

	_, adminCredential := newTestSession(t, sessionService, "admin-1", "admin-browser", "127.0.0.1")

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(sessionService, userService))
	app.Get("/admin/networks/importable", middleware.AdminRequiredWithUserService(userService), networkHandler.GetImportableNetworks)
	app.Post("/admin/networks/import", middleware.AdminRequiredWithUserService(userService), networkHandler.ImportNetworks)

	getReq := httptest.NewRequest(http.MethodGet, "/admin/networks/importable", nil)
	authenticateRequest(getReq, adminCredential)
	getResp, err := app.Test(getReq, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, getResp.StatusCode)

	var importableBody services.ImportableNetworksResult
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&importableBody))
	assert.Equal(t, services.ImportableNetworksSummary{Total: 1, Available: 1, Managed: 0, Blocked: 0}, importableBody.Summary)
	require.Len(t, importableBody.Candidates, 1)
	assert.Equal(t, "8056c2e21c000001", importableBody.Candidates[0].NetworkID)
	assert.Equal(t, "alpha", importableBody.Candidates[0].Name)
	assert.True(t, importableBody.Candidates[0].CanImport)

	postReq := httptest.NewRequest(http.MethodPost, "/admin/networks/import", bytes.NewBufferString(`{"network_ids":["8056c2e21c000001"],"owner_id":"user-1"}`))
	authenticateRequest(postReq, adminCredential)
	postReq.Header.Set("Content-Type", "application/json")
	postResp, err := app.Test(postReq, testRequestTimeout)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, postResp.StatusCode)

	var importBody services.ImportNetworksResult
	require.NoError(t, json.NewDecoder(postResp.Body).Decode(&importBody))
	assert.Equal(t, services.ImportTargetOwner{ID: "user-1", Username: "alice"}, importBody.TargetOwner)
	assert.Equal(t, services.ImportNetworksSummary{Requested: 1, Imported: 1, Failed: 0, Skipped: 0}, importBody.Summary)
	require.Len(t, importBody.Imported, 1)
	assert.Equal(t, "8056c2e21c000001", importBody.Imported[0].NetworkID)
	assert.Equal(t, "alice", importBody.Imported[0].OwnerUsername)
}
