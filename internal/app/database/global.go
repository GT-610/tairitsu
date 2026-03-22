package database

import (
	"sync"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

var (
	globalDB   DBInterface
	globalDBMu sync.RWMutex
)

// SetGlobalDB sets the global database instance
func SetGlobalDB(db DBInterface) {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()
	globalDB = db
	logger.Info("全局数据库实例已设置")
}

// GetGlobalDB returns the global database instance
func GetGlobalDB() DBInterface {
	globalDBMu.RLock()
	defer globalDBMu.RUnlock()
	return globalDB
}

// InitGlobalDB initializes the global database from config
func InitGlobalDB() error {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()

	config := LoadConfig()
	if config.Type == "" {
		logger.Warn("数据库配置为空，跳过全局数据库初始化")
		return nil
	}

	db, err := NewDatabase(config)
	if err != nil {
		logger.Error("创建数据库实例失败", zap.Error(err))
		return err
	}

	if err := db.Init(); err != nil {
		logger.Error("初始化数据库失败", zap.Error(err))
		return err
	}

	globalDB = db
	logger.Info("全局数据库初始化成功", zap.String("type", string(config.Type)))
	return nil
}
