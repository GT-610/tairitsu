package database

import (
	"sync"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

var (
	globalDB   DBInterface
	globalDBMu sync.RWMutex
)

// Deprecated: global database state is retained only as a compatibility layer.
// New code should pass database dependencies explicitly via assembly/runtime services.
// SetGlobalDB sets the global database instance.
func SetGlobalDB(db DBInterface) {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()
	if globalDB != nil && globalDB != db {
		globalDB.Close()
		logger.Info("已关闭旧的数据库连接")
	}
	globalDB = db
	logger.Info("全局数据库实例已设置")
}

// Deprecated: global database state is retained only as a compatibility layer.
// New code should pass database dependencies explicitly via assembly/runtime services.
// GetGlobalDB returns the global database instance.
func GetGlobalDB() DBInterface {
	globalDBMu.RLock()
	defer globalDBMu.RUnlock()
	return globalDB
}

// Deprecated: global database state is retained only as a compatibility layer.
// New code should pass database dependencies explicitly via assembly/runtime services.
// CloseGlobalDB closes and clears the global database instance.
func CloseGlobalDB() error {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()

	if globalDB == nil {
		return nil
	}

	err := globalDB.Close()
	globalDB = nil
	logger.Info("全局数据库连接已关闭")
	return err
}

// Deprecated: global database initialization is retained only as a compatibility layer.
// New code should use AppInitializer/RuntimeService with explicit dependencies.
// InitGlobalDB initializes the global database from the currently loaded app config.
func InitGlobalDB() error {
	return InitGlobalDBFromAppConfig(config.AppConfig)
}

// Deprecated: global database initialization is retained only as a compatibility layer.
// New code should use AppInitializer/RuntimeService with explicit dependencies.
// InitGlobalDBFromAppConfig initializes the global database from an explicit app config.
func InitGlobalDBFromAppConfig(cfg *config.Config) error {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()

	dbConfig := LoadConfigFromApp(cfg)
	if dbConfig.Type == "" {
		logger.Warn("数据库配置为空，跳过全局数据库初始化")
		return nil
	}

	db, err := NewDatabase(dbConfig)
	if err != nil {
		logger.Error("创建数据库实例失败", zap.Error(err))
		return err
	}

	if err := db.Init(); err != nil {
		logger.Error("初始化数据库失败", zap.Error(err))
		db.Close()
		return err
	}

	if globalDB != nil {
		globalDB.Close()
		logger.Info("已关闭旧的数据库连接")
	}
	globalDB = db
	logger.Info("全局数据库初始化成功", zap.String("type", string(dbConfig.Type)))
	return nil
}
