package services

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

// RuntimeService coordinates mutable runtime dependencies such as the active DB and ZeroTier client.
type RuntimeService struct {
	userService    *UserService
	networkService *NetworkService
}

func NewRuntimeService(userService *UserService, networkService *NetworkService) *RuntimeService {
	return &RuntimeService{
		userService:    userService,
		networkService: networkService,
	}
}

func (s *RuntimeService) CurrentDatabase() database.DBInterface {
	if s.userService != nil && s.userService.GetDB() != nil {
		return s.userService.GetDB()
	}
	if s.networkService != nil && s.networkService.GetDB() != nil {
		return s.networkService.GetDB()
	}
	return nil
}

func (s *RuntimeService) BindDatabase(db database.DBInterface) {
	if current := s.CurrentDatabase(); current != nil && current != db {
		if err := current.Close(); err != nil {
			logger.Warn("关闭旧数据库连接失败", zap.Error(err))
		}
	}

	if s.userService != nil {
		s.userService.SetDB(db)
	}
	if s.networkService != nil {
		s.networkService.SetDB(db)
	}
}

func (s *RuntimeService) CloseCurrentDatabase() {
	current := s.CurrentDatabase()
	if current == nil {
		return
	}

	if err := current.Close(); err != nil {
		logger.Warn("关闭当前数据库连接失败", zap.Error(err))
	}

	if s.userService != nil {
		s.userService.SetDB(nil)
	}
	if s.networkService != nil {
		s.networkService.SetDB(nil)
	}
}

func (s *RuntimeService) ReopenConfiguredDatabase() error {
	dbConfig := database.LoadConfig()
	if dbConfig.Type == "" {
		return fmt.Errorf("数据库尚未配置")
	}

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("重新打开数据库失败: %w", err)
	}

	if err := db.Init(); err != nil {
		db.Close()
		return fmt.Errorf("重新初始化数据库失败: %w", err)
	}

	s.BindDatabase(db)
	return nil
}

func (s *RuntimeService) BindZTClient(client *zerotier.Client) {
	if s.networkService == nil {
		return
	}
	s.networkService.SetZTClient(client)
}

func (s *RuntimeService) InitZTClientFromConfig() (*zerotier.Status, error) {
	ztClient, err := zerotier.NewClient()
	if err != nil {
		return nil, fmt.Errorf("创建ZeroTier客户端失败: %w", err)
	}

	s.BindZTClient(ztClient)

	status, err := s.networkService.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("ZeroTier客户端初始化后验证失败: %w", err)
	}

	return status, nil
}
