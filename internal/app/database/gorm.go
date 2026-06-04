package database

import (
	"fmt"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormDB is the GORM-based database implementation
type GormDB struct {
	db *gorm.DB
}

// Init initializes the database
func (g *GormDB) Init() error {
	// Auto-migrate user models
	if err := g.db.AutoMigrate(&models.User{}, &models.Network{}, &models.Session{}, &models.NetworkViewer{}); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}
	return nil
}

// WithTransaction executes database operations within a transaction
func (g *GormDB) WithTransaction(fn func(DBInterface) error) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		return fn(&GormDB{db: tx})
	})
}

// CreateUser creates a new user
func (g *GormDB) CreateUser(user *models.User) error {
	result := g.db.Create(user)
	return result.Error
}

// GetUserByID retrieves a user by ID
func (g *GormDB) GetUserByID(id string) (*models.User, error) {
	var user models.User
	result := g.db.First(&user, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (g *GormDB) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := g.db.First(&user, "username = ?", username)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetAllUsers retrieves all users
func (g *GormDB) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	result := g.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetUsersByIDs retrieves users by a list of IDs in a single query
func (g *GormDB) GetUsersByIDs(ids []string) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}
	var users []*models.User
	result := g.db.Where("id IN ?", ids).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUser updates a user
func (g *GormDB) UpdateUser(user *models.User) error {
	result := g.db.Save(user)
	return result.Error
}

// DeleteUser deletes a user
func (g *GormDB) DeleteUser(id string) error {
	result := g.db.Delete(&models.User{}, "id = ?", id)
	return result.Error
}

// CreateSession creates a new session
func (g *GormDB) CreateSession(session *models.Session) error {
	result := g.db.Create(session)
	return result.Error
}

// GetSessionByID retrieves a session by ID
func (g *GormDB) GetSessionByID(id string) (*models.Session, error) {
	var session models.Session
	result := g.db.First(&session, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &session, nil
}

// GetSessionsByUserID retrieves all sessions for a user
func (g *GormDB) GetSessionsByUserID(userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	result := g.db.Where("user_id = ?", userID).Order("last_seen_at desc").Find(&sessions)
	if result.Error != nil {
		return nil, result.Error
	}
	return sessions, nil
}

// UpdateSession updates a session
func (g *GormDB) UpdateSession(session *models.Session) error {
	result := g.db.Save(session)
	return result.Error
}

func (g *GormDB) DeleteExpiredSessions(before time.Time) error {
	result := g.db.Where("(revoked_at IS NOT NULL AND revoked_at < ?) OR (expires_at < ?)", before, before).Delete(&models.Session{})
	return result.Error
}

// HasAdminUser checks whether an admin user already exists
func (g *GormDB) HasAdminUser() (bool, error) {
	var count int64
	result := g.db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CreateNetwork creates a new network
func (g *GormDB) CreateNetwork(network *models.Network) error {
	result := g.db.Create(network)
	return result.Error
}

// GetNetworkByID retrieves a network by ID
func (g *GormDB) GetNetworkByID(id string) (*models.Network, error) {
	var network models.Network
	result := g.db.First(&network, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &network, nil
}

// GetNetworksByOwnerID retrieves all networks owned by the given user
func (g *GormDB) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	var networks []*models.Network
	result := g.db.Where("owner_id = ?", ownerID).Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

// GetAllNetworks retrieves all networks
func (g *GormDB) GetAllNetworks() ([]*models.Network, error) {
	var networks []*models.Network
	result := g.db.Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

// UpdateNetwork updates a network
func (g *GormDB) UpdateNetwork(network *models.Network) error {
	result := g.db.Save(network)
	return result.Error
}

// DeleteNetwork deletes a network
func (g *GormDB) DeleteNetwork(id string) error {
	result := g.db.Delete(&models.Network{}, "id = ?", id)
	return result.Error
}

func (g *GormDB) UpsertNetworkViewer(viewer *models.NetworkViewer) error {
	return g.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "network_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"granted_by", "updated_at"}),
	}).Create(viewer).Error
}

func (g *GormDB) GetNetworkViewer(networkID, userID string) (*models.NetworkViewer, error) {
	var viewer models.NetworkViewer
	result := g.db.First(&viewer, "network_id = ? AND user_id = ?", networkID, userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &viewer, nil
}

func (g *GormDB) GetNetworkViewers(networkID string) ([]*models.NetworkViewer, error) {
	var viewers []*models.NetworkViewer
	result := g.db.Where("network_id = ?", networkID).Order("created_at ASC").Find(&viewers)
	if result.Error != nil {
		return nil, result.Error
	}
	return viewers, nil
}

func (g *GormDB) GetSharedNetworksByUserID(userID string) ([]*models.Network, error) {
	var networks []*models.Network
	result := g.db.Table("networks").
		Select("networks.id, networks.name, networks.description, networks.owner_id, networks.created_at, networks.updated_at").
		Joins("JOIN network_viewers ON network_viewers.network_id = networks.id").
		Where("network_viewers.user_id = ?", userID).
		Order("networks.created_at DESC").
		Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

func (g *GormDB) DeleteNetworkViewer(networkID, userID string) error {
	return g.db.Delete(&models.NetworkViewer{}, "network_id = ? AND user_id = ?", networkID, userID).Error
}

func (g *GormDB) DeleteAllNetworkViewers(networkID string) error {
	return g.db.Delete(&models.NetworkViewer{}, "network_id = ?", networkID).Error
}

// Ping checks if the database connection is alive.
func (g *GormDB) Ping() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close closes the database connection
func (g *GormDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
