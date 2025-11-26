package services

import (
	"fmt"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// 全局ZeroTier客户端实例，在初始化后保持
var GlobalZTClient *zerotier.Client

// NetworkService 网络服务

type NetworkService struct {
	ztClient *zerotier.Client
}

// NewNetworkService 创建网络服务实例
func NewNetworkService(ztClient *zerotier.Client) *NetworkService {
	// 优先使用全局ZeroTier客户端（如果已初始化）
	// 这确保在路由重新加载后，新创建的NetworkService实例也能使用已初始化的客户端
	if GlobalZTClient != nil && ztClient == nil {
		ztClient = GlobalZTClient
	}
	return &NetworkService{
		ztClient: ztClient,
	}
}

// GetStatus 获取ZeroTier网络状态
func (s *NetworkService) GetStatus() (*zerotier.Status, error) {
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

	return status, nil
}

// GetAllNetworks 获取所有网络
func (s *NetworkService) GetAllNetworks() ([]zerotier.Network, error) {
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

	return networks, nil
}

// GetNetworkByID 根据ID获取网络
func (s *NetworkService) GetNetworkByID(id string) (*zerotier.Network, error) {
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

	return network, nil
}

// CreateNetwork 创建新网络
func (s *NetworkService) CreateNetwork(network *zerotier.Network) (*zerotier.Network, error) {
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

	return createdNetwork, nil
}

// UpdateNetwork 更新网络
func (s *NetworkService) UpdateNetwork(id string, network *zerotier.Network) (*zerotier.Network, error) {
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

	return updatedNetwork, nil
}

// DeleteNetwork 删除网络
func (s *NetworkService) DeleteNetwork(networkID string) error {
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

	return nil
}

// GetNetworkMembers 获取网络中的所有成员
func (s *NetworkService) GetNetworkMembers(networkID string) ([]zerotier.Member, error) {
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

	return members, nil
}

// GetNetworkMember 获取网络中的特定成员
func (s *NetworkService) GetNetworkMember(networkID, memberID string) (*zerotier.Member, error) {
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

	return member, nil
}

// UpdateNetworkMember 更新网络成员
func (s *NetworkService) UpdateNetworkMember(networkID, memberID string, member *zerotier.Member) (*zerotier.Member, error) {
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

	return updatedMember, nil
}

// RemoveNetworkMember 从网络中移除成员
func (s *NetworkService) RemoveNetworkMember(networkID, memberID string) error {
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