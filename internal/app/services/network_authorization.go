package services

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/models"
)

var (
	ErrNetworkNotFound     = errors.New("网络不存在")
	ErrNetworkAccessDenied = errors.New("无权限访问网络")
	ErrMemberAccessDenied  = errors.New("无权限访问网络成员")
	ErrViewerAccessDenied  = errors.New("无权限管理网络查看授权")
	ErrViewerTargetInvalid = errors.New("只能授权普通用户查看网络")
	ErrImportAccessDenied  = errors.New("只有管理员可以导入网络")
	ErrImportOwnerRequired = errors.New("必须指定网络所有者")
	ErrImportOwnerNotFound = errors.New("指定的网络所有者不存在")
)

func (s *NetworkService) getNetwork(networkID string) (*models.Network, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
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
		return nil, errors.New("数据库未初始化")
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
