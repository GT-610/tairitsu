package database

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
			config.Path = "tairitsu.db"
		}

		db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("无法连接到SQLite数据库: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("无法获取SQLite数据库实例: %w", err)
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
			return nil, fmt.Errorf("无法连接到MySQL数据库: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("无法获取MySQL数据库实例: %w", err)
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
			return nil, fmt.Errorf("无法连接到PostgreSQL数据库: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("无法获取PostgreSQL数据库实例: %w", err)
		}

		// 设置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		return &GormDB{db: db}, nil

	default:
		// 如果没有指定数据库类型，默认使用内存数据库（用于开发环境）
		if config.Type == "" {
			log.Println("警告: 未指定数据库类型，使用内存数据库")
			return NewMemoryDB(), nil
		}
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}
}

// LoadConfigFromEnv 从环境变量加载数据库配置
func LoadConfigFromEnv() Config {
	config := Config{
		Type: DatabaseType(os.Getenv("DATABASE_TYPE")),
		Path: os.Getenv("DATABASE_PATH"),
		Host: os.Getenv("DATABASE_HOST"),
		User: os.Getenv("DATABASE_USER"),
		Pass: os.Getenv("DATABASE_PASS"),
		Name: os.Getenv("DATABASE_NAME"),
	}

	// 解析端口号
	if portStr := os.Getenv("DATABASE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	return config
}