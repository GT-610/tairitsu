package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	apphandlers "github.com/GT-610/tairitsu/internal/app/handlers"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handlerStateDBStub struct {
	users []*models.User
}

func (s *handlerStateDBStub) Init() error { return nil }
func (s *handlerStateDBStub) CreateUser(user *models.User) error {
	s.users = append(s.users, user)
	return nil
}
func (s *handlerStateDBStub) GetUserByID(id string) (*models.User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}
func (s *handlerStateDBStub) GetUserByUsername(username string) (*models.User, error) {
	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}
func (s *handlerStateDBStub) GetAllUsers() ([]*models.User, error) { return s.users, nil }
func (s *handlerStateDBStub) UpdateUser(user *models.User) error   { return nil }
func (s *handlerStateDBStub) DeleteUser(id string) error           { return nil }
func (s *handlerStateDBStub) CreateNetwork(network *models.Network) error {
	return nil
}
func (s *handlerStateDBStub) GetNetworkByID(id string) (*models.Network, error) {
	return nil, nil
}
func (s *handlerStateDBStub) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	return []*models.Network{}, nil
}
func (s *handlerStateDBStub) GetAllNetworks() ([]*models.Network, error) {
	return []*models.Network{}, nil
}
func (s *handlerStateDBStub) UpdateNetwork(network *models.Network) error { return nil }
func (s *handlerStateDBStub) DeleteNetwork(id string) error               { return nil }
func (s *handlerStateDBStub) HasAdminUser() (bool, error) {
	for _, user := range s.users {
		if user.Role == "admin" {
			return true, nil
		}
	}
	return false, nil
}
func (s *handlerStateDBStub) Close() error { return nil }

func TestSystemHandler_GetSystemStatus_Uninitialized(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{Initialized: false}

	handler := apphandlers.NewSystemHandler(services.NewNetworkService(nil, nil), services.NewUserServiceWithoutDB())

	app := fiber.New()
	app.Get("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, false, body["initialized"])
}

func TestSystemHandler_GetSystemStatus_Initialized(t *testing.T) {
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = originalConfig
	})

	config.AppConfig = &config.Config{
		Initialized: true,
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: "data/test.db",
		},
	}

	userService := services.NewUserServiceWithDB(&handlerStateDBStub{
		users: []*models.User{
			{ID: "1", Username: "admin", Role: "admin"},
		},
	})
	networkService := services.NewNetworkService(nil, nil)
	handler := apphandlers.NewSystemHandler(networkService, userService)

	app := fiber.New()
	app.Get("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body struct {
		Initialized bool `json:"initialized"`
		HasDatabase bool `json:"hasDatabase"`
		HasAdmin    bool `json:"hasAdmin"`
		ZTStatus    struct {
			Online  bool   `json:"online"`
			Version string `json:"version"`
		} `json:"ztStatus"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body.Initialized)
	assert.True(t, body.HasDatabase)
	assert.True(t, body.HasAdmin)
	assert.False(t, body.ZTStatus.Online)
	assert.Equal(t, "unknown", body.ZTStatus.Version)
}
