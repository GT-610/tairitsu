package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GT-610/tairitsu/internal/app/crypto"
	"github.com/spf13/viper"
)

// DatabaseConfig Database configuration
type DatabaseConfig struct {
	Type DatabaseType `json:"type"`
	Path string       `json:"path"`
	Host string       `json:"host"`
	Port int          `json:"port"`
	User string       `json:"user"`
	Pass string       `json:"pass"` // Encrypted password
	Name string       `json:"name"`
}

// DatabaseType Database type
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
)

// ZeroTierConfig ZeroTier configuration
type ZeroTierConfig struct {
	URL       string `json:"url"`
	Token     string `json:"token"`     // Encrypted token
	TokenPath string `json:"tokenPath"` // Token file path
}

// ServerConfig Server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// SecurityConfig Security configuration
type SecurityConfig struct {
	JWTSecret     string `json:"jwt_secret"`
	SessionSecret string `json:"session_secret"`
}

// Config Application configuration structure
type Config struct {
	Initialized bool           `json:"initialized"` // Initialization status flag
	Database    DatabaseConfig `json:"database"`    // Database configuration
	ZeroTier    ZeroTierConfig `json:"zerotier"`    // ZeroTier configuration
	Server      ServerConfig   `json:"server"`      // Server configuration
	Security    SecurityConfig `json:"security"`    // Security configuration
}

// AppConfig Global configuration instance
var AppConfig *Config

// tempSettings In-memory temporary settings for non-persistent configuration
var tempSettings = make(map[string]string)
var tempSettingsMutex sync.RWMutex

const configFilePath = "./data/config.json"

// LoadConfig Load configuration (from config.json)
func LoadConfig() (*Config, error) {
	// Ensure data directory exists
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// First try to load from config.json
	cfg, err := loadConfigFromJSON()
	if err == nil {
		AppConfig = cfg
		return cfg, nil
	}

	// If config.json doesn't exist or reading fails, create default configuration
	cfg = createDefaultConfig()

	// Try to load partial configuration from .env file or environment variables
	loadEnvConfig(cfg)

	// Save default configuration to config.json
	if err := SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save default configuration: %w", err)
	}

	AppConfig = cfg
	return cfg, nil
}

// loadConfigFromJSON Load configuration from JSON file
func loadConfigFromJSON() (*Config, error) {
	// Check if configuration file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found")
	}

	// Read configuration file
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse JSON
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	return cfg, nil
}

// SaveConfig Save configuration to JSON file
func SaveConfig(cfg *Config) error {
	// Serialize configuration
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Ensure data directory exists
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Write to file with appropriate permissions (owner read/write only)
	return os.WriteFile(configFilePath, data, 0600)
}

// createDefaultConfig Create default configuration
func createDefaultConfig() *Config {
	return &Config{
		Initialized: false,
		Database:    DatabaseConfig{},
		ZeroTier:    ZeroTierConfig{},
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Security: SecurityConfig{
			JWTSecret:     "", // Empty initially, force user to generate during first setup
			SessionSecret: "", // Empty initially, force user to generate during first setup
		},
	}
}

// loadEnvConfig Load partial configuration from environment variables (only read existing configurations in .env)
func loadEnvConfig(cfg *Config) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()

	// Try to read .env file
	_ = viper.ReadInConfig() // Ignore error, just try to read

	// Update configuration from environment variables, only read existing configurations in .env
	if url := viper.GetString("ZT_CONTROLLER_URL"); url != "" {
		cfg.ZeroTier.URL = url
	}
	if port := viper.GetInt("SERVER_PORT"); port != 0 {
		cfg.Server.Port = port
	}
	if jwt := viper.GetString("JWT_SECRET"); jwt != "" {
		cfg.Security.JWTSecret = jwt
	}
	if session := viper.GetString("SESSION_SECRET"); session != "" {
		cfg.Security.SessionSecret = session
	}

	// Read ZT_TOKEN_PATH and try to read token from file
	if tokenPath := viper.GetString("ZT_TOKEN_PATH"); tokenPath != "" {
		cfg.ZeroTier.TokenPath = tokenPath
		// Try to read token file
		_ = LoadTokenFromPathInto(cfg, tokenPath)
	}
}

func GetZTTokenFrom(cfg *Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	// Decrypt token
	if cfg.ZeroTier.Token != "" {
		decryptedToken, err := decryptSensitiveDataWithConfig(cfg, cfg.ZeroTier.Token)
		if err != nil {
			// If decryption fails, it might be unencrypted data, try to return directly
			// This is for compatibility with possible unencrypted data
			return cfg.ZeroTier.Token, nil
		}
		return decryptedToken, nil
	}

	// If not in configuration, try to get from environment variables
	token := viper.GetString("ZT_TOKEN")
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("ZeroTier token not configured")
}

func SetZTTokenOn(cfg *Config, token string) error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	encryptedToken, err := encryptSensitiveDataWithConfig(cfg, token)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	cfg.ZeroTier.Token = encryptedToken
	return nil
}

func SetZTConfigOn(cfg *Config, url, tokenPath string) error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	cfg.ZeroTier.URL = url
	cfg.ZeroTier.TokenPath = tokenPath

	err := LoadTokenFromPathInto(cfg, tokenPath)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	return SaveConfig(cfg)
}

func LoadTokenFromPathInto(cfg *Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	tokenBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	// Successfully read file, remove newline and set token
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("token file is empty")
	}

	return SetZTTokenOn(cfg, token)
}

func GetDatabasePasswordFrom(cfg *Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	if cfg.Database.Pass != "" {
		decryptedPass, err := decryptSensitiveDataWithConfig(cfg, cfg.Database.Pass)
		if err != nil {
			// If decryption fails, it might be unencrypted data, try to return directly
			return cfg.Database.Pass, nil
		}
		return decryptedPass, nil
	}

	return "", nil
}

func SetDatabasePasswordOn(cfg *Config, password string) error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	encryptedPass, err := encryptSensitiveDataWithConfig(cfg, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	cfg.Database.Pass = encryptedPass
	return nil
}

// encryptSensitiveData Encrypt sensitive data
func encryptSensitiveData(data string) (string, error) {
	return encryptSensitiveDataWithConfig(AppConfig, data)
}

func encryptSensitiveDataWithConfig(cfg *Config, data string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	key := cfg.Security.JWTSecret
	encrypted, err := crypto.Encrypt(data, key)
	if err != nil {
		return "", err
	}

	return "encrypted:" + encrypted, nil
}

// decryptSensitiveData Decrypt sensitive data
func decryptSensitiveData(data string) (string, error) {
	return decryptSensitiveDataWithConfig(AppConfig, data)
}

func decryptSensitiveDataWithConfig(cfg *Config, data string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	if !strings.HasPrefix(data, "encrypted:") {
		return data, nil
	}

	encryptedData := strings.TrimPrefix(data, "encrypted:")
	key := cfg.Security.JWTSecret
	return crypto.Decrypt(encryptedData, key)
}

// IsInitialized Check if initialized
func IsInitialized() bool {
	if AppConfig == nil {
		return false
	}
	return AppConfig.Initialized
}

// Current returns the currently loaded config instance.
func Current() (*Config, error) {
	if AppConfig == nil {
		return nil, fmt.Errorf("配置未加载")
	}
	return AppConfig, nil
}

// ServerAddress returns the configured server listen address.
func ServerAddress() (string, error) {
	cfg, err := Current()
	if err != nil {
		return "", err
	}
	return ServerAddressFrom(cfg), nil
}

func ServerAddressFrom(cfg *Config) string {
	if cfg == nil {
		return ":8080"
	}
	return fmt.Sprintf(":%d", cfg.Server.Port)
}

// ZeroTierSettings returns the currently configured ZeroTier settings.
func ZeroTierSettings() (ZeroTierConfig, error) {
	cfg, err := Current()
	if err != nil {
		return ZeroTierConfig{}, err
	}
	return cfg.ZeroTier, nil
}

// ConfigPath returns the absolute path for the config file location.
func ConfigPath() (string, error) {
	return filepath.Abs(configFilePath)
}

// GetTempSetting Get temporary setting
// Temporary settings are stored in memory and not persisted to configuration file
func GetTempSetting(key string) string {
	tempSettingsMutex.RLock()
	defer tempSettingsMutex.RUnlock()
	return tempSettings[key]
}

// SetTempSetting Set temporary setting
// Temporary settings are stored in memory and not persisted to configuration file
func SetTempSetting(key, value string) {
	tempSettingsMutex.Lock()
	defer tempSettingsMutex.Unlock()
	tempSettings[key] = value
}
