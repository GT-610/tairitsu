package services

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// NetworkService handles network-related operations with ZeroTier
type NetworkService struct {
	ztClient *zerotier.Client     // ZeroTier client for API interactions
	db       database.Database    // Database for network ownership management
}

// NewNetworkService creates a new network service instance
func NewNetworkService(ztClient *zerotier.Client, db database.Database) *NetworkService {
	return &NetworkService{
		ztClient: ztClient,
		db:       db,
	}
}

// GetStatus retrieves the current ZeroTier network status
func (s *NetworkService) GetStatus() (*zerotier.Status, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	status, err := s.ztClient.GetStatus()
	if err != nil {
		logger.Error("NetworkService: Failed to get ZeroTier status", zap.Error(err))
		return nil, err
	}

	return status, nil
}

// GetAllNetworks retrieves all networks owned by a specific user
func (s *NetworkService) GetAllNetworks(ownerID string) ([]zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Get network IDs owned by the user from database
	ownedNetworks, err := s.db.GetNetworksByOwnerID(ownerID)
	if err != nil {
		logger.Error("NetworkService: Failed to get user networks", zap.String("owner_id", ownerID), zap.Error(err))
		return nil, err
	}

	// Return empty slice if no networks owned
	if len(ownedNetworks) == 0 {
		return []zerotier.Network{}, nil
	}

	// Get all networks from ZeroTier
	allNetworks, err := s.ztClient.GetNetworks()
	if err != nil {
		logger.Error("NetworkService: Failed to get networks", zap.Error(err))
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
			filteredNetworks = append(filteredNetworks, net)
		}
	}

	return filteredNetworks, nil
}

// GetNetworkByID retrieves a network by its ID with ownership check
func (s *NetworkService) GetNetworkByID(id string, userID string) (*zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(id)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", id))
		return nil, nil
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to access network", zap.String("network_id", id), zap.String("user_id", userID))
		return nil, nil
	}

	// Get network details from ZeroTier
	network, err := s.ztClient.GetNetwork(id)
	if err != nil {
		logger.Error("NetworkService: Failed to get network by ID", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	if network == nil {
		logger.Warn("NetworkService: Network not found", zap.String("network_id", id))
		return nil, nil
	}

	return network, nil
}

// CreateNetwork creates a new ZeroTier network with ownership
func (s *NetworkService) CreateNetwork(network *zerotier.Network, ownerID string) (*zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Force network to be private
	network.Config.Private = true

	// Create network in ZeroTier
	createdNetwork, err := s.ztClient.CreateNetwork(network)
	if err != nil {
		logger.Error("NetworkService: Failed to create network", zap.String("network_name", network.Name), zap.Error(err))
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
		logger.Error("NetworkService: Failed to save network ownership", zap.String("network_id", createdNetwork.ID), zap.Error(err))
		// Try to delete network from ZeroTier if database save fails
		if delErr := s.ztClient.DeleteNetwork(createdNetwork.ID); err != nil {
			logger.Error("NetworkService: Failed to rollback network creation", zap.String("network_id", createdNetwork.ID), zap.Error(delErr))
		}
		return nil, err
	}

	return createdNetwork, nil
}

// UpdateNetwork updates a network with ownership check and private network enforcement
func (s *NetworkService) UpdateNetwork(id string, network *zerotier.Network, userID string) (*zerotier.Network, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(id)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", id))
		return nil, fmt.Errorf("Network not found or no access")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to update network", zap.String("network_id", id), zap.String("user_id", userID))
		return nil, fmt.Errorf("No permission to update network")
	}

	// Force network to remain private
	network.Config.Private = true

	// Update network in ZeroTier
	updatedNetwork, err := s.ztClient.UpdateNetwork(id, network)
	if err != nil {
		logger.Error("NetworkService: Failed to update network", zap.String("network_id", id), zap.Error(err))
		return nil, err
	}

	// Update network in database
	ownedNetwork.Name = updatedNetwork.Name
	ownedNetwork.Description = updatedNetwork.Description
	if err := s.db.UpdateNetwork(ownedNetwork); err != nil {
		logger.Error("NetworkService: Failed to update network in database", zap.String("network_id", id), zap.Error(err))
		// Continue anyway, as ZeroTier update was successful
	}

	return updatedNetwork, nil
}

// DeleteNetwork deletes a network with ownership check
func (s *NetworkService) DeleteNetwork(networkID string, userID string) error {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", networkID), zap.Error(err))
		return err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", networkID))
		return fmt.Errorf("Network not found or no access")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to delete network", zap.String("network_id", networkID), zap.String("user_id", userID))
		return fmt.Errorf("No permission to delete network")
	}

	// Delete network from ZeroTier
	err = s.ztClient.DeleteNetwork(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to delete network", zap.String("network_id", networkID), zap.Error(err))
		return err
	}

	// Delete network from database
	if err := s.db.DeleteNetwork(networkID); err != nil {
		logger.Error("NetworkService: Failed to delete network from database", zap.String("network_id", networkID), zap.Error(err))
		// Continue anyway, as ZeroTier deletion was successful
	}

	return nil
}

// GetNetworkMembers retrieves all members of a network with ownership check
func (s *NetworkService) GetNetworkMembers(networkID string, userID string) ([]zerotier.Member, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", networkID))
		return nil, fmt.Errorf("Network not found or no access")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to access network members", zap.String("network_id", networkID), zap.String("user_id", userID))
		return nil, fmt.Errorf("No permission to access network members")
	}

	members, err := s.ztClient.GetMembers(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network members", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	return members, nil
}

// GetNetworkMember retrieves a specific member of a network with ownership check
func (s *NetworkService) GetNetworkMember(networkID, memberID string, userID string) (*zerotier.Member, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", networkID))
		return nil, fmt.Errorf("Network not found or no access")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to access network member", zap.String("network_id", networkID), zap.String("user_id", userID))
		return nil, fmt.Errorf("No permission to access network member")
	}

	member, err := s.ztClient.GetMember(networkID, memberID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return nil, err
	}

	if member == nil {
		logger.Warn("NetworkService: Network member not found", zap.String("network_id", networkID), zap.String("member_id", memberID))
		return nil, nil
	}

	return member, nil
}

// UpdateNetworkMember updates a network member with ownership check
func (s *NetworkService) UpdateNetworkMember(networkID, memberID string, member *zerotier.Member, userID string) (*zerotier.Member, error) {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return nil, fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return nil, fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", networkID), zap.Error(err))
		return nil, err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", networkID))
		return nil, fmt.Errorf("Network not found or no access")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to update network member", zap.String("network_id", networkID), zap.String("user_id", userID))
		return nil, fmt.Errorf("No permission to update network member")
	}

	updatedMember, err := s.ztClient.UpdateMember(networkID, memberID, member)
	if err != nil {
		logger.Error("NetworkService: Failed to update network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return nil, err
	}

	return updatedMember, nil
}

// RemoveNetworkMember removes a member from a network with ownership check
func (s *NetworkService) RemoveNetworkMember(networkID, memberID string, userID string) error {
	// Check if ZeroTier client is initialized
	if s.ztClient == nil {
		logger.Warn("NetworkService: ZeroTier client not initialized")
		return fmt.Errorf("ZeroTier client not initialized")
	}

	// Check if database is initialized
	if s.db == nil {
		logger.Warn("NetworkService: Database not initialized")
		return fmt.Errorf("Database not initialized")
	}

	// Check if network is owned by the user
	ownedNetwork, err := s.db.GetNetworkByID(networkID)
	if err != nil {
		logger.Error("NetworkService: Failed to get network ownership", zap.String("network_id", networkID), zap.Error(err))
		return err
	}

	if ownedNetwork == nil {
		logger.Warn("NetworkService: Network not found or no access", zap.String("network_id", networkID))
		return fmt.Errorf("Network not found or no access")
	}

	// Check if user is the owner
	if ownedNetwork.OwnerID != userID {
		logger.Warn("NetworkService: No permission to remove network member", zap.String("network_id", networkID), zap.String("user_id", userID))
		return fmt.Errorf("No permission to remove network member")
	}

	err = s.ztClient.DeleteMember(networkID, memberID)
	if err != nil {
		logger.Error("NetworkService: Failed to remove network member", zap.String("network_id", networkID), zap.String("member_id", memberID), zap.Error(err))
		return err
	}

	return nil
}

// SetZTClient sets the ZeroTier client
func (s *NetworkService) SetZTClient(client *zerotier.Client) {
	if client == nil {
		logger.Warn("NetworkService: Attempting to set nil ZeroTier client")
	}
	s.ztClient = client
}
