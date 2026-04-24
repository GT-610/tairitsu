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
	summaries := make([]services.ImportableNetworkCandidate, 0)

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
	assert.Equal(t, services.ImportTargetOwner{ID: "user-1", Username: "user-1"}, result.TargetOwner)
	assert.ElementsMatch(t, []services.ImportNetworkResultItem{
		{NetworkID: "8056c2e21c000001", Name: "alpha", OwnerID: "user-1", OwnerUsername: "user-1"},
		{NetworkID: "8056c2e21c000002", Name: "beta", OwnerID: "user-1", OwnerUsername: "user-1"},
	}, result.Imported)
	assert.Empty(t, result.Failed)
	assert.Empty(t, result.Skipped)
	assert.Equal(t, services.ImportNetworksSummary{Requested: 2, Imported: 2, Failed: 0, Skipped: 0}, result.Summary)

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
	require.Len(t, result.Imported, 1)
	assert.Equal(t, "8056c2e21c000001", result.Imported[0].NetworkID)
	require.Len(t, result.Failed, 1)
	require.Len(t, result.Skipped, 1)
	assert.Contains(t, result.Failed, services.ImportNetworkResultItem{
		NetworkID:     "8056c2e21c000099",
		OwnerID:       "user-1",
		OwnerUsername: "user-1",
		ReasonCode:    services.ImportReasonControllerNotFound,
		ReasonMessage: "网络不存在于 ZeroTier 控制器中",
	})
	assert.Contains(t, result.Skipped, services.ImportNetworkResultItem{
		NetworkID:     "8056c2e21c000010",
		Name:          "owned-by-other",
		OwnerID:       "other-user",
		ReasonCode:    services.ImportReasonAlreadyManaged,
		ReasonMessage: "网络已由其他 owner 接管，已跳过",
	})
}

func TestNetworkServiceGetImportableNetworks_EmptySliceWhenControllerHasNoNetworks(t *testing.T) {
	db := newTestSQLiteDB(t)
	client := newTestZTClient(t, map[string]zerotier.Network{})

	service := services.NewNetworkService(client, db)

	result, err := service.GetImportableNetworks()
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Candidates)
	assert.Equal(t, services.ImportableNetworksSummary{}, result.Summary)
}

func TestNetworkServiceGetRuntimeStatus_UsesZeroTierAndDatabaseTruth(t *testing.T) {
	db := newTestSQLiteDB(t)
	createTestUser(t, db, "admin-1", "admin")
	client := newTestZTClientWithStatus(t, map[string]zerotier.Network{}, zerotier.Status{
		Version:     "1.14.2",
		Address:     "f76fd3000b",
		Online:      true,
		TCPFallback: true,
		APIReady:    true,
	})

	service := services.NewNetworkService(client, db)

	status := service.GetRuntimeStatus()
	require.NotNil(t, status)
	assert.Equal(t, "1.14.2", status.Version)
	assert.Equal(t, "f76fd3000b", status.Address)
	assert.True(t, status.Online)
	assert.True(t, status.TCPFallbackAvailable)
	assert.True(t, status.APIReady)
	assert.Equal(t, "online", status.ZeroTierStatus)
	assert.Equal(t, "connected", status.DatabaseStatus)
	assert.Empty(t, status.ZeroTierError)
	assert.Empty(t, status.DatabaseError)
}

func TestNetworkServiceGetRuntimeStatus_ReturnsErrorStates(t *testing.T) {
	db := newTestSQLiteDB(t)
	client := newTestZTClientWithStatusError(t, map[string]zerotier.Network{}, http.StatusInternalServerError)

	service := services.NewNetworkService(client, db)

	status := service.GetRuntimeStatus()
	require.NotNil(t, status)
	assert.Equal(t, "error", status.ZeroTierStatus)
	assert.NotEmpty(t, status.ZeroTierError)
	assert.Equal(t, "connected", status.DatabaseStatus)
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
	require.Len(t, result.Candidates, 2)
	assert.Contains(t, result.Candidates, services.ImportableNetworkCandidate{
		NetworkID:     "8056c2e21c000001",
		Name:          "owned",
		Description:   "owned-desc",
		Status:        services.ImportCandidateManaged,
		CanImport:     false,
		ReasonCode:    services.ImportReasonAlreadyManaged,
		ReasonMessage: "网络已由其他 owner 接管",
		OwnerID:       "user-1",
	})
	assert.Contains(t, result.Candidates, services.ImportableNetworkCandidate{
		NetworkID:     "8056c2e21c000002",
		Name:          "orphan",
		Status:        services.ImportCandidateAvailable,
		CanImport:     true,
		ReasonCode:    services.ImportReasonUnregistered,
		ReasonMessage: "网络尚未登记到 Tairitsu，可直接接管",
	})
	assert.Equal(t, services.ImportableNetworksSummary{Total: 2, Available: 1, Managed: 1, Blocked: 0}, result.Summary)
}

func TestNetworkServiceGetImportableNetworks_IncludesUnassignedAndBlockedCandidates(t *testing.T) {
	db := newTestSQLiteDB(t)
	createTestUser(t, db, "user-1", "user")
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000001",
		Name:        "unassigned-in-db",
		Description: "db-description",
		OwnerID:     "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	client := newTestZTClientWithBrokenNetworkRead(t,
		map[string]zerotier.Network{
			"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "controller-name", Description: "controller-description", Status: "OK"},
		},
		map[string][]zerotier.Member{
			"8056c2e21c000001": {
				{ID: "member-1", Config: zerotier.MemberConfig{Authorized: true}},
				{ID: "member-2", Config: zerotier.MemberConfig{Authorized: false}},
			},
		},
		"8056c2e21c000099",
	)

	service := services.NewNetworkService(client, db)

	result, err := service.GetImportableNetworks()
	require.NoError(t, err)
	require.Len(t, result.Candidates, 2)
	assert.Contains(t, result.Candidates, services.ImportableNetworkCandidate{
		NetworkID:        "8056c2e21c000001",
		Name:             "controller-name",
		Description:      "controller-description",
		ControllerStatus: "OK",
		MemberCount:      intPtr(2),
		Status:           services.ImportCandidateAvailable,
		CanImport:        true,
		ReasonCode:       services.ImportReasonUnassigned,
		ReasonMessage:    "网络已登记但尚未分配 owner，可继续接管",
	})
	assert.Contains(t, result.Candidates, services.ImportableNetworkCandidate{
		NetworkID:     "8056c2e21c000099",
		Status:        services.ImportCandidateBlocked,
		CanImport:     false,
		ReasonCode:    services.ImportReasonControllerReadFailed,
		ReasonMessage: "读取控制器网络详情失败，暂时无法导入",
	})
	assert.Equal(t, services.ImportableNetworksSummary{Total: 2, Available: 1, Managed: 0, Blocked: 1}, result.Summary)
}

func TestNetworkServiceGetNetworkByIDRejectsDifferentOwner(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000001",
		Name:        "alpha",
		Description: "alpha-desc",
		OwnerID:     "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	service := services.NewNetworkService(newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha"},
	}), db)

	detail, err := service.GetNetworkByID("8056c2e21c000001", "owner-2")
	require.ErrorIs(t, err, services.ErrNetworkAccessDenied)
	assert.Nil(t, detail)
}

func TestNetworkServiceGetNetworkMembersRejectsDifferentOwner(t *testing.T) {
	db := newTestSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000001",
		Name:        "alpha",
		Description: "alpha-desc",
		OwnerID:     "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	service := services.NewNetworkService(newTestZTClientWithMembers(t,
		map[string]zerotier.Network{
			"8056c2e21c000001": {ID: "8056c2e21c000001", Name: "alpha"},
		},
		map[string][]zerotier.Member{
			"8056c2e21c000001": {{ID: "member-1"}},
		},
	), db)

	members, err := service.GetNetworkMembers("8056c2e21c000001", "owner-2")
	require.ErrorIs(t, err, services.ErrMemberAccessDenied)
	assert.Nil(t, members)
}

func TestNetworkServiceViewerGrantAndSharedList(t *testing.T) {
	db := newTestSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	createTestUser(t, db, "viewer-2", "user")
	createTestUser(t, db, "admin-1", "admin")

	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000021",
		Name:        "shared-network",
		Description: "shared-desc",
		OwnerID:     "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	service := services.NewNetworkService(nil, db)

	candidates, err := service.GetNetworkViewerCandidates("8056c2e21c000021", "owner-1")
	require.NoError(t, err)
	assert.ElementsMatch(t, []services.NetworkViewerSummary{
		{ID: "viewer-1", Username: "viewer-1", Role: "user"},
		{ID: "viewer-2", Username: "viewer-2", Role: "user"},
	}, candidates)

	require.NoError(t, service.GrantNetworkViewer("8056c2e21c000021", "viewer-1", "owner-1"))

	viewers, err := service.GetNetworkViewers("8056c2e21c000021", "owner-1")
	require.NoError(t, err)
	require.Len(t, viewers, 1)
	assert.Equal(t, "viewer-1", viewers[0].ID)
	assert.Equal(t, "viewer-1", viewers[0].Username)
	assert.Equal(t, "owner-1", viewers[0].GrantedBy)

	sharedNetworks, err := service.GetSharedNetworks("viewer-1")
	require.NoError(t, err)
	require.Len(t, sharedNetworks, 1)
	assert.Equal(t, "8056c2e21c000021", sharedNetworks[0].ID)
	assert.Equal(t, "shared-network", sharedNetworks[0].Name)
	assert.Equal(t, "owner-1", sharedNetworks[0].OwnerID)
	assert.Equal(t, "owner-1", sharedNetworks[0].OwnerUsername)

	remainingCandidates, err := service.GetNetworkViewerCandidates("8056c2e21c000021", "owner-1")
	require.NoError(t, err)
	assert.ElementsMatch(t, []services.NetworkViewerSummary{
		{ID: "viewer-2", Username: "viewer-2", Role: "user"},
	}, remainingCandidates)
}

func TestNetworkServiceViewerGrantRejectsNonOwnerAndInvalidTarget(t *testing.T) {
	db := newTestSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	createTestUser(t, db, "admin-1", "admin")

	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000022",
		Name:        "restricted-network",
		Description: "restricted-desc",
		OwnerID:     "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	service := services.NewNetworkService(nil, db)

	err := service.GrantNetworkViewer("8056c2e21c000022", "viewer-1", "viewer-1")
	require.ErrorIs(t, err, services.ErrViewerAccessDenied)

	err = service.GrantNetworkViewer("8056c2e21c000022", "admin-1", "owner-1")
	require.ErrorIs(t, err, services.ErrViewerTargetInvalid)

	err = service.GrantNetworkViewer("8056c2e21c000022", "owner-1", "owner-1")
	require.ErrorIs(t, err, services.ErrViewerTargetInvalid)
}

func TestNetworkServiceViewerRevokeRemovesSharedAccess(t *testing.T) {
	db := newTestSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")

	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          "8056c2e21c000023",
		Name:        "revoked-network",
		Description: "revoked-desc",
		OwnerID:     "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}))

	service := services.NewNetworkService(nil, db)

	require.NoError(t, service.GrantNetworkViewer("8056c2e21c000023", "viewer-1", "owner-1"))

	sharedBeforeRevoke, err := service.GetSharedNetworks("viewer-1")
	require.NoError(t, err)
	require.Len(t, sharedBeforeRevoke, 1)

	require.NoError(t, service.RevokeNetworkViewer("8056c2e21c000023", "viewer-1", "owner-1"))

	viewers, err := service.GetNetworkViewers("8056c2e21c000023", "owner-1")
	require.NoError(t, err)
	assert.Empty(t, viewers)

	sharedAfterRevoke, err := service.GetSharedNetworks("viewer-1")
	require.NoError(t, err)
	assert.Empty(t, sharedAfterRevoke)
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

func TestZTClientGetStatusSupportsTcpFallbackActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"version":           "1.14.2",
			"address":           "f76fd3000b",
			"online":            true,
			"tcpFallbackActive": true,
			"apiReady":          true,
		}))
	}))
	t.Cleanup(server.Close)

	client := &zerotier.Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: server.Client(),
	}

	status, err := client.GetStatus()
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.TCPFallback)
	assert.True(t, status.APIReady)
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

func newTestZTClientWithStatus(t *testing.T, networks map[string]zerotier.Network, status zerotier.Status) *zerotier.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/status":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"version":              status.Version,
				"address":              status.Address,
				"online":               status.Online,
				"tcpFallbackAvailable": status.TCPFallback,
				"apiReady":             status.APIReady,
			}))
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

func newTestZTClientWithStatusError(t *testing.T, networks map[string]zerotier.Network, statusCode int) *zerotier.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/status":
			http.Error(w, "broken status", statusCode)
		case "/controller/network":
			ids := make([]string, 0, len(networks))
			for id := range networks {
				ids = append(ids, id)
			}
			require.NoError(t, json.NewEncoder(w).Encode(ids))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return &zerotier.Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: server.Client(),
	}
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

func newTestZTClientWithBrokenNetworkRead(t *testing.T, networks map[string]zerotier.Network, members map[string][]zerotier.Member, brokenNetworkID string) *zerotier.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/controller/network":
			ids := make([]string, 0, len(networks)+1)
			for id := range networks {
				ids = append(ids, id)
			}
			ids = append(ids, brokenNetworkID)
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

			if path == brokenNetworkID {
				http.Error(w, "broken", http.StatusInternalServerError)
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

func intPtr(value int) *int {
	return &value
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
