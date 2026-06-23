package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
)

// Config holds the database configuration
type Config struct {
	Type DatabaseType
	Path string // SQLite database path
	Host string // PostgreSQL/MySQL host
	Port int    // PostgreSQL/MySQL port
	User string // PostgreSQL/MySQL user
	Pass string // PostgreSQL/MySQL password
	Name string // PostgreSQL/MySQL database name
}

// NewDatabase creates a database instance based on the given configuration
func NewDatabase(config Config) (DBInterface, error) {
	switch config.Type {
	case SQLite:
		// Use default path if none is specified
		if config.Path == "" {
			config.Path = "data/tairitsu.db"
		}

		if err := ensureSQLiteDir(config.Path); err != nil {
			return nil, err
		}

		db, err := gorm.Open(sqlite.Open(config.Path+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get SQLite database instance: %w", err)
		}

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)

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

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)

		return &GormDB{db: db}, nil

	case PostgreSQL:
		sslMode := "require"
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			config.Host, config.User, config.Pass, config.Name, config.Port, sslMode)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get PostgreSQL database instance: %w", err)
		}

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)

		return &GormDB{db: db}, nil

	default:
		// Return error if database type is not specified
		if config.Type == "" {
			return nil, fmt.Errorf("database type is required")
		}
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

func LoadConfigFromApp(cfg *config.Config) Config {
	if cfg == nil {
		return Config{}
	}

	password, _ := config.GetDatabasePasswordFrom(cfg)

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

// ResetDatabase is the generic handler for resetting a database.
// WARNING: This operation will delete all data in the database. Use with caution.
func ResetDatabase(config Config) error {
	logger.Info("starting database reset", zap.String("type", string(config.Type)))

	switch config.Type {
	case SQLite:
		// Use default path if none is specified
		if config.Path == "" {
			config.Path = "data/tairitsu.db"
		}

		logger.Info("resetting SQLite database", zap.String("path", config.Path))

		// Delete the SQLite database file to reset it
		err := os.Remove(config.Path)
		if err != nil && !os.IsNotExist(err) {
			logger.Error("failed to delete SQLite database file", zap.Error(err))
			return fmt.Errorf("failed to reset SQLite database: %w", err)
		}

		// Log info if file does not exist, but don't treat it as an error
		if os.IsNotExist(err) {
			logger.Info("SQLite database file does not exist; a new file will be created")
		}

		logger.Info("SQLite database reset successfully")
		return nil

	case MySQL:
		logger.Warn("MySQL database reset is not currently supported")
		return fmt.Errorf("only SQLite is currently supported; MySQL reset is not supported")

	case PostgreSQL:
		logger.Warn("PostgreSQL database reset is not currently supported")
		return fmt.Errorf("only SQLite is currently supported; PostgreSQL reset is not supported")

	default:
		logger.Error("unsupported database type", zap.String("type", string(config.Type)))
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

func ensureSQLiteDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create SQLite database directory: %w", err)
	}

	return nil
}

func SaveConfigToApp(cfg *config.Config, dbConfig Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not initialized")
	}

	// Update database configuration
	cfg.Database.Type = string(dbConfig.Type)
	cfg.Database.Path = dbConfig.Path
	cfg.Database.Host = dbConfig.Host
	cfg.Database.Port = dbConfig.Port
	cfg.Database.User = dbConfig.User
	cfg.Database.Name = dbConfig.Name

	if err := config.SetDatabasePasswordOn(cfg, dbConfig.Pass); err != nil {
		return fmt.Errorf("failed to save database password: %w", err)
	}

	return config.SaveConfig(cfg)
}
