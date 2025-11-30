package database

import (
	"github.com/GT-610/tairitsu/internal/app/models"
)

// DBInterface Define database interface, supporting multiple database backends
type DBInterface interface {
	// Initialize database
	Init() error
	
	// User-related operations
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
	
	// Network-related operations
	CreateNetwork(network *models.Network) error
	GetNetworkByID(id string) (*models.Network, error)
	GetNetworksByOwnerID(ownerID string) ([]*models.Network, error)
	UpdateNetwork(network *models.Network) error
	DeleteNetwork(id string) error
	
	// Check if admin user already exists
	HasAdminUser() (bool, error)
	
	// Close database connection
	Close() error
}