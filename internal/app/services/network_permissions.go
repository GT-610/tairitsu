package services

import "errors"

// NetworkRole is the role an account holds on one network. Ownership is stored
// on the network record itself; the other three roles are delegated through
// network_grants.
type NetworkRole string

const (
	NetworkRoleViewer   NetworkRole = "viewer"
	NetworkRoleOperator NetworkRole = "operator"
	NetworkRoleManager  NetworkRole = "manager"
	NetworkRoleOwner    NetworkRole = "owner"
)

// NetworkPermission is a single capability the backend checks before acting.
// The frontend never compares role names: it reads the permission list from the
// network detail response, so this table remains the single source of truth.
type NetworkPermission string

const (
	NetworkPermissionRead            NetworkPermission = "network.read"
	NetworkPermissionMembersRead     NetworkPermission = "members.read"
	NetworkPermissionMembersWrite    NetworkPermission = "members.write"
	NetworkPermissionMembersBridge   NetworkPermission = "members.bridge"
	NetworkPermissionSettingsWrite   NetworkPermission = "settings.write"
	NetworkPermissionAccessRead      NetworkPermission = "access.read"
	NetworkPermissionAccessWrite     NetworkPermission = "access.write"
	NetworkPermissionNetworkDelete   NetworkPermission = "network.delete"
	NetworkPermissionNetworkTransfer NetworkPermission = "network.transfer"
)

// allNetworkPermissions fixes the order of the permission list returned to the
// frontend so responses are stable and easy to read in tests.
var allNetworkPermissions = []NetworkPermission{
	NetworkPermissionRead,
	NetworkPermissionMembersRead,
	NetworkPermissionMembersWrite,
	NetworkPermissionMembersBridge,
	NetworkPermissionSettingsWrite,
	NetworkPermissionAccessRead,
	NetworkPermissionAccessWrite,
	NetworkPermissionNetworkDelete,
	NetworkPermissionNetworkTransfer,
}

var networkRolePermissions = map[NetworkRole][]NetworkPermission{
	NetworkRoleViewer: {
		NetworkPermissionRead,
		NetworkPermissionMembersRead,
		NetworkPermissionAccessRead,
	},
	NetworkRoleOperator: {
		NetworkPermissionRead,
		NetworkPermissionMembersRead,
		NetworkPermissionMembersWrite,
		NetworkPermissionAccessRead,
	},
	NetworkRoleManager: {
		NetworkPermissionRead,
		NetworkPermissionMembersRead,
		NetworkPermissionMembersWrite,
		NetworkPermissionMembersBridge,
		NetworkPermissionSettingsWrite,
		NetworkPermissionAccessRead,
		NetworkPermissionAccessWrite,
	},
	NetworkRoleOwner: allNetworkPermissions,
}

var networkRoleRank = map[NetworkRole]int{
	NetworkRoleViewer:   1,
	NetworkRoleOperator: 2,
	NetworkRoleManager:  3,
	NetworkRoleOwner:    4,
}

var (
	ErrNetworkPermissionDenied     = errors.New("network permission denied")
	ErrNetworkGrantNotFound        = errors.New("network grant not found")
	ErrNetworkGrantExists          = errors.New("network grant already exists")
	ErrNetworkGrantChanged         = errors.New("network grant changed while the request was in flight")
	ErrNetworkGrantRoleInvalid     = errors.New("invalid network role")
	ErrNetworkGrantTargetInvalid   = errors.New("network grant target is not allowed")
	ErrNetworkGrantSelfChange      = errors.New("cannot change your own network role")
	ErrNetworkOwnerTransferInvalid = errors.New("network ownership transfer is not allowed")
	ErrNetworkOwnerCannotLeave     = errors.New("the network owner cannot leave; transfer ownership first")
)

// ParseNetworkRole reports whether value is one of the four known roles.
func ParseNetworkRole(value string) (NetworkRole, bool) {
	role := NetworkRole(value)
	_, ok := networkRoleRank[role]
	return role, ok
}

// Rank orders roles for comparisons; an unknown role ranks below every role.
func (r NetworkRole) Rank() int {
	return networkRoleRank[r]
}

// Has reports whether the role grants permission.
func (r NetworkRole) Has(permission NetworkPermission) bool {
	for _, granted := range networkRolePermissions[r] {
		if granted == permission {
			return true
		}
	}
	return false
}

// Permissions returns the role's permissions in the canonical order.
func (r NetworkRole) Permissions() []NetworkPermission {
	result := make([]NetworkPermission, 0, len(networkRolePermissions[r]))
	for _, permission := range allNetworkPermissions {
		if r.Has(permission) {
			result = append(result, permission)
		}
	}
	return result
}

// CanAssign reports whether an actor with this role may grant target. Ownership
// is never assignable through the grants API; it only moves through a transfer.
func (r NetworkRole) CanAssign(target NetworkRole) bool {
	if target == NetworkRoleOwner {
		return false
	}
	switch r {
	case NetworkRoleOwner:
		return target == NetworkRoleViewer || target == NetworkRoleOperator || target == NetworkRoleManager
	case NetworkRoleManager:
		return target == NetworkRoleViewer || target == NetworkRoleOperator
	default:
		return false
	}
}

// CanManage reports whether an actor with this role may change or remove a
// grant that currently holds target. Managers deliberately cannot touch other
// managers, which keeps two delegated administrators from demoting each other.
func (r NetworkRole) CanManage(target NetworkRole) bool {
	if target == NetworkRoleOwner {
		return false
	}
	return r.Rank() > target.Rank() && r.Rank() >= NetworkRoleManager.Rank()
}
