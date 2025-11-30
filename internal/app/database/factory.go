package database

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseType Database type
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
)

// Config Database configuration
type Config struct {
	Type DatabaseType
	Path string // SQLite database path
	Host string // PostgreSQL / MySQL host
	Port int    // PostgreSQL / MySQL port
	User string // PostgreSQL / MySQL user
	Pass string // PostgreSQL / MySQL password
	Name string // PostgreSQL / MySQL database name
}

// NewDatabase Create database instance based on configuration
func NewDatabase(config Config) (DBInterface, error) {
	switch config.Type {
	case SQLite:
		// If no path specified, use default path
		if config.Path == "" {
			config.Path = "data/tairitsu.db"
		}

		db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get SQLite database instance: %w", err)
		}

		// Set connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	case MySQL:
		// Build DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Pass, config.Host, config.Port, config.Name)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get MySQL database instance: %w", err)
		}

		// Set connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	case PostgreSQL:
		// Build DSN
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host, config.User, config.Pass, config.Name, config.Port)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get PostgreSQL database instance: %w", err)
		}

		// Set connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	default:
		// If no database type specified, return error
		if config.Type == "" {
			return nil, fmt.Errorf("database type must be specified")
		}
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// LoadConfig Load database configuration from unified configuration management module
func LoadConfig() Config {
	// Get configuration from unified configuration management module
	cfg := config.AppConfig
	if cfg == nil {
		// If configuration not initialized, return empty configuration
		return Config{}
	}

	// Get decrypted password
	password, _ := config.GetDatabasePassword() // Ignore error, use empty password

	// Extract database-related information from configuration
	return Config{
		Type: DatabaseType(cfg.Database.Type),
		Path: cfg.Database.Path,
		Host: cfg.Database.Host,
		Port: cfg.Database.Port,
		User: cfg.Database.User,
		Pass: password,
		Name: cfg.Database.Name,
	}
}

// ResetDatabase Universal database reset function
// Note: This operation will clear all data in the database, use with caution
func ResetDatabase(config Config) error {
	logger.Info("Starting database reset", zap.String("type", string(config.Type)))

	switch config.Type {
	case SQLite:
		// If no path specified, use default path
		if config.Path == "" {
			config.Path = "data/tairitsu.db"
		}

		logger.Info("Resetting SQLite database", zap.String("path", config.Path))

		// Close existing database connection
		existingDB, err := NewDatabase(config)
		if err == nil {
			logger.Info("Closing existing database connection")
			existingDB.Close()
		} else {
			logger.Warn("Failed to create database connection to close", zap.Error(err))
		}

		// Delete SQLite database file to reset
		err = os.Remove(config.Path)
		if err != nil && !os.IsNotExist(err) {
			logger.Error("Failed to delete SQLite database file", zap.Error(err))
			return fmt.Errorf("failed to reset SQLite database: %w", err)
		}

		// If file doesn't exist, log info but don't error
		if os.IsNotExist(err) {
			logger.Info("SQLite database file doesn't exist, will create new file")
		}

		logger.Info("SQLite database reset completed successfully")
		return nil

	case MySQL:
		// TODO: Implement MySQL database reset functionality
		// Note: MySQL reset requires consideration of data backup, table structure rebuilding, etc.
		logger.Warn("MySQL database reset functionality not yet implemented")
		return nil

	case PostgreSQL:
		// TODO: Implement PostgreSQL database reset functionality
		// Note: PostgreSQL reset requires consideration of data backup, table structure rebuilding, etc.
		logger.Warn("PostgreSQL database reset functionality not yet implemented")
		return nil

	default:
		logger.Error("Unsupported database type", zap.String("type", string(config.Type)))
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// SaveConfig Save database configuration to unified configuration management module
func SaveConfig(dbConfig Config) error {
	// Get configuration instance
	cfg := config.AppConfig
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Update database configuration
	cfg.Database.Type = config.DatabaseType(dbConfig.Type)
	cfg.Database.Path = dbConfig.Path
	cfg.Database.Host = dbConfig.Host
	cfg.Database.Port = dbConfig.Port
	cfg.Database.User = dbConfig.User
	cfg.Database.Name = dbConfig.Name

	// Save password using encryption
	if err := config.SetDatabasePassword(dbConfig.Pass); err != nil {
		return fmt.Errorf("failed to save database password: %w", err)
	}

	// Save configuration to file
	return config.SaveConfig(cfg)
}
