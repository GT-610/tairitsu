package services

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// GlobalZTClient is the global ZeroTier client instance, maintained after initialization
var GlobalZTClient *zerotier.Client

// NetworkService handles network-related operations with ZeroTier
// This service provides methods to manage ZeroTier networks

type NetworkService struct {
	ztClient *zerotier.Client     // ZeroTier client for API interactions
	db       database.DBInterface // Database interface for network ownership management
}

// NewNetworkService creates a new network service instance
// If global ZT client is available and no specific client is provided,
// it will use the global client to ensure continuity after route reloading
func NewNetworkService(ztClient *zerotier.Client, db database.DBInterface) *NetworkService {
	if GlobalZTClient != nil && ztClient == nil {
		ztClient = GlobalZTClient
	}
	return &NetworkService{
		ztClient: ztClient,
		db:       db,
	}
}

// GetStatus retrieves the current ZeroTier network status
func (s *NetworkService) GetStatus() (*zerotier.Status, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	status, err := s.ztClient.GetStatus()
	if err != nil {
		logger.Error("服务层：获取ZeroTier网络状态失败", zap.Error(err))
		return nil, err
	}

	return status, nil
}

// GetAllNetworks retrieves all networks owned by a specific user
func (s *NetworkService) GetAllNetworks(ownerID string) ([]zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Get network IDs owned by the user from database
	ownedNetworks, err := s.db.GetNetworksByOwnerID(ownerID)
	if err != nil {
		logger.Error("服务层：获取用户网络列表失败", zap.String("owner_id", ownerID), zap.Error(err))
		return nil, err
	}

	// If no networks owned, return empty slice
	if len(ownedNetworks) == 0 {
		return []zerotier.Network{}, nil
	}

	// Get all networks from ZeroTier
	allNetworks, err := s.ztClient.GetNetworks()
	if err != nil {
		logger.Error("服务层：获取网络列表失败", zap.Error(err))
		return nil, err
	}

	// Filter networks to only include owned ones
	var filteredNetworks []zerotier.Network
	ownedNetworkIDs := make(map[string]bool)
	for _, net := range ownedNetworks {
		ownedNetworkIDs[net.ID] = true
	}

	for _, net := range allNetworks {
		if ownedNetworkIDs[net.ID] {
			// 获取网络状态
			net.Status = s.getNetworkStatus(net.ID)
			filteredNetworks = append(filteredNetworks, net)
		}
	}

	return filteredNetworks, nil
}

// getNetworkStatus 获取网络状态
func (s *NetworkService) getNetworkStatus(networkID string) string {
	status, err := s.ztClient.GetNetworkStatus(networkID)
	if err != nil {
		logger.Warn("获取网络状态失败", zap.String("network_id", networkID), zap.Error(err))
		return "unknown"
	}
	return status
}

// GetNetworkByID retrieves a network by its ID with ownership check
func (s *NetworkService) GetNetworkByID(id string, userID string) (*zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(id)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", id))
		return nil, nil
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限访问网络", zap.String("network_id", id), zap.String("user_id", userID))
		return nil, nil
	}

	// Get network details from ZeroTier
	network, err := s.ztClient.GetNetwork(id)
	if err != nil {
		logger.Error("服务层：根据ID获取网络失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	if network == nil {
		logger.Warn("服务层：网络不存在", zap.String("network_id", id))
		return nil, nil
	}

	return network, nil
}

// CreateNetwork creates a new ZeroTier network with ownership
func (s *NetworkService) CreateNetwork(network *zerotier.Network, ownerID string) (*zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Force network to be private
	network.Config.Private = true

	// Create network in ZeroTier
	createdNetwork, err := s.ztClient.CreateNetwork(network)
	if err != nil {
		logger.Error("服务层：创建网络失败", zap.String("network_name", network.Name), zap.Error(err))
		return nil, err
	}

	// Save network to database with owner information
	dbNetwork := &models.Network{
		ID:          createdNetwork.ID,
		Name:        createdNetwork.Name,
		Description: createdNetwork.Description,
		OwnerID:     ownerID,
	}

	if err := s.db.CreateNetwork(dbNetwork); err != nil {
		logger.Error("服务层：保存网络所有权失败", zap.String("network_id", createdNetwork.ID), zap.Error(err))
		// Try to delete network from ZeroTier if database save fails
		if delErr := s.ztClient.DeleteNetwork(createdNetwork.ID); err != nil {
			logger.Error("服务层：回滚删除网络失败", zap.String("network_id", createdNetwork.ID), zap.Error(delErr))
		}
		return nil, err
	}

	return createdNetwork, nil
}

// UpdateNetwork 更新网络 with ownership check and private network enforcement
func (s *NetworkService) UpdateNetwork(id string, network *zerotier.Network, userID string) (*zerotier.Network, error) {
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(id)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", id))
		return nil, fmt.Errorf("网络不存在或无权限访问")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限更新网络", zap.String("network_id", id), zap.String("user_id", userID))
		return nil, fmt.Errorf("无权限更新网络")
	}

	// Force network to remain private
	network.Config.Private = true

	// Update network in ZeroTier
	updatedNetwork, err := s.ztClient.UpdateNetwork(id, network)
	if err != nil {
		logger.Error("服务层：更新网络失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	// Update network in database
	ownedNetwork.Name = updatedNetwork.Name
	ownedNetwork.Description = updatedNetwork.Description
	if err := s.db.UpdateNetwork(ownedNetwork); err != nil {
		logger.Error("服务层：更新数据库中网络信息失败", zap.String("network_id", id), zap.Error(err))
		// Continue anyway, as ZeroTier update was successful
	}

	return updatedNetwork, nil
}

// DeleteNetwork 删除网络 with ownership check
func (s *NetworkService) DeleteNetwork(networkID string, userID string) error {
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", networkID), zap.Error(err))
		return err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", networkID))
		return fmt.Errorf("网络不存在或无权限访问")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限删除网络", zap.String("network_id", networkID), zap.String("user_id", userID))
		return fmt.Errorf("无权限删除网络")
	}

	// Delete network from ZeroTier
	err = s.ztClient.DeleteNetwork(networkID)
	if err != nil {
		logger.Error("服务层：删除网络失败", zap.String("network_id", networkID), zap.Error(err))
		return err
	}

	// Delete network from database
	if err := s.db.DeleteNetwork(networkID); err != nil {
		logger.Error("服务层：从数据库删除网络失败", zap.String("network_id", networkID), zap.Error(err))
		// Continue anyway, as ZeroTier deletion was successful
	}

	return nil
}

// GetNetworkMembers 获取网络中的所有成员 with ownership check
func (s *NetworkService) GetNetworkMembers(networkID string, userID string) ([]zerotier.Member, error) {
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", networkID))
		return nil, fmt.Errorf("网络不存在或无权限访问")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限访问网络成员", zap.String("network_id", networkID), zap.String("user_id", userID))
		return nil, fmt.Errorf("无权限访问网络成员")
	}

	members, err := s.ztClient.GetMembers(networkID)
	if err != nil {
		logger.Error("服务层：获取网络成员列表失败", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	return members, nil
}

// GetNetworkMember 获取网络中的特定成员 with ownership check
func (s *NetworkService) GetNetworkMember(networkID, memberID string, userID string) (*zerotier.Member, error) {
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", networkID))
		return nil, fmt.Errorf("网络不存在或无权限访问")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限访问网络成员", zap.String("network_id", networkID), zap.String("user_id", userID))
		return nil, fmt.Errorf("无权限访问网络成员")
	}

	member, err := s.ztClient.GetMember(networkID, memberID)
	if err != nil {
		logger.Error("服务层：获取网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return nil, err
	}

	if member == nil {
		logger.Warn("服务层：网络成员不存在", zap.String("network_id", networkID), zap.String("member_id", memberID))
		return nil, nil
	}

	return member, nil
}

// UpdateNetworkMember 更新网络成员 with ownership check
func (s *NetworkService) UpdateNetworkMember(networkID, memberID string, member *zerotier.Member, userID string) (*zerotier.Member, error) {
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", networkID))
		return nil, fmt.Errorf("网络不存在或无权限访问")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限更新网络成员", zap.String("network_id", networkID), zap.String("user_id", userID))
		return nil, fmt.Errorf("无权限更新网络成员")
	}

	updatedMember, err := s.ztClient.UpdateMember(networkID, memberID, member)
	if err != nil {
		logger.Error("服务层：更新网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return nil, err
	}

	return updatedMember, nil
}

// RemoveNetworkMember 从网络中移除成员 with ownership check
func (s *NetworkService) RemoveNetworkMember(networkID, memberID string, userID string) error {
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return fmt.Errorf("ZeroTier客户端未初始化")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化")
		return fmt.Errorf("数据库未初始化")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("服务层：获取网络所有权失败", zap.String("network_id", networkID), zap.Error(err))
		return err
	}

	if ownedNetwork == nil {
		logger.Warn("服务层：网络不存在或无权限访问", zap.String("network_id", networkID))
		return fmt.Errorf("网络不存在或无权限访问")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("服务层：无权限移除网络成员", zap.String("network_id", networkID), zap.String("user_id", userID))
		return fmt.Errorf("无权限移除网络成员")
	}

	err = s.ztClient.DeleteMember(networkID, memberID)
	if err != nil {
		logger.Error("服务层：从网络中移除成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return err
	}

	return nil
}

// SetZTClient 设置ZeroTier客户端
func (s *NetworkService) SetZTClient(client *zerotier.Client) {
	if client != nil {
		// 同时更新全局客户端实例，确保路由重新加载后仍然保持
		GlobalZTClient = client
	} else {
		logger.Warn("服务层：尝试设置空的ZeroTier客户端")
		// 也清空全局实例
		GlobalZTClient = nil
	}
	s.ztClient = client
}
