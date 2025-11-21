package services

import (
	"github.com/tairitsu/tairitsu/internal/zerotier"
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

// GetStatus 获取ZeroTier状态
func (s *NetworkService) GetStatus() (*zerotier.Status, error) {
	return s.ztClient.GetStatus()
}

// GetAllNetworks 获取所有网络
func (s *NetworkService) GetAllNetworks() ([]zerotier.Network, error) {
	return s.ztClient.GetNetworks()
}

// GetNetworkByID 根据ID获取网络
func (s *NetworkService) GetNetworkByID(networkID string) (*zerotier.Network, error) {
	return s.ztClient.GetNetwork(networkID)
}

// CreateNetwork 创建网络
func (s *NetworkService) CreateNetwork(network *zerotier.Network) (*zerotier.Network, error) {
	return s.ztClient.CreateNetwork(network)
}

// UpdateNetwork 更新网络
func (s *NetworkService) UpdateNetwork(networkID string, network *zerotier.Network) (*zerotier.Network, error) {
	return s.ztClient.UpdateNetwork(networkID, network)
}

// DeleteNetwork 删除网络
func (s *NetworkService) DeleteNetwork(networkID string) error {
	return s.ztClient.DeleteNetwork(networkID)
}

// GetNetworkMembers 获取网络成员列表
func (s *NetworkService) GetNetworkMembers(networkID string) ([]zerotier.Member, error) {
	return s.ztClient.GetMembers(networkID)
}

// GetNetworkMember 获取单个网络成员
func (s *NetworkService) GetNetworkMember(networkID, memberID string) (*zerotier.Member, error) {
	return s.ztClient.GetMember(networkID, memberID)
}

// UpdateNetworkMember 更新网络成员
func (s *NetworkService) UpdateNetworkMember(networkID, memberID string, member *zerotier.Member) (*zerotier.Member, error) {
	return s.ztClient.UpdateMember(networkID, memberID, member)
}

// RemoveNetworkMember 移除网络成员
func (s *NetworkService) RemoveNetworkMember(networkID, memberID string) error {
	return s.ztClient.DeleteMember(networkID, memberID)
}