package database

import (
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
)

// DBInterface defines the database interface, supporting multiple database backends
type DBInterface interface {
	// Initialize the database
	Init() error

	// Execute database operations within a transaction
	WithTransaction(fn func(DBInterface) error) error

	// User operations
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	GetUsersByIDs(ids []string) ([]*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
	CreateSession(session *models.Session) error
	GetSessionByID(id string) (*models.Session, error)
	GetSessionsByUserID(userID string) ([]*models.Session, error)
	UpdateSession(session *models.Session) error
	DeleteExpiredSessions(before time.Time) error

	// Network operations
	CreateNetwork(network *models.Network) error
	GetNetworkByID(id string) (*models.Network, error)
	GetNetworksByOwnerID(ownerID string) ([]*models.Network, error)
	GetAllNetworks() ([]*models.Network, error)
	UpdateNetwork(network *models.Network) error
	DeleteNetwork(id string) error
	UpsertNetworkViewer(viewer *models.NetworkViewer) error
	GetNetworkViewer(networkID, userID string) (*models.NetworkViewer, error)
	GetNetworkViewers(networkID string) ([]*models.NetworkViewer, error)
	GetSharedNetworksByUserID(userID string) ([]*models.Network, error)
	DeleteNetworkViewer(networkID, userID string) error
	DeleteAllNetworkViewers(networkID string) error

	// Check whether an admin user already exists
	HasAdminUser() (bool, error)

	// Check whether the database connection is alive
	Ping() error

	// Close the database connection
	Close() error
}
