package database

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/models"
	"gorm.io/gorm"
)

// GormDB 是基于GORM的数据库实现
type GormDB struct {
	db *gorm.DB
}

// Init 初始化数据库
func (g *GormDB) Init() error {
	// 自动迁移用户模型
	if err := g.db.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf("自动迁移用户模型失败: %w", err)
	}
	return nil
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

// GetUserByEmail 根据邮箱获取用户
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

// HasAdminUser 检查是否已存在管理员用户
func (g *GormDB) HasAdminUser() (bool, error) {
	var count int64
	result := g.db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// Close 关闭数据库连接
func (g *GormDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}