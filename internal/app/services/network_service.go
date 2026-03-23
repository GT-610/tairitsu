package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

var GlobalZTClient *zerotier.Client

type NetworkService struct {
	ztClient *zerotier.Client
	db       database.DBInterface
	mutex    sync.RWMutex
}

func NewNetworkService(ztClient *zerotier.Client, db database.DBInterface) *NetworkService {
	if GlobalZTClient != nil && ztClient == nil {
		ztClient = GlobalZTClient
	}
	return &NetworkService{
		ztClient: ztClient,
		db:       db,
	}
}

func (s *NetworkService) SetDB(db database.DBInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.db = db
	logger.Info("网络服务数据库连接已更新")
}

func (s *NetworkService) getDB() database.DBInterface {
	s.mutex.RLock()
	db := s.db
	s.mutex.RUnlock()

	if db == nil {
		return database.GetGlobalDB()
	}
	return db
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

// NetworkSummary 网络摘要信息（从数据库获取）
type NetworkSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NetworkDetail 网络详情（包含控制器信息和数据库描述）
type NetworkDetail struct {
	*zerotier.Network
	DBDescription string `json:"db_description"`
}

// GetAllNetworks retrieves all networks owned by a specific user from database
func (s *NetworkService) GetAllNetworks(ownerID string) ([]NetworkSummary, error) {
	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetworks, err := db.GetNetworksByOwnerID(ownerID)
	if err != nil {
		logger.Error("服务层：获取用户网络列表失败", zap.String("owner_id", ownerID), zap.Error(err))
		return nil, err
	}

	// If no networks owned, return empty slice
	if len(ownedNetworks) == 0 {
		return []NetworkSummary{}, nil
	}

	// Convert to NetworkSummary
	var networkSummaries []NetworkSummary
	for _, net := range ownedNetworks {
		networkSummaries = append(networkSummaries, NetworkSummary{
			ID:          net.ID,
			Name:        net.Name,
			Description: net.Description,
			OwnerID:     net.OwnerID,
			MemberCount: 0,
			CreatedAt:   net.CreatedAt,
			UpdatedAt:   net.UpdatedAt,
		})
	}

	return networkSummaries, nil
}

func (s *NetworkService) GetNetworkByID(id string, userID string) (*NetworkDetail, error) {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(id)
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

	logger.Info("服务层：获取网络详情成功",
		zap.String("network_id", id),
		zap.Any("ipAssignmentPools", network.Config.IpAssignmentPools),
		zap.String("db_description", ownedNetwork.Description))

	// Return combined network detail with database description
	return &NetworkDetail{
		Network:       network,
		DBDescription: ownedNetwork.Description,
	}, nil
}

// CreateNetwork creates a new ZeroTier network with ownership
func (s *NetworkService) CreateNetwork(network *zerotier.Network, ownerID string) (*zerotier.Network, error) {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	network.Config.Private = true

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

	if err := db.CreateNetwork(dbNetwork); err != nil {
		logger.Error("服务层：保存网络所有权失败", zap.String("network_id", createdNetwork.ID), zap.Error(err))
		// Try to delete network from ZeroTier if database save fails
		if delErr := s.ztClient.DeleteNetwork(createdNetwork.ID); delErr != nil {
			logger.Error("服务层：回滚删除网络失败", zap.String("network_id", createdNetwork.ID), zap.Error(delErr))
		}
		return nil, err
	}

	return createdNetwork, nil
}

// UpdateNetwork 更新网络 with ownership check and private network enforcement
func (s *NetworkService) UpdateNetwork(id string, updateReq *zerotier.NetworkUpdateRequest, userID string) (*zerotier.Network, error) {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(id)
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
	updateReq.Private = true

	// Update network in ZeroTier using partial update
	updatedNetwork, err := s.ztClient.PartialUpdateNetwork(id, updateReq)
	if err != nil {
		logger.Error("服务层：更新网络失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	// Update network name and description in database
	if updateReq.Name != "" {
		ownedNetwork.Name = updateReq.Name
	}
	if updateReq.Description != "" {
		ownedNetwork.Description = updateReq.Description
	}
	if updateReq.Name != "" || updateReq.Description != "" {
		if err := db.UpdateNetwork(ownedNetwork); err != nil {
			logger.Error("服务层：更新数据库中网络信息失败", zap.String("network_id", id), zap.Error(err))
		}
	}

	return updatedNetwork, nil
}

func (s *NetworkService) UpdateNetworkMetadata(id string, name string, description string, userID string) (*zerotier.Network, error) {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(id)
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

	// 更新控制器中的网络名称（名称同步更新）
	updateReq := &zerotier.NetworkUpdateRequest{
		Name: name,
	}
	updatedNetwork, err := s.ztClient.PartialUpdateNetwork(id, updateReq)
	if err != nil {
		logger.Error("服务层：更新控制器网络名称失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	// 更新数据库中的名称和描述（描述只更新数据库）
	ownedNetwork.Name = name
	ownedNetwork.Description = description
	if err := db.UpdateNetwork(ownedNetwork); err != nil {
		logger.Error("服务层：更新数据库网络信息失败", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	logger.Info("成功更新网络元数据", zap.String("network_id", id), zap.String("name", name))

	return updatedNetwork, nil
}

// DeleteNetwork 删除网络 with ownership check
func (s *NetworkService) DeleteNetwork(networkID string, userID string) error {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(networkID)
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
	if err := db.DeleteNetwork(networkID); err != nil {
		logger.Error("服务层：从数据库删除网络失败", zap.String("network_id", networkID), zap.Error(err))
		// Continue anyway, as ZeroTier deletion was successful
	}

	return nil
}

// GetNetworkMembers 获取网络中的所有成员 with ownership check
func (s *NetworkService) GetNetworkMembers(networkID string, userID string) ([]zerotier.Member, error) {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(networkID)
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
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(networkID)
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
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(networkID)
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
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return fmt.Errorf("数据库未初始化")
	}

	ownedNetwork, err := db.GetNetworkByID(networkID)
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

// ImportableNetworkSummary 可导入网络摘要（轻量级，只包含ID）
type ImportableNetworkSummary struct {
	NetworkID    string `json:"network_id"`
	Reason       string `json:"reason"`
	IsImportable bool   `json:"is_importable"`
}

// GetImportableNetworks 获取可导入的网络ID列表（轻量级）
func (s *NetworkService) GetImportableNetworks(userID string) ([]ImportableNetworkSummary, error) {
	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	ztNetworkIDs, err := s.ztClient.GetNetworkIDs()
	if err != nil {
		logger.Error("服务层：获取ZeroTier网络ID列表失败", zap.Error(err))
		return nil, err
	}

	dbNetworks, err := db.GetAllNetworks()
	if err != nil {
		logger.Error("服务层：获取数据库网络列表失败", zap.Error(err))
		return nil, err
	}

	// 创建网络ID到数据库记录的映射
	dbNetworkMap := make(map[string]*models.Network)
	for _, net := range dbNetworks {
		dbNetworkMap[net.ID] = net
	}

	// 筛选可导入的网络
	var importableNetworks []ImportableNetworkSummary
	for _, networkID := range ztNetworkIDs {
		dbNet, exists := dbNetworkMap[networkID]

		importable := ImportableNetworkSummary{
			NetworkID:    networkID,
			Reason:       "",
			IsImportable: false,
		}

		if !exists {
			importable.Reason = "网络不在数据库中"
			importable.IsImportable = true
		} else if dbNet.OwnerID == "" {
			importable.Reason = "网络无所有者"
			importable.IsImportable = true
		} else if dbNet.OwnerID != userID {
			importable.Reason = "网络属于其他用户"
			importable.IsImportable = false
		} else {
			importable.Reason = "您已是该网络的所有者"
			importable.IsImportable = false
		}

		importableNetworks = append(importableNetworks, importable)
	}

	return importableNetworks, nil
}

// ImportNetworks 导入指定的网络
func (s *NetworkService) ImportNetworks(networkIDs []string, userID string) ([]string, error) {
	if s.ztClient == nil {
		logger.Warn("服务层：ZeroTier客户端未初始化")
		return nil, fmt.Errorf("ZeroTier客户端未初始化")
	}

	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化")
		return nil, fmt.Errorf("数据库未初始化")
	}

	dbNetworks, err := db.GetAllNetworks()
	if err != nil {
		logger.Error("服务层：获取数据库网络列表失败", zap.Error(err))
		return nil, err
	}

	// 创建网络ID到数据库记录的映射
	dbNetworkMap := make(map[string]*models.Network)
	for _, net := range dbNetworks {
		dbNetworkMap[net.ID] = net
	}

	// 获取所有ZeroTier网络ID列表
	ztNetworkIDs, err := s.ztClient.GetNetworkIDs()
	if err != nil {
		logger.Error("服务层：获取ZeroTier网络ID列表失败", zap.Error(err))
		return nil, err
	}

	// 创建ZeroTier网络ID集合
	ztNetworkSet := make(map[string]bool)
	for _, id := range ztNetworkIDs {
		ztNetworkSet[id] = true
	}

	var importedIDs []string
	now := time.Now()

	for _, networkID := range networkIDs {
		if !ztNetworkSet[networkID] {
			logger.Warn("服务层：网络不存在于ZeroTier控制器", zap.String("network_id", networkID))
			continue
		}

		dbNet, dbExists := dbNetworkMap[networkID]

		if !dbExists {
			// 从ZeroTier获取完整网络信息
			ztNet, err := s.ztClient.GetNetwork(networkID)
			if err != nil {
				logger.Error("服务层：获取网络详情失败", zap.String("network_id", networkID), zap.Error(err))
				continue
			}

			// 创建新记录，包含完整信息
			newNetwork := &models.Network{
				ID:          networkID,
				Name:        ztNet.Name,
				Description: ztNet.Description,
				OwnerID:     userID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := db.CreateNetwork(newNetwork); err != nil {
				logger.Error("服务层：创建网络记录失败", zap.String("network_id", networkID), zap.Error(err))
				continue
			}

			logger.Info("服务层：成功导入网络", zap.String("network_id", networkID), zap.String("network_name", ztNet.Name))
			importedIDs = append(importedIDs, networkID)
		} else if dbNet.OwnerID == "" {
			// 网络在数据库中但没有所有者，更新所有者
			dbNet.OwnerID = userID
			dbNet.UpdatedAt = now

			if err := db.UpdateNetwork(dbNet); err != nil {
				logger.Error("服务层：更新网络所有者失败", zap.String("network_id", networkID), zap.Error(err))
				continue
			}

			logger.Info("服务层：成功认领网络", zap.String("network_id", networkID))
			importedIDs = append(importedIDs, networkID)
		} else if dbNet.OwnerID != userID {
			logger.Warn("服务层：网络属于其他用户，无法导入", zap.String("network_id", networkID), zap.String("owner_id", dbNet.OwnerID))
			continue
		} else {
			logger.Info("服务层：网络已属于当前用户，跳过", zap.String("network_id", networkID))
		}
	}

	return importedIDs, nil
}
