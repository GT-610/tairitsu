package initializer

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseInitializer 数据库初始化器
type DatabaseInitializer struct {
	db database.DBInterface
}

// NewDatabaseInitializer 创建新的数据库初始化器
func NewDatabaseInitializer() *DatabaseInitializer {
	return &DatabaseInitializer{}
}

// Initialize 初始化数据库
func (di *DatabaseInitializer) Initialize() (database.DBInterface, error) {
	// 加载数据库配置
	dbConfig := database.LoadConfig()

	// 如果未配置数据库类型，跳过初始化
	if dbConfig.Type == "" {
		logger.Info("未配置数据库类型，跳过数据库初始化")
		return nil, nil
	}

	logger.Info("开始初始化数据库", zap.String("type", string(dbConfig.Type)))

	// 创建数据库实例
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("创建数据库实例失败: %w", err)
	}

	// 初始化数据库架构
	if err := db.Init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	di.db = db
	logger.Info("数据库初始化完成", zap.String("type", string(dbConfig.Type)))
	return db, nil
}

// GetDatabase 获取已初始化的数据库实例
func (di *DatabaseInitializer) GetDatabase() database.DBInterface {
	return di.db
}

// Close 关闭数据库连接
func (di *DatabaseInitializer) Close() error {
	if di.db != nil {
		return di.db.Close()
	}
	return nil
}

// ResetDatabase 重置数据库（用于测试或重新初始化）
func (di *DatabaseInitializer) ResetDatabase() error {
	if di.db != nil {
		di.Close()
	}

	dbConfig := database.LoadConfig()
	if dbConfig.Type == "" {
		return fmt.Errorf("未配置数据库类型")
	}

	return database.ResetDatabase(dbConfig)
}