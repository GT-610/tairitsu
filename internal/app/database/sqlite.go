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

// SQLiteDB SQLite database implementation
type SQLiteDB struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewSQLiteDB Create a new SQLite database instance
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Ensure database directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connect to database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	return &SQLiteDB{
		db:   db,
		path: dbPath,
	}, nil
}

// Init Initialize database table structure
func (s *SQLiteDB) Init() error {
	// Create user table
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
		return fmt.Errorf("failed to create user table: %w", err)
	}

	return nil
}

// CreateUser Create user
func (s *SQLiteDB) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO users (id, username, password, email, role, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, user.ID, user.Username, user.Password, user.Email, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID Get user by ID
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
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername Get user by username
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
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail Get user by email
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
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetAllUsers Get all users
func (s *SQLiteDB) GetAllUsers() ([]*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, username, password, email, role, created_at, updated_at FROM users`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user data: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// UpdateUser Update user
func (s *SQLiteDB) UpdateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	UPDATE users 
	SET username = ?, password = ?, email = ?, role = ?, updated_at = ?
	WHERE id = ?`

	_, err := s.db.Exec(query, user.Username, user.Password, user.Email, user.Role, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser Delete user
func (s *SQLiteDB) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM users WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// HasAdminUser Check if admin user already exists
func (s *SQLiteDB) HasAdminUser() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT COUNT(*) FROM users WHERE role = 'admin'`
	row := s.db.QueryRow(query)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check admin user: %w", err)
	}

	return count > 0, nil
}

// Close Close database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}