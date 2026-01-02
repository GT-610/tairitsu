package database

import (
	"github.com/GT-610/tairitsu/internal/app/models"
)

// DBInterface 定义数据库接口，支持多种数据库后端
type DBInterface interface {
	// 初始化数据库
	Init() error

	// 用户相关操作
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error

	// 网络相关操作
	CreateNetwork(network *models.Network) error
	GetNetworkByID(id string) (*models.Network, error)
	GetNetworksByOwnerID(ownerID string) ([]*models.Network, error)
	GetAllNetworks() ([]*models.Network, error)
	UpdateNetwork(network *models.Network) error
	DeleteNetwork(id string) error

	// 检查是否已存在管理员用户
	HasAdminUser() (bool, error)

	// 关闭数据库连接
	Close() error
}
