package initializer

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/zerotier"
)

// ZeroTierInitializer ZeroTier client initializer
type ZeroTierInitializer struct {
	client *zerotier.Client
}

// NewZeroTierInitializer Create a new ZeroTier initializer
func NewZeroTierInitializer() *ZeroTierInitializer {
	return &ZeroTierInitializer{}
}

// Initialize Initialize ZeroTier client
func (zi *ZeroTierInitializer) Initialize() (*zerotier.Client, error) {
	// Get configuration
	cfg := config.AppConfig
	if cfg == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	// Only automatically initialize ZeroTier client if system is already initialized
	if !cfg.Initialized {
		logger.Info("System not initialized, skipping ZeroTier client automatic initialization")
		return nil, nil
	}

	logger.Info("System initialized, starting ZeroTier client initialization")

	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create ZeroTier client: %w", err)
	}

	// Verify client is working properly
	_, err = ztClient.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("ZeroTier client verification failed: %w", err)
	}

	zi.client = ztClient
	logger.Info("ZeroTier client initialization completed")
	return ztClient, nil
}

// GetClient Get initialized ZeroTier client
func (zi *ZeroTierInitializer) GetClient() *zerotier.Client {
	return zi.client
}

// TestConnection Test ZeroTier connection
func (zi *ZeroTierInitializer) TestConnection() error {
	if zi.client == nil {
		return fmt.Errorf("ZeroTier client not initialized")
	}

	_, err := zi.client.GetStatus()
	if err != nil {
		return fmt.Errorf("ZeroTier connection test failed: %w", err)
	}

	return nil
}

// SaveConfig Save ZeroTier configuration
func (zi *ZeroTierInitializer) SaveConfig(url, token string) error {
	cfg := config.AppConfig
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Update configuration
	cfg.ZeroTier.URL = url
	cfg.ZeroTier.Token = token

	// Save configuration to file
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}
