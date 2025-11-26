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

// LoadConfig 从统一配置管理模块加载数据库配置
func LoadConfig() Config {
	// 从统一配置管理模块获取配置
	cfg := config.AppConfig
	if cfg == nil {
		// 如果配置未初始化，返回空配置
		return Config{}
	}

	// 获取解密后的密码
	password, _ := config.GetDatabasePassword() // 忽略错误，使用空密码

	// 从配置中提取数据库相关信息
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
	logger.Info("开始重置数据库", zap.String("type", string(config.Type)))

	switch config.Type {
	case SQLite:
		// 如果没有指定路径，使用默认路径
		if config.Path == "" {
			config.Path = "tairitsu.db"
		}

		logger.Info("重置SQLite数据库", zap.String("path", config.Path))

		// 关闭现有的数据库连接
		existingDB, err := NewDatabase(config)
		if err == nil {
			logger.Info("关闭现有数据库连接")
			existingDB.Close()
		} else {
			logger.Warn("无法创建数据库连接以关闭", zap.Error(err))
		}

		// 删除SQLite数据库文件以实现重置
		err = os.Remove(config.Path)
		if err != nil && !os.IsNotExist(err) {
			logger.Error("删除SQLite数据库文件失败", zap.Error(err))
			return fmt.Errorf("重置SQLite数据库失败: %w", err)
		}

		// 如果文件不存在，记录信息但不报错
		if os.IsNotExist(err) {
			logger.Info("SQLite数据库文件不存在，将创建新文件")
		}

		logger.Info("SQLite数据库重置成功")
		return nil

	case MySQL:
		// TODO: 实现MySQL数据库重置功能
		// 注意：MySQL重置需要考虑数据备份、表结构重建等复杂操作
		logger.Warn("MySQL数据库重置功能尚未实现")
		return nil

	case PostgreSQL:
		// TODO: 实现PostgreSQL数据库重置功能
		// 注意：PostgreSQL重置需要考虑数据备份、表结构重建等复杂操作
		logger.Warn("PostgreSQL数据库重置功能尚未实现")
		return nil

	default:
		logger.Error("不支持的数据库类型", zap.String("type", string(config.Type)))
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}
}

// SaveConfig 保存数据库配置到统一配置管理模块
func SaveConfig(dbConfig Config) error {
	// 获取配置实例
	cfg := config.AppConfig
	if cfg == nil {
		return fmt.Errorf("配置未初始化")
	}

	// 更新数据库配置
	cfg.Database.Type = config.DatabaseType(dbConfig.Type)
	cfg.Database.Path = dbConfig.Path
	cfg.Database.Host = dbConfig.Host
	cfg.Database.Port = dbConfig.Port
	cfg.Database.User = dbConfig.User
	cfg.Database.Name = dbConfig.Name

	// 使用加密方式保存密码
	if err := config.SetDatabasePassword(dbConfig.Pass); err != nil {
		return fmt.Errorf("保存数据库密码失败: %w", err)
	}

	// 保存配置到文件
	return config.SaveConfig(cfg)
}
