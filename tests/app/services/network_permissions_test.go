package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkRolePermissionMatrix(t *testing.T) {
	tests := []struct {
		role        services.NetworkRole
		permissions map[services.NetworkPermission]bool
	}{
		{
			role: services.NetworkRoleViewer,
			permissions: map[services.NetworkPermission]bool{
				services.NetworkPermissionRead:          true,
				services.NetworkPermissionMembersRead:   true,
				services.NetworkPermissionAccessRead:    true,
				services.NetworkPermissionMembersWrite:  false,
				services.NetworkPermissionSettingsWrite: false,
				services.NetworkPermissionAccessWrite:   false,
				services.NetworkPermissionNetworkDelete: false,
			},
		},
		{
			role: services.NetworkRoleOperator,
			permissions: map[services.NetworkPermission]bool{
				services.NetworkPermissionRead:          true,
				services.NetworkPermissionMembersRead:   true,
				services.NetworkPermissionMembersWrite:  true,
				services.NetworkPermissionAccessRead:    true,
				services.NetworkPermissionMembersBridge: false,
				services.NetworkPermissionSettingsWrite: false,
				services.NetworkPermissionAccessWrite:   false,
			},
		},
		{
			role: services.NetworkRoleManager,
			permissions: map[services.NetworkPermission]bool{
				services.NetworkPermissionRead:            true,
				services.NetworkPermissionMembersWrite:    true,
				services.NetworkPermissionMembersBridge:   true,
				services.NetworkPermissionSettingsWrite:   true,
				services.NetworkPermissionAccessRead:      true,
				services.NetworkPermissionAccessWrite:     true,
				services.NetworkPermissionNetworkDelete:   false,
				services.NetworkPermissionNetworkTransfer: false,
			},
		},
		{
			role: services.NetworkRoleOwner,
			permissions: map[services.NetworkPermission]bool{
				services.NetworkPermissionRead:            true,
				services.NetworkPermissionMembersWrite:    true,
				services.NetworkPermissionMembersBridge:   true,
				services.NetworkPermissionSettingsWrite:   true,
				services.NetworkPermissionAccessWrite:     true,
				services.NetworkPermissionNetworkDelete:   true,
				services.NetworkPermissionNetworkTransfer: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.role), func(t *testing.T) {
			for permission, want := range test.permissions {
				assert.Equal(t, want, test.role.Has(permission), "permission %s", permission)
			}
		})
	}
}

func TestNetworkRoleAssignmentAndManagementRules(t *testing.T) {
	assert.True(t, services.NetworkRoleOwner.CanAssign(services.NetworkRoleManager))
	assert.True(t, services.NetworkRoleOwner.CanAssign(services.NetworkRoleViewer))
	assert.False(t, services.NetworkRoleOwner.CanAssign(services.NetworkRoleOwner))

	assert.True(t, services.NetworkRoleManager.CanAssign(services.NetworkRoleOperator))
	assert.True(t, services.NetworkRoleManager.CanAssign(services.NetworkRoleViewer))
	assert.False(t, services.NetworkRoleManager.CanAssign(services.NetworkRoleManager))
	assert.False(t, services.NetworkRoleManager.CanAssign(services.NetworkRoleOwner))

	assert.False(t, services.NetworkRoleOperator.CanAssign(services.NetworkRoleViewer))
	assert.False(t, services.NetworkRoleViewer.CanAssign(services.NetworkRoleViewer))

	assert.True(t, services.NetworkRoleOwner.CanManage(services.NetworkRoleManager))
	assert.True(t, services.NetworkRoleManager.CanManage(services.NetworkRoleOperator))
	assert.True(t, services.NetworkRoleManager.CanManage(services.NetworkRoleViewer))
	assert.False(t, services.NetworkRoleManager.CanManage(services.NetworkRoleManager))
	assert.False(t, services.NetworkRoleManager.CanManage(services.NetworkRoleOwner))
	assert.False(t, services.NetworkRoleOperator.CanManage(services.NetworkRoleViewer))
}

func TestParseNetworkRole(t *testing.T) {
	for _, value := range []string{"viewer", "operator", "manager", "owner"} {
		role, ok := services.ParseNetworkRole(value)
		require.True(t, ok)
		assert.Equal(t, services.NetworkRole(value), role)
	}

	_, ok := services.ParseNetworkRole("superuser")
	assert.False(t, ok)

	_, ok = services.ParseNetworkRole("")
	assert.False(t, ok)
}
