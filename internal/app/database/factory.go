package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseType 数据库类型
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
)

// Config 数据库配置
type Config struct {
	Type DatabaseType
	Path string // SQLite数据库路径
	Host string // PostgreSQL/MySQL主机
	Port int    // PostgreSQL/MySQL端口
	User string // PostgreSQL/MySQL用户
	Pass string // PostgreSQL/MySQL密码
	Name string // PostgreSQL/MySQL数据库名
}

// NewDatabase 根据配置创建数据库实例
func NewDatabase(config Config) (DBInterface, error) {
	switch config.Type {
	case SQLite:
		// 如果没有指定路径，使用默认路径
		if config.Path == "" {
			config.Path = "data/tairitsu.db"
		}

		if err := ensureSQLiteDir(config.Path); err != nil {
			return nil, err
		}

		db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get SQLite database instance: %w", err)
		}

		// 设置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	case MySQL:
		// 构建DSN
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

		// 设置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	case PostgreSQL:
		// 构建DSN
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

		// 设置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	default:
		// 如果没有指定数据库类型，返回错误
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

// ResetDatabase 重置数据库的通用处理函数
// 注意：此操作将清空数据库中的所有数据，请谨慎使用
func ResetDatabase(config Config) error {
	logger.Info("starting database reset", zap.String("type", string(config.Type)))

	switch config.Type {
	case SQLite:
		// 如果没有指定路径，使用默认路径
		if config.Path == "" {
			config.Path = "data/tairitsu.db"
		}

		logger.Info("resetting SQLite database", zap.String("path", config.Path))

		// 删除SQLite数据库文件以实现重置
		err := os.Remove(config.Path)
		if err != nil && !os.IsNotExist(err) {
			logger.Error("failed to delete SQLite database file", zap.Error(err))
			return fmt.Errorf("failed to reset SQLite database: %w", err)
		}

		// 如果文件不存在，记录信息但不报错
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

	// 更新数据库配置
	cfg.Database.Type = config.DatabaseType(dbConfig.Type)
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
