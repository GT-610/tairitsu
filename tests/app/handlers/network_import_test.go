package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkHandler_ImportConsoleAndImportResponseShape(t *testing.T) {
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

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

	ztClient := newImportHandlerTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha", Description: "alpha-desc", Status: "OK"},
	})

	networkService := services.NewNetworkService(ztClient, db)
	userService := services.NewUserServiceWithDB(db)
	jwtService := services.NewJWTService("test-secret")
	sessionService := services.NewSessionServiceWithDB(db)
	networkHandler := apphandlers.NewNetworkHandler(networkService)

	adminSession, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     "admin-1",
		UserAgent:  "admin-browser",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	admin, err := userService.GetUserByID("admin-1")
	require.NoError(t, err)
	adminToken, err := jwtService.GenerateToken(admin, adminSession.ID)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(jwtService, sessionService))
	app.Get("/admin/networks/importable", middleware.AdminRequiredWithUserService(userService), networkHandler.GetImportableNetworks)
	app.Post("/admin/networks/import", middleware.AdminRequiredWithUserService(userService), networkHandler.ImportNetworks)

	getReq := httptest.NewRequest(http.MethodGet, "/admin/networks/importable", nil)
	getReq.Header.Set("Authorization", "Bearer "+adminToken)
	getResp, err := app.Test(getReq)
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
	postReq.Header.Set("Authorization", "Bearer "+adminToken)
	postReq.Header.Set("Content-Type", "application/json")
	postResp, err := app.Test(postReq)
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

func newImportHandlerTestZTClient(t *testing.T, networks map[string]zerotier.Network) *zerotier.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/controller/network":
			ids := make([]string, 0, len(networks))
			for id := range networks {
				ids = append(ids, id)
			}
			require.NoError(t, json.NewEncoder(w).Encode(ids))
		default:
			if len(r.URL.Path) <= len("/controller/network/") {
				http.NotFound(w, r)
				return
			}
			path := r.URL.Path[len("/controller/network/"):]
			if len(path) > len("/member") && path[len(path)-len("/member"):] == "/member" {
				require.NoError(t, json.NewEncoder(w).Encode([]zerotier.Member{}))
				return
			}

			network, ok := networks[path]
			if !ok {
				http.NotFound(w, r)
				return
			}
			require.NoError(t, json.NewEncoder(w).Encode(network))
		}
	}))
	t.Cleanup(server.Close)

	return &zerotier.Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: server.Client(),
	}
}
