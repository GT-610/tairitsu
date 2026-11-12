package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"go.uber.org/zap"
)

// NetworkAccessEntry is one delegated account role on a network.
type NetworkAccessEntry struct {
	UserID            string      `json:"user_id"`
	Username          string      `json:"username"`
	Role              NetworkRole `json:"role"`
	GrantedBy         string      `json:"granted_by"`
	GrantedByUsername string      `json:"granted_by_username"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// NetworkAccessDetail is the access list for one network. Ownership is returned
// separately from the delegated entries so the UI can pin the owner row.
type NetworkAccessDetail struct {
	OwnerID       string               `json:"owner_id"`
	OwnerUsername string               `json:"owner_username"`
	MyRole        NetworkRole          `json:"my_role"`
	Entries       []NetworkAccessEntry `json:"entries"`
}

// NetworkAccessCandidate is an account that can still be added to a network.
type NetworkAccessCandidate struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (s *NetworkService) GetNetworkAccess(networkID, actorID string) (*NetworkAccessDetail, error) {
	network, actorRole, err := s.authorizeNetwork(networkID, actorID, NetworkPermissionAccessRead)
	if err != nil {
		return nil, err
	}

	db := s.getDB()
	if db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	grants, err := db.GetNetworkGrants(network.ID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(grants)*2+1)
	userIDs = append(userIDs, network.OwnerID)
	for _, grant := range grants {
		if grant == nil || grant.SubjectType != models.NetworkGrantSubjectUser {
			continue
		}
		userIDs = append(userIDs, grant.SubjectID, grant.GrantedBy)
	}

	users, err := db.GetUsersByIDs(uniqueNonEmpty(userIDs))
	if err != nil {
		return nil, err
	}
	usersByID := make(map[string]*models.User, len(users))
	for _, user := range users {
		if user != nil {
			usersByID[user.ID] = user
		}
	}

	detail := &NetworkAccessDetail{
		OwnerID: network.OwnerID,
		MyRole:  actorRole,
		Entries: make([]NetworkAccessEntry, 0, len(grants)),
	}
	if owner := usersByID[network.OwnerID]; owner != nil {
		detail.OwnerUsername = owner.Username
	}

	for _, grant := range grants {
		if grant == nil || grant.SubjectType != models.NetworkGrantSubjectUser {
			continue
		}
		role, ok := ParseNetworkRole(grant.Role)
		if !ok {
			logger.Warn("service: ignoring network grant with unknown role",
				zap.String("network_id", network.ID),
				zap.String("user_id", grant.SubjectID),
				zap.String("role", grant.Role))
			continue
		}
		user := usersByID[grant.SubjectID]
		if user == nil {
			continue
		}
		entry := NetworkAccessEntry{
			UserID:    user.ID,
			Username:  user.Username,
			Role:      role,
			GrantedBy: grant.GrantedBy,
			CreatedAt: grant.CreatedAt,
			UpdatedAt: grant.UpdatedAt,
		}
		if grantor := usersByID[grant.GrantedBy]; grantor != nil {
			entry.GrantedByUsername = grantor.Username
		}
		detail.Entries = append(detail.Entries, entry)
	}

	sort.Slice(detail.Entries, func(i, j int) bool {
		left, right := detail.Entries[i], detail.Entries[j]
		if left.Role.Rank() != right.Role.Rank() {
			return left.Role.Rank() > right.Role.Rank()
		}
		return strings.ToLower(left.Username) < strings.ToLower(right.Username)
	})

	return detail, nil
}

func (s *NetworkService) GetNetworkAccessCandidates(networkID, actorID, query string) ([]NetworkAccessCandidate, error) {
	network, _, err := s.authorizeNetwork(networkID, actorID, NetworkPermissionAccessWrite)
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return []NetworkAccessCandidate{}, nil
	}

	db := s.getDB()
	if db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	users, err := db.SearchUsers(query, 20)
	if err != nil {
		return nil, err
	}
	grants, err := db.GetNetworkGrants(network.ID)
	if err != nil {
		return nil, err
	}

	granted := make(map[string]struct{}, len(grants))
	for _, grant := range grants {
		if grant != nil && grant.SubjectType == models.NetworkGrantSubjectUser {
			granted[grant.SubjectID] = struct{}{}
		}
	}

	candidates := make([]NetworkAccessCandidate, 0, len(users))
	for _, user := range users {
		if user == nil || user.ID == network.OwnerID {
			continue
		}
		if _, exists := granted[user.ID]; exists {
			continue
		}
		candidates = append(candidates, NetworkAccessCandidate{ID: user.ID, Username: user.Username})
	}

	return candidates, nil
}

func (s *NetworkService) GrantNetworkAccess(networkID, actorID, targetUserID, roleValue string) error {
	network, actorRole, err := s.authorizeNetwork(networkID, actorID, NetworkPermissionAccessWrite)
	if err != nil {
		return err
	}

	role, ok := ParseNetworkRole(roleValue)
	if !ok {
		return ErrNetworkGrantRoleInvalid
	}
	if !actorRole.CanAssign(role) {
		return ErrNetworkPermissionDenied
	}

	db := s.getDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	target, err := db.GetUserByID(targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	if target.ID == network.OwnerID {
		return ErrNetworkGrantTargetInvalid
	}

	existing, err := db.GetNetworkGrant(network.ID, models.NetworkGrantSubjectUser, target.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrNetworkGrantExists
	}

	now := time.Now()
	if err := db.CreateNetworkGrant(&models.NetworkGrant{
		NetworkID:   network.ID,
		SubjectType: models.NetworkGrantSubjectUser,
		SubjectID:   target.ID,
		Role:        string(role),
		GrantedBy:   actorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return err
	}

	logger.Info("service: network access granted",
		zap.String("network_id", network.ID),
		zap.String("actor_id", actorID),
		zap.String("target_user_id", target.ID),
		zap.String("role", string(role)))
	return nil
}

func (s *NetworkService) UpdateNetworkAccess(networkID, actorID, targetUserID, roleValue string) error {
	network, actorRole, err := s.authorizeNetwork(networkID, actorID, NetworkPermissionAccessWrite)
	if err != nil {
		return err
	}

	if targetUserID == actorID {
		return ErrNetworkGrantSelfChange
	}
	if targetUserID == network.OwnerID {
		return ErrNetworkGrantTargetInvalid
	}

	role, ok := ParseNetworkRole(roleValue)
	if !ok {
		return ErrNetworkGrantRoleInvalid
	}

	db := s.getDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	grant, err := db.GetNetworkGrant(network.ID, models.NetworkGrantSubjectUser, targetUserID)
	if err != nil {
		return err
	}
	if grant == nil {
		return ErrNetworkGrantNotFound
	}

	currentRole, ok := ParseNetworkRole(grant.Role)
	if !ok {
		return ErrNetworkGrantNotFound
	}
	if !actorRole.CanManage(currentRole) || !actorRole.CanAssign(role) {
		return ErrNetworkPermissionDenied
	}

	grant.Role = string(role)
	grant.GrantedBy = actorID
	grant.UpdatedAt = time.Now()
	if err := db.UpdateNetworkGrant(grant); err != nil {
		return err
	}

	logger.Info("service: network access updated",
		zap.String("network_id", network.ID),
		zap.String("actor_id", actorID),
		zap.String("target_user_id", targetUserID),
		zap.String("role", string(role)))
	return nil
}

func (s *NetworkService) RevokeNetworkAccess(networkID, actorID, targetUserID string) error {
	network, actorRole, err := s.authorizeNetwork(networkID, actorID, NetworkPermissionAccessWrite)
	if err != nil {
		return err
	}

	if targetUserID == actorID {
		return ErrNetworkGrantSelfChange
	}
	if targetUserID == network.OwnerID {
		return ErrNetworkGrantTargetInvalid
	}

	db := s.getDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	grant, err := db.GetNetworkGrant(network.ID, models.NetworkGrantSubjectUser, targetUserID)
	if err != nil {
		return err
	}
	if grant == nil {
		return ErrNetworkGrantNotFound
	}

	currentRole, ok := ParseNetworkRole(grant.Role)
	if !ok {
		return ErrNetworkGrantNotFound
	}
	if !actorRole.CanManage(currentRole) {
		return ErrNetworkPermissionDenied
	}

	if err := db.DeleteNetworkGrant(network.ID, models.NetworkGrantSubjectUser, targetUserID); err != nil {
		return err
	}

	logger.Info("service: network access revoked",
		zap.String("network_id", network.ID),
		zap.String("actor_id", actorID),
		zap.String("target_user_id", targetUserID))
	return nil
}

func (s *NetworkService) LeaveNetwork(networkID, userID string) error {
	network, err := s.getNetwork(networkID)
	if err != nil {
		return err
	}
	if network.OwnerID == userID {
		return ErrNetworkOwnerCannotLeave
	}

	db := s.getDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	grant, err := db.GetNetworkGrant(network.ID, models.NetworkGrantSubjectUser, userID)
	if err != nil {
		return err
	}
	if grant == nil {
		return ErrNetworkAccessDenied
	}

	if err := db.DeleteNetworkGrant(network.ID, models.NetworkGrantSubjectUser, userID); err != nil {
		return err
	}

	logger.Info("service: user left network", zap.String("network_id", network.ID), zap.String("user_id", userID))
	return nil
}

func (s *NetworkService) TransferNetworkOwnership(networkID, actorID, targetUserID, previousOwnerRole string) error {
	network, _, err := s.authorizeNetwork(networkID, actorID, NetworkPermissionNetworkTransfer)
	if err != nil {
		return err
	}

	if targetUserID == "" || targetUserID == actorID || targetUserID == network.OwnerID {
		return ErrNetworkOwnerTransferInvalid
	}

	db := s.getDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	target, err := db.GetUserByID(targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}

	targetGrant, err := db.GetNetworkGrant(network.ID, models.NetworkGrantSubjectUser, target.ID)
	if err != nil {
		return err
	}
	if targetGrant == nil {
		return ErrNetworkOwnerTransferInvalid
	}

	previousRole := NetworkRoleManager
	switch previousOwnerRole {
	case "":
		// Default to manager so the former owner can still help.
	case "none":
		previousRole = ""
	default:
		parsed, ok := ParseNetworkRole(previousOwnerRole)
		if !ok || parsed == NetworkRoleOwner {
			return ErrNetworkOwnerTransferInvalid
		}
		previousRole = parsed
	}

	now := time.Now()
	if err := db.WithTransaction(func(tx database.DBInterface) error {
		network.OwnerID = target.ID
		network.UpdatedAt = now
		if err := tx.UpdateNetwork(network); err != nil {
			return err
		}

		// The new owner holds the role implicitly; drop their old grant.
		if err := tx.DeleteNetworkGrant(network.ID, models.NetworkGrantSubjectUser, target.ID); err != nil {
			return err
		}

		if err := tx.DeleteNetworkGrant(network.ID, models.NetworkGrantSubjectUser, actorID); err != nil {
			return err
		}
		if previousRole == "" {
			return nil
		}
		return tx.CreateNetworkGrant(&models.NetworkGrant{
			NetworkID:   network.ID,
			SubjectType: models.NetworkGrantSubjectUser,
			SubjectID:   actorID,
			Role:        string(previousRole),
			GrantedBy:   target.ID,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}); err != nil {
		return err
	}

	logger.Info("service: network ownership transferred",
		zap.String("network_id", network.ID),
		zap.String("from_user_id", actorID),
		zap.String("to_user_id", target.ID),
		zap.String("previous_owner_role", string(previousRole)))
	return nil
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
