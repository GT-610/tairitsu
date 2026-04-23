package services

import (
	"errors"

	"github.com/GT-610/tairitsu/internal/app/models"
)

var (
	ErrNetworkNotFound     = errors.New("网络不存在")
	ErrNetworkAccessDenied = errors.New("无权限访问网络")
	ErrMemberAccessDenied  = errors.New("无权限访问网络成员")
	ErrImportAccessDenied  = errors.New("只有管理员可以导入网络")
	ErrImportOwnerRequired = errors.New("必须指定网络所有者")
	ErrImportOwnerNotFound = errors.New("指定的网络所有者不存在")
)

func (s *NetworkService) getOwnedNetwork(networkID, userID string) (*models.Network, error) {
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
	if network.OwnerID != userID {
		return nil, ErrNetworkAccessDenied
	}

	return network, nil
}

func (s *NetworkService) authorizeOwnedNetwork(networkID, userID string) (*models.Network, error) {
	return s.getOwnedNetwork(networkID, userID)
}

func (s *NetworkService) authorizeMemberAccess(networkID, userID string) (*models.Network, error) {
	network, err := s.getOwnedNetwork(networkID, userID)
	if err != nil {
		if IsNetworkAccessDenied(err) {
			return nil, ErrMemberAccessDenied
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
	return errors.Is(err, ErrNetworkAccessDenied) || errors.Is(err, ErrMemberAccessDenied)
}
