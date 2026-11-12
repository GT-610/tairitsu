package services

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedAccessNetwork(t *testing.T, db database.DBInterface, networkID, ownerID string) {
	t.Helper()

	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID:          networkID,
		Name:        "access-network",
		Description: "access test",
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}))
}

func TestNetworkServiceAccessGrantAndSharedList(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	createTestUser(t, db, "operator-1", "user")
	createTestUser(t, db, "manager-1", "user")
	createTestUser(t, db, "outside-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000021", "owner-1")

	service := services.NewNetworkService(nil, db)

	candidates, err := service.GetNetworkAccessCandidates("8056c2e21c000021", "owner-1", "viewer")
	require.NoError(t, err)
	assert.Equal(t, []services.NetworkAccessCandidate{{ID: "viewer-1", Username: "viewer-1"}}, candidates)

	emptyCandidates, err := service.GetNetworkAccessCandidates("8056c2e21c000021", "owner-1", " ")
	require.NoError(t, err)
	assert.Empty(t, emptyCandidates)

	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000021", "owner-1", "viewer-1", "viewer"))
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000021", "owner-1", "operator-1", "operator"))
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000021", "owner-1", "manager-1", "manager"))

	access, err := service.GetNetworkAccess("8056c2e21c000021", "owner-1")
	require.NoError(t, err)
	assert.Equal(t, "owner-1", access.OwnerID)
	assert.Equal(t, "owner-1", access.OwnerUsername)
	assert.Equal(t, services.NetworkRoleOwner, access.MyRole)
	require.Len(t, access.Entries, 3)
	assert.Equal(t, "manager-1", access.Entries[0].Username)
	assert.Equal(t, services.NetworkRoleManager, access.Entries[0].Role)
	assert.Equal(t, "operator-1", access.Entries[1].Username)
	assert.Equal(t, "viewer-1", access.Entries[2].Username)
	assert.Equal(t, "owner-1", access.Entries[0].GrantedByUsername)

	for userID, wantRole := range map[string]services.NetworkRole{
		"viewer-1":   services.NetworkRoleViewer,
		"operator-1": services.NetworkRoleOperator,
		"manager-1":  services.NetworkRoleManager,
	} {
		shared, err := service.GetSharedNetworks(userID)
		require.NoError(t, err)
		require.Len(t, shared, 1)
		assert.Equal(t, "8056c2e21c000021", shared[0].ID)
		assert.Equal(t, wantRole, shared[0].Role)
		assert.Equal(t, "owner-1", shared[0].OwnerUsername)
	}

	ownerShared, err := service.GetSharedNetworks("owner-1")
	require.NoError(t, err)
	assert.Empty(t, ownerShared)

	remaining, err := service.GetNetworkAccessCandidates("8056c2e21c000021", "owner-1", "viewer")
	require.NoError(t, err)
	assert.Empty(t, remaining)

	viewerAccess, err := service.GetNetworkAccess("8056c2e21c000021", "viewer-1")
	require.NoError(t, err)
	assert.Equal(t, services.NetworkRoleViewer, viewerAccess.MyRole)

	_, err = service.GetNetworkAccessCandidates("8056c2e21c000021", "viewer-1", "outside")
	require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)
}

func TestNetworkServiceAccessRoleConstraints(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "manager-1", "user")
	createTestUser(t, db, "manager-2", "user")
	createTestUser(t, db, "operator-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000022", "owner-1")

	service := services.NewNetworkService(nil, db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000022", "owner-1", "manager-1", "manager"))
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000022", "owner-1", "manager-2", "manager"))
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000022", "owner-1", "viewer-1", "viewer"))

	t.Run("manager may add an operator", func(t *testing.T) {
		require.NoError(t, service.GrantNetworkAccess("8056c2e21c000022", "manager-1", "operator-1", "operator"))
	})

	t.Run("manager may not appoint a manager", func(t *testing.T) {
		err := service.UpdateNetworkAccess("8056c2e21c000022", "manager-1", "operator-1", "manager")
		require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)
	})

	t.Run("manager may not touch another manager", func(t *testing.T) {
		err := service.UpdateNetworkAccess("8056c2e21c000022", "manager-1", "manager-2", "viewer")
		require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)

		err = service.RevokeNetworkAccess("8056c2e21c000022", "manager-1", "manager-2")
		require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)
	})

	t.Run("manager may demote a viewer", func(t *testing.T) {
		require.NoError(t, service.UpdateNetworkAccess("8056c2e21c000022", "manager-1", "viewer-1", "operator"))
	})

	t.Run("owner may appoint a manager", func(t *testing.T) {
		require.NoError(t, service.UpdateNetworkAccess("8056c2e21c000022", "owner-1", "operator-1", "manager"))
	})

	t.Run("unknown and owner roles are rejected", func(t *testing.T) {
		err := service.GrantNetworkAccess("8056c2e21c000022", "owner-1", "viewer-1", "superuser")
		require.ErrorIs(t, err, services.ErrNetworkGrantRoleInvalid)

		err = service.GrantNetworkAccess("8056c2e21c000022", "owner-1", "viewer-1", "owner")
		require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)
	})

	t.Run("duplicate and self changes are rejected", func(t *testing.T) {
		err := service.GrantNetworkAccess("8056c2e21c000022", "owner-1", "manager-1", "manager")
		require.ErrorIs(t, err, services.ErrNetworkGrantExists)

		err = service.UpdateNetworkAccess("8056c2e21c000022", "manager-1", "manager-1", "viewer")
		require.ErrorIs(t, err, services.ErrNetworkGrantSelfChange)
	})

	t.Run("owner cannot be revoked or leave", func(t *testing.T) {
		err := service.RevokeNetworkAccess("8056c2e21c000022", "owner-1", "owner-1")
		require.ErrorIs(t, err, services.ErrNetworkGrantSelfChange)

		err = service.LeaveNetwork("8056c2e21c000022", "owner-1")
		require.ErrorIs(t, err, services.ErrNetworkOwnerCannotLeave)
	})

	t.Run("non-owner can leave", func(t *testing.T) {
		require.NoError(t, service.LeaveNetwork("8056c2e21c000022", "manager-2"))

		grant, err := db.GetNetworkGrant("8056c2e21c000022", models.NetworkGrantSubjectUser, "manager-2")
		require.NoError(t, err)
		assert.Nil(t, grant)
	})
}

func TestNetworkServiceAccessRevokeAndLeave(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000023", "owner-1")

	service := services.NewNetworkService(nil, db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000023", "owner-1", "viewer-1", "viewer"))

	sharedBefore, err := service.GetSharedNetworks("viewer-1")
	require.NoError(t, err)
	require.Len(t, sharedBefore, 1)

	require.NoError(t, service.RevokeNetworkAccess("8056c2e21c000023", "owner-1", "viewer-1"))

	access, err := service.GetNetworkAccess("8056c2e21c000023", "owner-1")
	require.NoError(t, err)
	assert.Empty(t, access.Entries)

	sharedAfter, err := service.GetSharedNetworks("viewer-1")
	require.NoError(t, err)
	assert.Empty(t, sharedAfter)
}

func TestNetworkServiceDeleteNetworkRemovesGrants(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000024", "owner-1")

	service := services.NewNetworkService(newTestZTClient(t, map[string]zerotier.Network{
		"8056c2e21c000024": {ID: "8056c2e21c000024", Name: "delete-target"},
	}), db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000024", "owner-1", "viewer-1", "viewer"))

	require.NoError(t, service.DeleteNetwork("8056c2e21c000024", "owner-1"))

	network, err := db.GetNetworkByID("8056c2e21c000024")
	require.NoError(t, err)
	assert.Nil(t, network)

	grant, err := db.GetNetworkGrant("8056c2e21c000024", models.NetworkGrantSubjectUser, "viewer-1")
	require.NoError(t, err)
	assert.Nil(t, grant)
}

func TestNetworkServiceTransferNetworkOwnership(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "target-1", "user")
	createTestUser(t, db, "outside-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000025", "owner-1")

	service := services.NewNetworkService(nil, db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000025", "owner-1", "target-1", "manager"))

	require.NoError(t, service.TransferNetworkOwnership("8056c2e21c000025", "owner-1", "target-1", "manager"))

	network, err := db.GetNetworkByID("8056c2e21c000025")
	require.NoError(t, err)
	require.NotNil(t, network)
	assert.Equal(t, "target-1", network.OwnerID)

	targetGrant, err := db.GetNetworkGrant("8056c2e21c000025", models.NetworkGrantSubjectUser, "target-1")
	require.NoError(t, err)
	assert.Nil(t, targetGrant)

	previousOwnerGrant, err := db.GetNetworkGrant("8056c2e21c000025", models.NetworkGrantSubjectUser, "owner-1")
	require.NoError(t, err)
	require.NotNil(t, previousOwnerGrant)
	assert.Equal(t, "manager", previousOwnerGrant.Role)

	shared, err := service.GetSharedNetworks("owner-1")
	require.NoError(t, err)
	require.Len(t, shared, 1)
	assert.Equal(t, services.NetworkRoleManager, shared[0].Role)

	t.Run("target must already be a member", func(t *testing.T) {
		err := service.TransferNetworkOwnership("8056c2e21c000025", "target-1", "outside-1", "manager")
		require.ErrorIs(t, err, services.ErrNetworkOwnerTransferInvalid)
	})

	t.Run("only the owner can transfer", func(t *testing.T) {
		err := service.TransferNetworkOwnership("8056c2e21c000025", "owner-1", "target-1", "manager")
		require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)
	})

	t.Run("new owner cannot leave", func(t *testing.T) {
		err := service.LeaveNetwork("8056c2e21c000025", "target-1")
		require.ErrorIs(t, err, services.ErrNetworkOwnerCannotLeave)
	})
}

func TestNetworkServiceAccessControlsMemberAndSettingsWrites(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	createTestUser(t, db, "operator-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000026", "owner-1")

	service := services.NewNetworkService(newTestZTClientWithMembers(t,
		map[string]zerotier.Network{
			"8056c2e21c000026": {ID: "8056c2e21c000026", Name: "guarded"},
		},
		map[string][]zerotier.Member{
			"8056c2e21c000026": {{ID: "member-1", Config: zerotier.MemberConfig{Authorized: true}}},
		},
	), db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000026", "owner-1", "viewer-1", "viewer"))
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000026", "owner-1", "operator-1", "operator"))

	members, err := service.GetNetworkMembers("8056c2e21c000026", "viewer-1")
	require.NoError(t, err)
	require.Len(t, members, 1)

	_, err = service.UpdateNetwork("8056c2e21c000026", &zerotier.NetworkUpdateRequest{Name: "nope"}, "viewer-1")
	require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)

	_, err = service.UpdateNetworkMember("8056c2e21c000026", "member-1", &zerotier.MemberUpdateRequest{Name: "nope"}, "viewer-1")
	require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)

	err = service.RemoveNetworkMember("8056c2e21c000026", "member-1", "viewer-1")
	require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)

	_, err = service.UpdateNetwork("8056c2e21c000026", &zerotier.NetworkUpdateRequest{Name: "nope"}, "operator-1")
	require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)

	enabled := true
	_, err = service.UpdateNetworkMember("8056c2e21c000026", "member-1", &zerotier.MemberUpdateRequest{ActiveBridge: &enabled}, "operator-1")
	require.ErrorIs(t, err, services.ErrNetworkPermissionDenied)

	err = service.RemoveNetworkMember("8056c2e21c000026", "member-1", "outside-1")
	require.ErrorIs(t, err, services.ErrNetworkAccessDenied)
}

func TestNetworkServiceGetNetworkByIDReturnsAccessInfo(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "viewer-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000027", "owner-1")

	service := services.NewNetworkService(newTestZTClientWithMembers(t,
		map[string]zerotier.Network{
			"8056c2e21c000027": {ID: "8056c2e21c000027", Name: "readable"},
		},
		map[string][]zerotier.Member{
			"8056c2e21c000027": {{ID: "member-1", Config: zerotier.MemberConfig{Authorized: true}}},
		},
	), db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000027", "owner-1", "viewer-1", "viewer"))

	ownerDetail, err := service.GetNetworkByID("8056c2e21c000027", "owner-1")
	require.NoError(t, err)
	assert.Equal(t, services.NetworkRoleOwner, ownerDetail.Access.Role)
	assert.Contains(t, ownerDetail.Access.Permissions, services.NetworkPermissionNetworkTransfer)

	viewerDetail, err := service.GetNetworkByID("8056c2e21c000027", "viewer-1")
	require.NoError(t, err)
	assert.Equal(t, services.NetworkRoleViewer, viewerDetail.Access.Role)
	assert.Contains(t, viewerDetail.Access.Permissions, services.NetworkPermissionMembersRead)
	assert.NotContains(t, viewerDetail.Access.Permissions, services.NetworkPermissionMembersWrite)
	assert.NotContains(t, viewerDetail.Access.Permissions, services.NetworkPermissionSettingsWrite)

	_, err = service.GetNetworkByID("8056c2e21c000027", "outside-1")
	require.ErrorIs(t, err, services.ErrNetworkAccessDenied)
}

func TestNetworkServiceTransferCanRemoveThePreviousOwner(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	createTestUser(t, db, "owner-1", "user")
	createTestUser(t, db, "target-1", "user")
	seedAccessNetwork(t, db, "8056c2e21c000028", "owner-1")

	service := services.NewNetworkService(nil, db)
	require.NoError(t, service.GrantNetworkAccess("8056c2e21c000028", "owner-1", "target-1", "operator"))

	require.NoError(t, service.TransferNetworkOwnership("8056c2e21c000028", "owner-1", "target-1", "none"))

	network, err := db.GetNetworkByID("8056c2e21c000028")
	require.NoError(t, err)
	require.NotNil(t, network)
	assert.Equal(t, "target-1", network.OwnerID)

	previousOwnerGrant, err := db.GetNetworkGrant("8056c2e21c000028", models.NetworkGrantSubjectUser, "owner-1")
	require.NoError(t, err)
	assert.Nil(t, previousOwnerGrant)

	targetGrant, err := db.GetNetworkGrant("8056c2e21c000028", models.NetworkGrantSubjectUser, "target-1")
	require.NoError(t, err)
	assert.Nil(t, targetGrant)
}
