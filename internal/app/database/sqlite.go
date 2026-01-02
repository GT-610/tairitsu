package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/GT-610/tairitsu/internal/app/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB SQLite数据库实现
type SQLiteDB struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewSQLiteDB 创建新的SQLite数据库实例
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("连接SQLite数据库失败: %w", err)
	}

	return &SQLiteDB{
		db:   db,
		path: dbPath,
	}, nil
}

// Init 初始化数据库表结构
func (s *SQLiteDB) Init() error {
	// 创建用户表
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	_, err := s.db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	// 创建网络表
	createNetworkTable := `
	CREATE TABLE IF NOT EXISTS networks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		owner_id TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	_, err = s.db.Exec(createNetworkTable)
	if err != nil {
		return fmt.Errorf("创建网络表失败: %w", err)
	}

	return nil
}

// CreateUser 创建用户
func (s *SQLiteDB) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO users (id, username, password, role, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, user.ID, user.Username, user.Password, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *SQLiteDB) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, username, password, role, created_at, updated_at FROM users WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *SQLiteDB) GetUserByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, username, password, role, created_at, updated_at FROM users WHERE username = ?`
	row := s.db.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// GetAllUsers 获取所有用户
func (s *SQLiteDB) GetAllUsers() ([]*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, username, password, role, created_at, updated_at FROM users`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询所有用户失败: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描用户数据失败: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// UpdateUser 更新用户
func (s *SQLiteDB) UpdateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	UPDATE users 
	SET username = ?, password = ?, role = ?, updated_at = ?
	WHERE id = ?`

	_, err := s.db.Exec(query, user.Username, user.Password, user.Role, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (s *SQLiteDB) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM users WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// HasAdminUser 检查是否已存在管理员用户
func (s *SQLiteDB) HasAdminUser() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT COUNT(*) FROM users WHERE role = 'admin'`
	row := s.db.QueryRow(query)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查管理员用户失败: %w", err)
	}

	return count > 0, nil
}

// CreateNetwork 创建网络
func (s *SQLiteDB) CreateNetwork(network *models.Network) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO networks (id, name, description, owner_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, network.ID, network.Name, network.Description, network.OwnerID, network.CreatedAt, network.UpdatedAt)
	if err != nil {
		return fmt.Errorf("创建网络失败: %w", err)
	}

	return nil
}

// GetNetworkByID 根据ID获取网络
func (s *SQLiteDB) GetNetworkByID(id string) (*models.Network, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, description, owner_id, created_at, updated_at FROM networks WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var network models.Network
	err := row.Scan(&network.ID, &network.Name, &network.Description, &network.OwnerID, &network.CreatedAt, &network.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询网络失败: %w", err)
	}

	return &network, nil
}

// GetNetworksByOwnerID 根据所有者ID获取网络列表
func (s *SQLiteDB) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, description, owner_id, created_at, updated_at FROM networks WHERE owner_id = ?`
	rows, err := s.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("查询网络列表失败: %w", err)
	}
	defer rows.Close()

	var networks []*models.Network
	for rows.Next() {
		var network models.Network
		err := rows.Scan(&network.ID, &network.Name, &network.Description, &network.OwnerID, &network.CreatedAt, &network.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描网络数据失败: %w", err)
		}
		networks = append(networks, &network)
	}

	return networks, nil
}

// GetAllNetworks 获取所有网络
func (s *SQLiteDB) GetAllNetworks() ([]*models.Network, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, description, owner_id, created_at, updated_at FROM networks`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询所有网络失败: %w", err)
	}
	defer rows.Close()

	var networks []*models.Network
	for rows.Next() {
		var network models.Network
		err := rows.Scan(&network.ID, &network.Name, &network.Description, &network.OwnerID, &network.CreatedAt, &network.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描网络数据失败: %w", err)
		}
		networks = append(networks, &network)
	}

	return networks, nil
}

// UpdateNetwork 更新网络
func (s *SQLiteDB) UpdateNetwork(network *models.Network) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	UPDATE networks 
	SET name = ?, description = ?, owner_id = ?, updated_at = ?
	WHERE id = ?`

	_, err := s.db.Exec(query, network.Name, network.Description, network.OwnerID, network.UpdatedAt, network.ID)
	if err != nil {
		return fmt.Errorf("更新网络失败: %w", err)
	}

	return nil
}

// DeleteNetwork 删除网络
func (s *SQLiteDB) DeleteNetwork(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM networks WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除网络失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}
