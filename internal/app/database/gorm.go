package database

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormDB 是基于GORM的数据库实现
type GormDB struct {
	db *gorm.DB
}

// Init 初始化数据库
func (g *GormDB) Init() error {
	// 自动迁移用户模型
	if err := g.db.AutoMigrate(&models.User{}, &models.Network{}, &models.Session{}, &models.NetworkViewer{}); err != nil {
		return fmt.Errorf("自动迁移模型失败: %w", err)
	}
	return nil
}

// WithTransaction 在事务中执行数据库操作
func (g *GormDB) WithTransaction(fn func(DBInterface) error) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		return fn(&GormDB{db: tx})
	})
}

// CreateUser 创建用户
func (g *GormDB) CreateUser(user *models.User) error {
	result := g.db.Create(user)
	return result.Error
}

// GetUserByID 根据ID获取用户
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

// GetUserByUsername 根据用户名获取用户
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

// GetAllUsers 获取所有用户
func (g *GormDB) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	result := g.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUser 更新用户
func (g *GormDB) UpdateUser(user *models.User) error {
	result := g.db.Save(user)
	return result.Error
}

// DeleteUser 删除用户
func (g *GormDB) DeleteUser(id string) error {
	result := g.db.Delete(&models.User{}, "id = ?", id)
	return result.Error
}

// CreateSession 创建会话
func (g *GormDB) CreateSession(session *models.Session) error {
	result := g.db.Create(session)
	return result.Error
}

// GetSessionByID 根据ID获取会话
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

// GetSessionsByUserID 获取用户会话列表
func (g *GormDB) GetSessionsByUserID(userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	result := g.db.Where("user_id = ?", userID).Order("last_seen_at desc").Find(&sessions)
	if result.Error != nil {
		return nil, result.Error
	}
	return sessions, nil
}

// UpdateSession 更新会话
func (g *GormDB) UpdateSession(session *models.Session) error {
	result := g.db.Save(session)
	return result.Error
}

// HasAdminUser 检查是否已存在管理员用户
func (g *GormDB) HasAdminUser() (bool, error) {
	var count int64
	result := g.db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CreateNetwork 创建网络
func (g *GormDB) CreateNetwork(network *models.Network) error {
	result := g.db.Create(network)
	return result.Error
}

// GetNetworkByID 根据ID获取网络
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

// GetNetworksByOwnerID 根据所有者ID获取网络列表
func (g *GormDB) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	var networks []*models.Network
	result := g.db.Where("owner_id = ?", ownerID).Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

// GetAllNetworks 获取所有网络
func (g *GormDB) GetAllNetworks() ([]*models.Network, error) {
	var networks []*models.Network
	result := g.db.Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

// UpdateNetwork 更新网络
func (g *GormDB) UpdateNetwork(network *models.Network) error {
	result := g.db.Save(network)
	return result.Error
}

// DeleteNetwork 删除网络
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
		Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

func (g *GormDB) DeleteNetworkViewer(networkID, userID string) error {
	return g.db.Delete(&models.NetworkViewer{}, "network_id = ? AND user_id = ?", networkID, userID).Error
}

// Close 关闭数据库连接
func (g *GormDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
