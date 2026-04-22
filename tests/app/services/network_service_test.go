package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkService(t *testing.T) {
	// Act
	networkService := services.NewNetworkService(nil, nil)

	// Assert
	assert.NotNil(t, networkService)
}

func TestImportableNetworkSummariesStartAsEmptySlice(t *testing.T) {
	summaries := make([]services.ImportableNetworkSummary, 0)

	assert.NotNil(t, summaries)
	assert.Len(t, summaries, 0)
}

func TestNetworkServiceImportNetworks_AllSuccess(t *testing.T) {
	db := newTestSQLiteDB(t)
	client := newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha", Description: "alpha-desc"},
		"8056c2e21c000002": {ID: "8056c2e21c000002", Name: "beta", Description: "beta-desc"},
	})

	service := services.NewNetworkService(client, db)

	result, err := service.ImportNetworks([]string{"8056c2e21c000001", "8056c2e21c000002"}, "user-1")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.ElementsMatch(t, []string{"8056c2e21c000001", "8056c2e21c000002"}, result.ImportedIDs)
	assert.Empty(t, result.Failed)

	allNetworks, err := db.GetAllNetworks()
	require.NoError(t, err)
	require.Len(t, allNetworks, 2)
	for _, network := range allNetworks {
		assert.Equal(t, "user-1", network.OwnerID)
	}
}

func TestNetworkServiceImportNetworks_PartialFailures(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000010",
		Name:        "owned-by-other",
		Description: "existing",
		OwnerID:     "other-user",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	client := newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha", Description: "alpha-desc"},
		"8056c2e21c000010": {ID: "8056c2e21c000010", Name: "owned-by-other", Description: "existing"},
	})

	service := services.NewNetworkService(client, db)

	result, err := service.ImportNetworks([]string{"8056c2e21c000001", "8056c2e21c000010", "8056c2e21c000099"}, "user-1")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"8056c2e21c000001"}, result.ImportedIDs)
	require.Len(t, result.Failed, 2)
	assert.Contains(t, result.Failed, services.ImportNetworkFailure{
		NetworkID: "8056c2e21c000010",
		Reason:    "网络属于其他用户，无法导入",
	})
	assert.Contains(t, result.Failed, services.ImportNetworkFailure{
		NetworkID: "8056c2e21c000099",
		Reason:    "网络不存在于 ZeroTier 控制器中",
	})
}

func TestNetworkServiceGetImportableNetworks_EmptySliceWhenControllerHasNoNetworks(t *testing.T) {
	db := newTestSQLiteDB(t)
	client := newTestZTClient(t, map[string]zerotier.Network{})

	service := services.NewNetworkService(client, db)

	result, err := service.GetImportableNetworks("user-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func newTestSQLiteDB(t *testing.T) *database.SQLiteDB {
	t.Helper()

	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func newTestZTClient(t *testing.T, networks map[string]zerotier.Network) *zerotier.Client {
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
			network, ok := networks[r.URL.Path[len("/controller/network/"):]]
			if !ok || len(r.URL.Path) <= len("/controller/network/") {
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
