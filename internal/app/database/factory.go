package database

import (
	"encoding/json"
	"fmt"
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
		// 如果没有指定数据库类型，返回错误
		if config.Type == "" {
			return nil, fmt.Errorf("必须指定数据库类型")
		}
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}
}

// LoadConfig 从JSON配置文件加载数据库配置
func LoadConfig() Config {
	// 从JSON配置文件读取
	config, err := loadConfigFromJSON()
	if err != nil {
		// 配置文件不存在或读取失败时返回空配置
		return Config{}
	}
	return config
}

// loadConfigFromJSON 从JSON文件加载数据库配置
func loadConfigFromJSON() (Config, error) {
	config := Config{}

	// 读取数据库配置文件
	data, err := os.ReadFile("./database_config.json")
	if err != nil {
		return config, fmt.Errorf("读取数据库配置文件失败: %w", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("解析数据库配置失败: %w", err)
	}

	return config, nil
}

// SaveConfigToJSON 将数据库配置保存到JSON文件
// 注意：此函数会强制覆盖现有的配置文件，不进行文件存在性检查
func SaveConfigToJSON(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化数据库配置失败: %w", err)
	}

	// 强制覆盖模式：os.WriteFile默认会覆盖文件，无需额外检查
	// 确保文件权限为0644（所有者可读写，其他用户可读）
	return os.WriteFile("./database_config.json", data, 0644)
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
