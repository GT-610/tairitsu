package services

import (
	"fmt"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// NetworkService 网络服务

type NetworkService struct {
	ztClient *zerotier.Client
}

// NewNetworkService 创建网络服务实例
func NewNetworkService(ztClient *zerotier.Client) *NetworkService {
	return &NetworkService{
		ztClient: ztClient,
	}
}

// GetStatus 获取ZeroTier网络状态
func (s *NetworkService) GetStatus() (*zerotier.Status, error) {
	logger.Info("服务层：开始获取ZeroTier网络状态")
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	status, err := s.ztClient.GetStatus()
	if err != nil {
		logger.Error("服务层：获取ZeroTier网络状态失败", zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：成功获取ZeroTier网络状态")

	return status, nil
}

// GetAllNetworks 获取所有网络
func (s *NetworkService) GetAllNetworks() ([]zerotier.Network, error) {
	logger.Info("服务层：开始获取所有网络列表")
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	networks, err := s.ztClient.GetNetworks()
	if err != nil {
		logger.Error("服务层：获取网络列表失败", zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：成功获取网络列表", zap.Int("network_count", len(networks)))

	return networks, nil
}

// GetNetworkByID 根据ID获取网络
func (s *NetworkService) GetNetworkByID(id string) (*zerotier.Network, error) {
	logger.Info("服务层：开始根据ID获取网络", zap.String("network_id", id))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	network, err := s.ztClient.GetNetwork(id)
	if err != nil {
		logger.Error("服务层：根据ID获取网络失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}
	
	if network == nil {
		logger.Warn("服务层：网络不存在", zap.String("network_id", id))
		return nil, nil
	}
	
	logger.Info("服务层：成功根据ID获取网络", zap.String("network_id", id))

	return network, nil
}

// CreateNetwork 创建新网络
func (s *NetworkService) CreateNetwork(network *zerotier.Network) (*zerotier.Network, error) {
	logger.Info("服务层：开始创建新网络", zap.String("network_name", network.Name))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	createdNetwork, err := s.ztClient.CreateNetwork(network)
	if err != nil {
		logger.Error("服务层：创建网络失败", zap.String("network_name", network.Name), zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：成功创建网络", zap.String("network_id", createdNetwork.ID), zap.String("network_name", createdNetwork.Name))

	return createdNetwork, nil
}

// UpdateNetwork 更新网络
func (s *NetworkService) UpdateNetwork(id string, network *zerotier.Network) (*zerotier.Network, error) {
	logger.Info("服务层：开始更新网络", zap.String("network_id", id))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	updatedNetwork, err := s.ztClient.UpdateNetwork(id, network)
	if err != nil {
		logger.Error("服务层：更新网络失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：成功更新网络", zap.String("network_id", updatedNetwork.ID))

	return updatedNetwork, nil
}

// DeleteNetwork 删除网络
func (s *NetworkService) DeleteNetwork(networkID string) error {
	logger.Info("服务层：开始删除网络", zap.String("network_id", networkID))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	err := s.ztClient.DeleteNetwork(networkID)
	if err != nil {
		logger.Error("服务层：删除网络失败", zap.String("network_id", networkID), zap.Error(err))
		return err
	}
	
	logger.Info("服务层：成功删除网络", zap.String("network_id", networkID))

	return nil
}

// GetNetworkMembers 获取网络中的所有成员
func (s *NetworkService) GetNetworkMembers(networkID string) ([]zerotier.Member, error) {
	logger.Info("服务层：开始获取网络中的所有成员", zap.String("network_id", networkID))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	members, err := s.ztClient.GetMembers(networkID)
	if err != nil {
		logger.Error("服务层：获取网络成员列表失败", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：成功获取网络成员列表", zap.String("network_id", networkID), zap.Int("member_count", len(members)))

	return members, nil
}

// GetNetworkMember 获取网络中的特定成员
func (s *NetworkService) GetNetworkMember(networkID, memberID string) (*zerotier.Member, error) {
	logger.Info("服务层：开始获取网络中的特定成员", zap.String("network_id", networkID), zap.String("member_id", memberID))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
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
	
	logger.Info("服务层：成功获取网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	return member, nil
}

// UpdateNetworkMember 更新网络成员
func (s *NetworkService) UpdateNetworkMember(networkID, memberID string, member *zerotier.Member) (*zerotier.Member, error) {
	logger.Info("服务层：开始更新网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	updatedMember, err := s.ztClient.UpdateMember(networkID, memberID, member)
	if err != nil {
		logger.Error("服务层：更新网络成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：成功更新网络成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	return updatedMember, nil
}

// RemoveNetworkMember 从网络中移除成员
func (s *NetworkService) RemoveNetworkMember(networkID, memberID string) error {
	logger.Info("服务层：开始从网络中移除成员", zap.String("network_id", networkID), zap.String("member_id", memberID))
	
	// 检查ZeroTier客户端是否已初始化
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return fmt.Errorf("ZeroTier客户端未初始化")
	}
	
	err := s.ztClient.DeleteMember(networkID, memberID)
	if err != nil {
		logger.Error("服务层：从网络中移除成员失败", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return err
	}
	
	logger.Info("服务层：成功从网络中移除成员", zap.String("network_id", networkID), zap.String("member_id", memberID))

	return nil
}

// SetZTClient 设置ZeroTier客户端
func (s *NetworkService) SetZTClient(client *zerotier.Client) {
	if client != nil {
		logger.Info("服务层：ZeroTier客户端已设置")
	} else {
		logger.Warn("服务层：尝试设置空的ZeroTier客户端")
	}
	s.ztClient = client
}