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
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	return nil
}

// CreateUser 创建用户
func (s *SQLiteDB) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO users (id, username, password, email, role, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, user.ID, user.Username, user.Password, user.Email, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *SQLiteDB) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, username, password, email, role, created_at, updated_at FROM users WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
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

	query := `SELECT id, username, password, email, role, created_at, updated_at FROM users WHERE username = ?`
	row := s.db.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *SQLiteDB) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, username, password, email, role, created_at, updated_at FROM users WHERE email = ?`
	row := s.db.QueryRow(query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
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

	query := `SELECT id, username, password, email, role, created_at, updated_at FROM users`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询所有用户失败: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
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
	SET username = ?, password = ?, email = ?, role = ?, updated_at = ?
	WHERE id = ?`

	_, err := s.db.Exec(query, user.Username, user.Password, user.Email, user.Role, user.UpdatedAt, user.ID)
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

// Close 关闭数据库连接
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}