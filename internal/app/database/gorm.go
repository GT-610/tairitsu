package database

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/models"
	"gorm.io/gorm"
)

// GormDB is a GORM-based database implementation
type GormDB struct {
	db *gorm.DB
}

// Init Initialize database
func (g *GormDB) Init() error {
	// Auto migrate user models
	if err := g.db.AutoMigrate(&models.User{}, &models.Network{}); err != nil {
		return fmt.Errorf("自动迁移模型失败: %w", err)
	}
	return nil
}

// CreateUser Create user
func (g *GormDB) CreateUser(user *models.User) error {
	result := g.db.Create(user)
	return result.Error
}

// GetUserByID Get user by ID
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

// GetUserByUsername Get user by username
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

// GetUserByEmail Get user by email
func (g *GormDB) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := g.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetAllUsers Get all users
func (g *GormDB) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	result := g.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUser Update user
func (g *GormDB) UpdateUser(user *models.User) error {
	result := g.db.Save(user)
	return result.Error
}

// DeleteUser Delete user
func (g *GormDB) DeleteUser(id string) error {
	result := g.db.Delete(&models.User{}, "id = ?", id)
	return result.Error
}

// HasAdminUser Check if admin user already exists
func (g *GormDB) HasAdminUser() (bool, error) {
	var count int64
	result := g.db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CreateNetwork Create network
func (g *GormDB) CreateNetwork(network *models.Network) error {
	result := g.db.Create(network)
	return result.Error
}

// GetNetworkByID Get network by ID
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

// GetNetworksByOwnerID Get networks by owner ID
func (g *GormDB) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	var networks []*models.Network
	result := g.db.Where("owner_id = ?", ownerID).Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

// UpdateNetwork Update network
func (g *GormDB) UpdateNetwork(network *models.Network) error {
	result := g.db.Save(network)
	return result.Error
}

// DeleteNetwork Delete network
func (g *GormDB) DeleteNetwork(id string) error {
	result := g.db.Delete(&models.Network{}, "id = ?", id)
	return result.Error
}

// Close Close database connection
func (g *GormDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}