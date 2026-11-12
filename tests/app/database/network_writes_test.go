package database

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateNetworkMetadataPreservesOwnership(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateNetwork(&models.Network{
		ID: "8056c2e21c000031", Name: "before", Description: "before", OwnerID: "owner-1", CreatedAt: now, UpdatedAt: now,
	}))

	require.NoError(t, db.UpdateNetworkMetadata("8056c2e21c000031", "after", "after-desc", now.Add(time.Second)))

	network, err := db.GetNetworkByID("8056c2e21c000031")
	require.NoError(t, err)
	require.NotNil(t, network)
	assert.Equal(t, "after", network.Name)
	assert.Equal(t, "after-desc", network.Description)
	assert.Equal(t, "owner-1", network.OwnerID)
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

func TestConditionalGrantWritesRejectAChangedRole(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
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
	rows, err := db.UpdateNetworkGrantIfRole(grant, "manager")
	require.NoError(t, err)
	assert.Zero(t, rows)

	stored, err := db.GetNetworkGrant(grant.NetworkID, grant.SubjectType, grant.SubjectID)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, "viewer", stored.Role)

	rows, err = db.UpdateNetworkGrantIfRole(grant, "viewer")
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	rows, err = db.DeleteNetworkGrantIfRole(grant.NetworkID, grant.SubjectType, grant.SubjectID, "viewer")
	require.NoError(t, err)
	assert.Zero(t, rows)

	rows, err = db.DeleteNetworkGrantIfRole(grant.NetworkID, grant.SubjectType, grant.SubjectID, "operator")
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	stored, err = db.GetNetworkGrant(grant.NetworkID, grant.SubjectType, grant.SubjectID)
	require.NoError(t, err)
	assert.Nil(t, stored)
}
