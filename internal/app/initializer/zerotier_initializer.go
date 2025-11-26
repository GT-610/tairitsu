package initializer

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
)

// ZeroTierInitializer ZeroTier客户端初始化器
type ZeroTierInitializer struct {
	client *zerotier.Client
}

// NewZeroTierInitializer 创建新的ZeroTier初始化器
func NewZeroTierInitializer() *ZeroTierInitializer {
	return &ZeroTierInitializer{}
}

// Initialize 初始化ZeroTier客户端
func (zi *ZeroTierInitializer) Initialize() (*zerotier.Client, error) {
	// 获取配置
	cfg := config.AppConfig
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	// 只有在系统已初始化的情况下才自动初始化ZeroTier客户端
	if !cfg.Initialized {
		logger.Info("系统未初始化，跳过ZeroTier客户端自动初始化")
		return nil, nil
	}

	logger.Info("系统已初始化，开始初始化ZeroTier客户端")

	// 动态创建ZeroTier客户端
	ztClient, err := zerotier.NewClient()
	if err != nil {
		return nil, fmt.Errorf("创建ZeroTier客户端失败: %w", err)
	}

	// 验证客户端是否正常工作
	_, err = ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("ZeroTier客户端验证失败: %w", err)
	}

	zi.client = ztClient
	logger.Info("ZeroTier客户端初始化完成")
	return ztClient, nil
}

// GetClient 获取已初始化的ZeroTier客户端
func (zi *ZeroTierInitializer) GetClient() *zerotier.Client {
	return zi.client
}

// TestConnection 测试ZeroTier连接
func (zi *ZeroTierInitializer) TestConnection() error {
	if zi.client == nil {
		return fmt.Errorf("ZeroTier客户端未初始化")
	}

	_, err := zi.client.GetStatus()
	if err != nil {
		return fmt.Errorf("ZeroTier连接测试失败: %w", err)
	}

	return nil
}

// SaveConfig 保存ZeroTier配置
func (zi *ZeroTierInitializer) SaveConfig(url, token string) error {
	cfg := config.AppConfig
	if cfg == nil {
		return fmt.Errorf("配置未加载")
	}

	// 更新配置
	cfg.ZeroTier.URL = url
	cfg.ZeroTier.Token = token

	// 保存配置到文件
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	return nil
}
