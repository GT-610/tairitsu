package database

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateNetworkMetadataWritesOnlyProvidedFields(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: "8056c2e21c000031", Name: "before", Description: "before", OwnerID: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))

	name, description := "after", "after-desc"
	require.NoError(t, db.UpdateNetworkMetadata("8056c2e21c000031", &name, &description, now.Add(time.Second)))

	network, err := db.GetNetworkByID("8056c2e21c000031")
	require.NoError(t, err)
	require.NotNil(t, network)
	assert.Equal(t, "after", network.Name)
	assert.Equal(t, "after-desc", network.Description)
	assert.Equal(t, "owner-1", network.OwnerID)

	// A description-only update must not restore the old name.
	name = "renamed"
	require.NoError(t, db.UpdateNetworkMetadata("8056c2e21c000031", &name, nil, now.Add(2*time.Second)))
	network, err = db.GetNetworkByID("8056c2e21c000031")
	require.NoError(t, err)
	assert.Equal(t, "renamed", network.Name)
	assert.Equal(t, "after-desc", network.Description)
}

func TestUpdateNetworkOwnerIfCurrentRejectsAStaleOwner(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: "8056c2e21c000032", Name: "network", OwnerID: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))

	rows, err := db.UpdateNetworkOwnerIfCurrent("8056c2e21c000032", "someone-else", "owner-2", now)
	require.NoError(t, err)
	assert.Zero(t, rows)

	rows, err = db.UpdateNetworkOwnerIfCurrent("8056c2e21c000032", "owner-1", "owner-2", now)
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	network, err := db.GetNetworkByID("8056c2e21c000032")
	require.NoError(t, err)
	assert.Equal(t, "owner-2", network.OwnerID)
}

func TestConditionalGrantWritesRejectAChangedRoleOrRevokedActor(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: "8056c2e21c000033", Name: "network", OwnerID: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))
	grant := &models.NetworkGrant{
		NetworkID:   "8056c2e21c000033",
		SubjectType: models.NetworkGrantSubjectUser,
		SubjectID:   "user-1",
		Role:        "viewer",
		GrantedBy:   "owner-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, db.CreateNetworkGrant(grant))

	grant.Role = "operator"
	rows, err := db.UpdateNetworkGrantIfActorCanWriteAndRole(grant, "owner-1", "manager")
	require.NoError(t, err)
	assert.Zero(t, rows)

	stored, err := db.GetNetworkGrant(grant.NetworkID, grant.SubjectType, grant.SubjectID)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, "viewer", stored.Role)

	// A caller without write access cannot mutate the grant even with the
	// expected role.
	rows, err = db.UpdateNetworkGrantIfActorCanWriteAndRole(grant, "intruder", "viewer")
	require.NoError(t, err)
	assert.Zero(t, rows)

	rows, err = db.UpdateNetworkGrantIfActorCanWriteAndRole(grant, "owner-1", "viewer")
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	rows, err = db.DeleteNetworkGrantIfActorCanWriteAndRole(grant.NetworkID, grant.SubjectType, grant.SubjectID, "owner-1", "viewer")
	require.NoError(t, err)
	assert.Zero(t, rows)

	rows, err = db.DeleteNetworkGrantIfActorCanWriteAndRole(grant.NetworkID, grant.SubjectType, grant.SubjectID, "owner-1", "operator")
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	stored, err = db.GetNetworkGrant(grant.NetworkID, grant.SubjectType, grant.SubjectID)
	require.NoError(t, err)
	assert.Nil(t, stored)
}

func TestConditionalGrantWritesFollowManagerAccess(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: "8056c2e21c000034", Name: "network", OwnerID: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, db.CreateNetworkGrant(&models.NetworkGrant{
		NetworkID: "8056c2e21c000034", SubjectType: models.NetworkGrantSubjectUser, SubjectID: "manager-1",
		Role: "manager", GrantedBy: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))
	target := &models.NetworkGrant{
		NetworkID: "8056c2e21c000034", SubjectType: models.NetworkGrantSubjectUser, SubjectID: "user-1",
		Role: "viewer", GrantedBy: "owner-1", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.CreateNetworkGrant(target))

	target.Role = "operator"
	rows, err := db.UpdateNetworkGrantIfActorCanWriteAndRole(target, "manager-1", "viewer")
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	require.NoError(t, db.CreateNetworkGrant(&models.NetworkGrant{
		NetworkID: "8056c2e21c000034", SubjectType: models.NetworkGrantSubjectUser, SubjectID: "manager-2",
		Role: "manager", GrantedBy: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))
	managerTarget := &models.NetworkGrant{
		NetworkID: "8056c2e21c000034", SubjectType: models.NetworkGrantSubjectUser, SubjectID: "manager-2",
		Role: "manager", GrantedBy: "owner-1", CreatedAt: now, UpdatedAt: now,
	}
	managerTarget.Role = "viewer"
	rows, err = db.UpdateNetworkGrantIfActorCanWriteAndRole(managerTarget, "manager-1", "manager")
	require.NoError(t, err)
	assert.Zero(t, rows, "a manager must not demote another manager")

	rows, err = db.CreateNetworkGrantIfActorCanWrite(&models.NetworkGrant{
		NetworkID: "8056c2e21c000034", SubjectType: models.NetworkGrantSubjectUser, SubjectID: "user-3",
		Role: "manager", GrantedBy: "manager-1", CreatedAt: now, UpdatedAt: now,
	}, "manager-1")
	require.NoError(t, err)
	assert.Zero(t, rows, "a manager must not appoint a manager")

	require.NoError(t, db.DeleteNetworkGrant("8056c2e21c000034", models.NetworkGrantSubjectUser, "manager-1"))

	target.Role = "viewer"
	rows, err = db.UpdateNetworkGrantIfActorCanWriteAndRole(target, "manager-1", "operator")
	require.NoError(t, err)
	assert.Zero(t, rows)
}

func TestConditionalGrantCreateFollowsActorAccess(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: "8056c2e21c000035", Name: "network", OwnerID: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, db.CreateNetworkGrant(&models.NetworkGrant{
		NetworkID: "8056c2e21c000035", SubjectType: models.NetworkGrantSubjectUser, SubjectID: "manager-1",
		Role: "manager", GrantedBy: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))

	newGrant := func(userID string) *models.NetworkGrant {
		return &models.NetworkGrant{
			NetworkID: "8056c2e21c000035", SubjectType: models.NetworkGrantSubjectUser, SubjectID: userID,
			Role: "viewer", GrantedBy: "manager-1", CreatedAt: now, UpdatedAt: now,
		}
	}

	rows, err := db.CreateNetworkGrantIfActorCanWrite(newGrant("user-1"), "manager-1")
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	rows, err = db.CreateNetworkGrantIfActorCanWrite(newGrant("user-1"), "owner-1")
	require.NoError(t, err)
	assert.Zero(t, rows, "an existing grant must not be duplicated")

	require.NoError(t, db.DeleteNetworkGrant("8056c2e21c000035", models.NetworkGrantSubjectUser, "manager-1"))

	rows, err = db.CreateNetworkGrantIfActorCanWrite(newGrant("user-2"), "manager-1")
	require.NoError(t, err)
	assert.Zero(t, rows, "a revoked manager must not create grants")
}
