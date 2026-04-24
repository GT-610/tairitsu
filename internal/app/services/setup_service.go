package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"go.uber.org/zap"
)

type SetupService struct {
	runtimeService *RuntimeService
	stateService   *StateService
	userService    *UserService
	networkService *NetworkService
}

func NewSetupService(runtimeService *RuntimeService, stateService *StateService, userService *UserService, networkService *NetworkService) *SetupService {
	return &SetupService{
		runtimeService: runtimeService,
		stateService:   stateService,
		userService:    userService,
		networkService: networkService,
	}
}

func (s *SetupService) ConfigureDatabase(dbConfig models.DatabaseConfig) (database.Config, error) {
	if dbConfig.Type != string(database.SQLite) {
		return database.Config{}, fmt.Errorf("当前仅支持 SQLite，请使用 SQLite 完成初始化")
	}

	dbCfg := database.Config{
		Type: database.DatabaseType(dbConfig.Type),
		Path: dbConfig.Path,
		Host: dbConfig.Host,
		Port: dbConfig.Port,
		User: dbConfig.User,
		Pass: dbConfig.Pass,
		Name: dbConfig.Name,
	}

	db, err := database.NewDatabase(dbCfg)
	if err != nil {
		return database.Config{}, fmt.Errorf("数据库连接失败: %w", err)
	}

	if err := db.Init(); err != nil {
		db.Close()
		return database.Config{}, fmt.Errorf("数据库初始化失败: %w", err)
	}

	if dbCfg.Type == database.SQLite && dbCfg.Path == "" {
		dbCfg.Path = "data/tairitsu.db"
	}

	if err := s.stateService.SaveDatabaseConfig(dbCfg); err != nil {
		db.Close()
		return database.Config{}, fmt.Errorf("保存数据库配置失败: %w", err)
	}

	s.runtimeService.BindDatabase(db)
	return dbCfg, nil
}

func (s *SetupService) SaveZeroTierConfig(controllerURL, tokenPath string) (*zerotier.Status, error) {
	if err := s.stateService.SaveZeroTierConfig(controllerURL, tokenPath); err != nil {
		return nil, fmt.Errorf("保存ZeroTier配置失败: %w", err)
	}

	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("创建ZeroTier客户端失败: %w", err)
	}

	status, err := ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("ZeroTier客户端验证失败: %w", err)
	}

	s.runtimeService.BindZTClient(ztClient)
	return status, nil
}

func (s *SetupService) TestZeroTierConnection() (*zerotier.Status, error) {
	ztClient, err := s.stateService.CreateZTClient()
	if err != nil {
		return nil, fmt.Errorf("创建ZeroTier客户端失败: %w", err)
	}

	status, err := ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("无法连接到ZeroTier控制器: %w", err)
	}

	return status, nil
}

func (s *SetupService) InitZTClientFromConfig() (*zerotier.Status, error) {
	return s.runtimeService.InitZTClientFromConfig()
}

func (s *SetupService) GetSetupStatus() SetupStatus {
	return s.stateService.GetSetupStatus(s.userService, s.networkService)
}

func (s *SetupService) GetRuntimeSettings() RuntimeSettings {
	return s.stateService.RuntimeSettings()
}

func (s *SetupService) UpdateRuntimeSettings(settings RuntimeSettings) error {
	return s.stateService.SaveRuntimeSettings(settings)
}

func (s *SetupService) InitializeAdminCreation() (string, error) {
	dbConfig := s.stateService.DatabaseConfig()
	if dbConfig.Type == "" {
		return "", fmt.Errorf("尚未完成数据库配置，请先配置 SQLite 数据库")
	}
	if dbConfig.Type != database.SQLite {
		return "", fmt.Errorf("当前仅支持 SQLite，%s 初始化暂不支持", dbConfig.Type)
	}

	if s.stateService.IsInitialized() {
		return "", fmt.Errorf("系统已初始化，不能再次执行首次管理员创建准备")
	}

	resetDoneKey := "admin_creation_reset_done"
	if config.GetTempSetting(resetDoneKey) == "true" {
		return string(dbConfig.Type), nil
	}

	hasAdmin, err := s.userService.HasAdminUser()
	if err == nil && hasAdmin {
		config.SetTempSetting(resetDoneKey, "true")
		return string(dbConfig.Type), nil
	}

	s.runtimeService.CloseCurrentDatabase()

	if err := database.ResetDatabase(dbConfig); err != nil {
		return "", fmt.Errorf("初始化管理员账户创建步骤失败: %w", err)
	}

	if err := s.runtimeService.ReopenConfiguredDatabase(); err != nil {
		return "", fmt.Errorf("数据库重置后重新初始化失败: %w", err)
	}

	config.SetTempSetting(resetDoneKey, "true")
	return string(dbConfig.Type), nil
}

func (s *SetupService) SetInitialized(initialized bool) error {
	if initialized {
		if err := s.validateInitializationReady(); err != nil {
			return err
		}

		cfg := s.stateService.Config()
		if cfg.Security.JWTSecret == "" {
			cfg.Security.JWTSecret = generateRandomSecret(32)
			logger.Info("生成新的JWT密钥")
		}

		if cfg.Security.SessionSecret == "" {
			cfg.Security.SessionSecret = generateRandomSecret(32)
			logger.Info("生成新的会话密钥")
		}

		if err := s.stateService.SaveConfig(); err != nil {
			return fmt.Errorf("生成安全密钥失败: %w", err)
		}
	}

	if err := s.stateService.SetInitialized(initialized); err != nil {
		return fmt.Errorf("设置初始化状态失败: %w", err)
	}

	return nil
}

func (s *SetupService) validateInitializationReady() error {
	if !s.stateService.DatabaseConfigured() {
		return fmt.Errorf("尚未完成数据库配置，请先配置 SQLite 数据库")
	}

	if s.runtimeService.CurrentDatabase() == nil {
		if err := s.runtimeService.ReopenConfiguredDatabase(); err != nil {
			return fmt.Errorf("数据库未就绪: %w", err)
		}
	}

	hasAdmin, err := s.userService.HasAdminUser()
	if err != nil {
		return fmt.Errorf("无法确认管理员状态: %w", err)
	}
	if !hasAdmin {
		return fmt.Errorf("请先创建首个管理员账户")
	}

	if _, err := s.stateService.CreateZTClient(); err != nil {
		return fmt.Errorf("ZeroTier 配置未完成: %w", err)
	}

	status, err := s.runtimeService.InitZTClientFromConfig()
	if err != nil {
		return fmt.Errorf("ZeroTier 连接验证失败: %w", err)
	}
	if status == nil || !status.Online {
		return fmt.Errorf("ZeroTier 控制器当前不可用")
	}

	return nil
}

func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		logger.Warn("使用crypto/rand生成随机密钥失败，将使用math/rand作为备选方案", zap.Error(err))

		r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		for i := range bytes {
			bytes[i] = byte(r.Intn(256))
		}
	}

	return base64.URLEncoding.EncodeToString(bytes)
}
