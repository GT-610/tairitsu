package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, db *database.SQLiteDB, id string, role string) {
	t.Helper()

	require.NoError(t, db.CreateUser(&models.User{
		ID:        id,
		Username:  id,
		Password:  "hashed-password",
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))
}

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
	createTestUser(t, db, "user-1", "user")
	client := newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha", Description: "alpha-desc"},
		"8056c2e21c000002": {ID: "8056c2e21c000002", Name: "beta", Description: "beta-desc"},
	})

	service := services.NewNetworkService(client, db)

	result, err := service.ImportNetworks([]string{"8056c2e21c000001", "8056c2e21c000002"}, "user-1", "admin")
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
	createTestUser(t, db, "user-1", "user")
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

	result, err := service.ImportNetworks([]string{"8056c2e21c000001", "8056c2e21c000010", "8056c2e21c000099"}, "user-1", "admin")
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

	result, err := service.GetImportableNetworks()
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestNetworkServiceImportNetworks_RejectsNonAdmin(t *testing.T) {
	db := newTestSQLiteDB(t)
	createTestUser(t, db, "user-1", "user")
	client := newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha"},
	})

	service := services.NewNetworkService(client, db)

	result, err := service.ImportNetworks([]string{"8056c2e21c000001"}, "user-1", "user")
	require.ErrorIs(t, err, services.ErrImportAccessDenied)
	assert.Nil(t, result)
}

func TestNetworkServiceImportNetworks_RequiresExistingOwner(t *testing.T) {
	db := newTestSQLiteDB(t)
	client := newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha"},
	})

	service := services.NewNetworkService(client, db)

	result, err := service.ImportNetworks([]string{"8056c2e21c000001"}, "missing-owner", "admin")
	require.ErrorIs(t, err, services.ErrImportOwnerNotFound)
	assert.Nil(t, result)
}

func TestNetworkServiceGetImportableNetworks_MarksOwnedNetworksAsNotImportable(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000001",
		Name:        "owned",
		Description: "owned-desc",
		OwnerID:     "user-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))
	client := newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "owned"},
		"8056c2e21c000002": {ID: "8056c2e21c000002", Name: "orphan"},
	})

	service := services.NewNetworkService(client, db)

	result, err := service.GetImportableNetworks()
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Contains(t, result, services.ImportableNetworkSummary{
		NetworkID:    "8056c2e21c000001",
		Reason:       "网络已有所有者",
		IsImportable: false,
	})
	assert.Contains(t, result, services.ImportableNetworkSummary{
		NetworkID:    "8056c2e21c000002",
		Reason:       "网络尚未登记到 Tairitsu",
		IsImportable: true,
	})
}

func TestNetworkServiceGetAllNetworksIncludesMemberStats(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000001",
		Name:        "alpha",
		Description: "alpha-desc",
		OwnerID:     "user-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	client := newTestZTClientWithMembers(t,
		map[string]zerotier.Network{
			"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha", Description: "alpha-desc"},
		},
		map[string][]zerotier.Member{
			"8056c2e21c000001": {
				{ID: "member-1", Config: zerotier.MemberConfig{Authorized: true}},
				{ID: "member-2", Config: zerotier.MemberConfig{Authorized: true}},
				{ID: "member-3", Config: zerotier.MemberConfig{Authorized: false}},
			},
		},
	)

	service := services.NewNetworkService(client, db)

	summaries, err := service.GetAllNetworks("user-1")
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Equal(t, 3, summaries[0].MemberCount)
	assert.Equal(t, 2, summaries[0].AuthorizedMemberCount)
	assert.Equal(t, 1, summaries[0].PendingMemberCount)
}

func TestNetworkServiceGetNetworkByIDIncludesMembers(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "f76fd3000b86b177",
		Name:        "detail-network",
		Description: "detail-network-desc",
		OwnerID:     "user-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	client := newTestZTClientWithMemberIndexResponse(t,
		map[string]zerotier.Network{
			"f76fd3000b86b177": {ID: "f76fd3000b86b177", Name: "detail-network"},
		},
		map[string]map[string]int{
			"f76fd3000b86b177": {
				"8789af2692": 1,
			},
		},
		map[string]zerotier.Member{
			"8789af2692": {ID: "8789af2692", Name: "pending-node", Config: zerotier.MemberConfig{Authorized: false}},
		},
	)

	service := services.NewNetworkService(client, db)

	detail, err := service.GetNetworkByID("f76fd3000b86b177", "user-1")
	require.NoError(t, err)
	require.NotNil(t, detail)
	require.Len(t, detail.Members, 1)
	assert.Equal(t, "8789af2692", detail.Members[0].ID)
	assert.False(t, detail.Members[0].Config.Authorized)
}

func TestNetworkServiceGetAllNetworksIncludesMemberStatsFromObjectResponse(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000009",
		Name:        "object-members",
		Description: "object-members-desc",
		OwnerID:     "user-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	client := newTestZTClientWithMemberObjectResponse(t,
		map[string]zerotier.Network{
			"8056c2e21c000009": {ID: "8056c2e21c000009", Name: "object-members"},
		},
		map[string]map[string]zerotier.Member{
			"8056c2e21c000009": {
				"member-b": {Config: zerotier.MemberConfig{Authorized: false}},
				"member-a": {Config: zerotier.MemberConfig{Authorized: true}},
			},
		},
	)

	service := services.NewNetworkService(client, db)

	summaries, err := service.GetAllNetworks("user-1")
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Equal(t, 2, summaries[0].MemberCount)
	assert.Equal(t, 1, summaries[0].AuthorizedMemberCount)
	assert.Equal(t, 1, summaries[0].PendingMemberCount)
}

func TestZTClientGetMembersReturnsPreviewOnUnexpectedShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
			"unexpected": "not-a-member-object",
		}))
	}))
	t.Cleanup(server.Close)

	client := &zerotier.Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: server.Client(),
	}

	members, err := client.GetMembers("8056c2e21c000009")
	require.Nil(t, members)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "响应预览")
	assert.Contains(t, err.Error(), "\"unexpected\"")
}

func TestNetworkServiceGetAllNetworksIncludesMemberStatsFromIndexResponse(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "f76fd3000b86b177",
		Name:        "index-members",
		Description: "index-members-desc",
		OwnerID:     "user-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	client := newTestZTClientWithMemberIndexResponse(t,
		map[string]zerotier.Network{
			"f76fd3000b86b177": {ID: "f76fd3000b86b177", Name: "index-members"},
		},
		map[string]map[string]int{
			"f76fd3000b86b177": {
				"8789af2692": 1,
				"1234567890": 2,
			},
		},
		map[string]zerotier.Member{
			"8789af2692": {ID: "8789af2692", Config: zerotier.MemberConfig{Authorized: false}},
			"1234567890": {ID: "1234567890", Config: zerotier.MemberConfig{Authorized: true}},
		},
	)

	service := services.NewNetworkService(client, db)

	summaries, err := service.GetAllNetworks("user-1")
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Equal(t, 2, summaries[0].MemberCount)
	assert.Equal(t, 1, summaries[0].AuthorizedMemberCount)
	assert.Equal(t, 1, summaries[0].PendingMemberCount)
}

func TestNetworkServiceGetAllNetworksFetchesMemberStatsConcurrently(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("8056c2e21c00000%d", i)
		require.NoError(t, db.CreateNetwork(&models.Network{
			ID:          id,
			Name:        fmt.Sprintf("net-%d", i),
			Description: "desc",
			OwnerID:     "user-1",
			CreatedAt:   now,
			UpdatedAt:   now,
		}))
	}

	var activeRequests atomic.Int32
	var maxConcurrent atomic.Int32
	client := newDelayedMemberZTClient(t, 150*time.Millisecond, &activeRequests, &maxConcurrent)
	service := services.NewNetworkService(client, db)

	start := time.Now()
	summaries, err := service.GetAllNetworks("user-1")
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.Len(t, summaries, 5)
	assert.GreaterOrEqual(t, maxConcurrent.Load(), int32(2))
	assert.Less(t, elapsed, 600*time.Millisecond)
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
	return newTestZTClientWithMembers(t, networks, map[string][]zerotier.Member{})
}

func newTestZTClientWithMembers(t *testing.T, networks map[string]zerotier.Network, members map[string][]zerotier.Member) *zerotier.Client {
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
			for networkID, memberList := range members {
				if path == networkID+"/member" {
					require.NoError(t, json.NewEncoder(w).Encode(memberList))
					return
				}
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

func newDelayedMemberZTClient(t *testing.T, delay time.Duration, activeRequests *atomic.Int32, maxConcurrent *atomic.Int32) *zerotier.Client {
	t.Helper()

	networks := map[string]zerotier.Network{}
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("8056c2e21c00000%d", i)
		networks[id] = zerotier.Network{ID: id, Name: fmt.Sprintf("net-%d", i)}
	}

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
			for networkID := range networks {
				if path == networkID+"/member" {
					current := activeRequests.Add(1)
					for {
						observed := maxConcurrent.Load()
						if current <= observed || maxConcurrent.CompareAndSwap(observed, current) {
							break
						}
					}
					time.Sleep(delay)
					activeRequests.Add(-1)
					require.NoError(t, json.NewEncoder(w).Encode([]zerotier.Member{
						{ID: "member-1", Config: zerotier.MemberConfig{Authorized: true}},
					}))
					return
				}
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

func newTestZTClientWithMemberObjectResponse(t *testing.T, networks map[string]zerotier.Network, members map[string]map[string]zerotier.Member) *zerotier.Client {
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
			for networkID, memberMap := range members {
				if path == networkID+"/member" {
					require.NoError(t, json.NewEncoder(w).Encode(memberMap))
					return
				}
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

func newTestZTClientWithMemberIndexResponse(t *testing.T, networks map[string]zerotier.Network, memberIndexes map[string]map[string]int, memberDetails map[string]zerotier.Member) *zerotier.Client {
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

			for networkID, memberIndex := range memberIndexes {
				if path == networkID+"/member" {
					require.NoError(t, json.NewEncoder(w).Encode(memberIndex))
					return
				}
			}

			for memberID, member := range memberDetails {
				for networkID := range memberIndexes {
					if path == networkID+"/member/"+memberID {
						require.NoError(t, json.NewEncoder(w).Encode(member))
						return
					}
				}
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
