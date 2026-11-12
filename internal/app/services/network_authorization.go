package services

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/models"
)

var (
	ErrNetworkNotFound     = errors.New("network not found")
	ErrNetworkAccessDenied = errors.New("network access denied")
	ErrImportAccessDenied  = errors.New("only administrators can import networks")
	ErrImportOwnerRequired = errors.New("network owner is required")
	ErrImportOwnerNotFound = errors.New("specified network owner was not found")
)

func (s *NetworkService) getNetwork(networkID string) (*models.Network, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}

	network, err := db.GetNetworkByID(networkID)
	if err != nil {
		return nil, err
	}
	if network == nil {
		return nil, ErrNetworkNotFound
	}

	return network, nil
}

// getNetworkRole resolves the caller's effective role on an already-loaded
// network: ownership first, then the account's delegated grant.
func (s *NetworkService) getNetworkRole(network *models.Network, userID string) (NetworkRole, error) {
	if network.OwnerID == userID {
		return NetworkRoleOwner, nil
	}

	db := s.getDB()
	if db == nil {
		return "", errors.New("database is not initialized")
	}

	grant, err := db.GetNetworkGrant(network.ID, models.NetworkGrantSubjectUser, userID)
	if err != nil {
		return "", err
	}
	if grant == nil {
		return "", ErrNetworkAccessDenied
	}

	role, ok := ParseNetworkRole(grant.Role)
	if !ok {
		return "", ErrNetworkAccessDenied
	}

	return role, nil
}

// EffectiveNetworkRole returns the role the user holds on the network, or
// ErrNetworkAccessDenied when they hold none.
func (s *NetworkService) EffectiveNetworkRole(networkID, userID string) (NetworkRole, error) {
	network, err := s.getNetwork(networkID)
	if err != nil {
		return "", err
	}
	return s.getNetworkRole(network, userID)
}

// authorizeNetwork is the one place handlers and services ask whether a caller
// may act. It loads the network, resolves the role, and checks the requested
// permission, so a missing grant and an insufficient role both fail closed.
func (s *NetworkService) authorizeNetwork(networkID, userID string, permission NetworkPermission) (*models.Network, NetworkRole, error) {
	network, err := s.getNetwork(networkID)
	if err != nil {
		return nil, "", err
	}

	role, err := s.getNetworkRole(network, userID)
	if err != nil {
		return nil, "", err
	}

	if !role.Has(permission) {
		return nil, role, ErrNetworkPermissionDenied
	}

	return network, role, nil
}

func authorizeImport(actorRole, ownerID string) error {
	if actorRole != "admin" {
		return ErrImportAccessDenied
	}
	if ownerID == "" {
		return ErrImportOwnerRequired
	}
	return nil
}

func IsNetworkNotFound(err error) bool {
	return errors.Is(err, ErrNetworkNotFound)
}

func IsNetworkAccessDenied(err error) bool {
	return errors.Is(err, ErrNetworkAccessDenied) ||
		errors.Is(err, ErrNetworkPermissionDenied)
}
