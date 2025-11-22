package database

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tairitsu/tairitsu/internal/app/models"
)

// MemoryDB 内存数据库实现（临时使用，最终会被移除）
type MemoryDB struct {
	users           map[string]*models.User
	usersByUsername map[string]*models.User
	usersByEmail    map[string]*models.User
	mutex           sync.RWMutex
}

// NewMemoryDB 创建新的内存数据库实例
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		users:           make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
		usersByEmail:    make(map[string]*models.User),
	}
}

// Init 初始化数据库
func (m *MemoryDB) Init() error {
	// 内存数据库不需要初始化
	return nil
}

// CreateUser 创建用户
func (m *MemoryDB) CreateUser(user *models.User) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查用户名是否已存在
	if _, exists := m.usersByUsername[user.Username]; exists {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, exists := m.usersByEmail[user.Email]; exists {
		return errors.New("邮箱已被使用")
	}

	// 如果ID为空，生成一个新的UUID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// 设置创建和更新时间
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	// 保存用户
	m.users[user.ID] = user
	m.usersByUsername[user.Username] = user
	m.usersByEmail[user.Email] = user

	return nil
}

// GetUserByID 根据ID获取用户
func (m *MemoryDB) GetUserByID(id string) (*models.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, nil
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (m *MemoryDB) GetUserByUsername(username string) (*models.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user, exists := m.usersByUsername[username]
	if !exists {
		return nil, nil
	}

	return user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (m *MemoryDB) GetUserByEmail(email string) (*models.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, nil
	}

	return user, nil
}

// GetAllUsers 获取所有用户
func (m *MemoryDB) GetAllUsers() ([]*models.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser 更新用户
func (m *MemoryDB) UpdateUser(user *models.User) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查用户是否存在
	existingUser, exists := m.users[user.ID]
	if !exists {
		return errors.New("用户不存在")
	}

	// 更新用户信息
	existingUser.Username = user.Username
	existingUser.Password = user.Password
	existingUser.Email = user.Email
	existingUser.Role = user.Role
	existingUser.UpdatedAt = time.Now()

	// 更新索引
	m.usersByUsername[user.Username] = existingUser
	m.usersByEmail[user.Email] = existingUser

	return nil
}

// DeleteUser 删除用户
func (m *MemoryDB) DeleteUser(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	user, exists := m.users[id]
	if !exists {
		return errors.New("用户不存在")
	}

	// 删除索引
	delete(m.usersByUsername, user.Username)
	delete(m.usersByEmail, user.Email)
	delete(m.users, id)

	return nil
}

// HasAdminUser 检查是否已存在管理员用户
func (m *MemoryDB) HasAdminUser() (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, user := range m.users {
		if user.Role == "admin" {
			return true, nil
		}
	}

	return false, nil
}

// Close 关闭数据库连接
func (m *MemoryDB) Close() error {
	// 内存数据库不需要关闭连接
	return nil
}