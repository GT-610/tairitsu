package services

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/models"
)

var (
	ErrNetworkNotFound     = errors.New("network not found")
	ErrNetworkAccessDenied = errors.New("network access denied")
	ErrMemberAccessDenied  = errors.New("network member access denied")
	ErrViewerAccessDenied  = errors.New("network viewer access denied")
	ErrViewerTargetInvalid = errors.New("only regular users can be granted network viewer access")
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

func (s *NetworkService) getOwnedNetwork(networkID, userID string) (*models.Network, error) {
	network, err := s.getNetwork(networkID)
	if err != nil {
		return nil, err
	}
	if network.OwnerID != userID {
		return nil, ErrNetworkAccessDenied
	}

	return network, nil
}

func (s *NetworkService) authorizeOwnedNetwork(networkID, userID string) (*models.Network, error) {
	return s.getOwnedNetwork(networkID, userID)
}

func (s *NetworkService) authorizeMemberReadAccess(networkID, userID string) (*models.Network, error) {
	network, err := s.getNetwork(networkID)
	if err != nil {
		return nil, err
	}
	if network.OwnerID == userID {
		return network, nil
	}

	db := s.getDB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	viewer, viewerErr := db.GetNetworkViewer(networkID, userID)
	if viewerErr != nil {
		return nil, viewerErr
	}
	if viewer == nil {
		return nil, ErrMemberAccessDenied
	}

	return network, nil
}

func (s *NetworkService) authorizeMemberWriteAccess(networkID, userID string) (*models.Network, error) {
	network, err := s.getOwnedNetwork(networkID, userID)
	if err != nil {
		if IsNetworkAccessDenied(err) {
			return nil, ErrMemberAccessDenied
		}
		return nil, err
	}

	return network, nil
}

func (s *NetworkService) authorizeViewerManagement(networkID, userID string) (*models.Network, error) {
	network, err := s.getOwnedNetwork(networkID, userID)
	if err != nil {
		if IsNetworkAccessDenied(err) {
			return nil, ErrViewerAccessDenied
		}
		return nil, err
	}
	return network, nil
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
	return errors.Is(err, ErrNetworkAccessDenied) || errors.Is(err, ErrMemberAccessDenied) || errors.Is(err, ErrViewerAccessDenied)
}
