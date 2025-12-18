package config

import (
	"encoding/json"
	"fmt"
	"os"
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
		_ = LoadTokenFromPath(tokenPath)
	}
}

// GetZTToken Get ZeroTier token (auto-decrypted)
func GetZTToken() (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	// Decrypt token
	if AppConfig.ZeroTier.Token != "" {
		decryptedToken, err := decryptSensitiveData(AppConfig.ZeroTier.Token)
		if err != nil {
			// If decryption fails, it might be unencrypted data, try to return directly
			// This is for compatibility with possible unencrypted data
			return AppConfig.ZeroTier.Token, nil
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

// SetZTToken Set ZeroTier token (auto-encrypted)
func SetZTToken(token string) error {
	if AppConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Encrypt token
	encryptedToken, err := encryptSensitiveData(token)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	AppConfig.ZeroTier.Token = encryptedToken
	return nil
}

// SetZTConfig Set ZeroTier configuration
func SetZTConfig(url, tokenPath string) error {
	if AppConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	AppConfig.ZeroTier.URL = url
	AppConfig.ZeroTier.TokenPath = tokenPath

	// Try to read token from new tokenPath
	err := LoadTokenFromPath(tokenPath)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	return SaveConfig(AppConfig)
}

// LoadTokenFromPath Load ZeroTier token from specified path
func LoadTokenFromPath(path string) error {
	if AppConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Try to read token file
	tokenBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	// Successfully read file, remove newline and set token
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("token file is empty")
	}

	return SetZTToken(token)
}

// GetDatabasePassword Get database password (auto-decrypted)
func GetDatabasePassword() (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	if AppConfig.Database.Pass != "" {
		decryptedPass, err := decryptSensitiveData(AppConfig.Database.Pass)
		if err != nil {
			// If decryption fails, it might be unencrypted data, try to return directly
			return AppConfig.Database.Pass, nil
		}
		return decryptedPass, nil
	}

	return "", nil
}

// SetDatabasePassword Set database password (auto-encrypted)
func SetDatabasePassword(password string) error {
	if AppConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Encrypt password
	encryptedPass, err := encryptSensitiveData(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	AppConfig.Database.Pass = encryptedPass
	return nil
}

// encryptSensitiveData Encrypt sensitive data
func encryptSensitiveData(data string) (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	// Use JWT secret as encryption key (simplified solution)
	// In production, it's recommended to use a separate encryption key
	key := AppConfig.Security.JWTSecret

	// Encrypt data
	encrypted, err := crypto.Encrypt(data, key)
	if err != nil {
		return "", err
	}

	// Add encryption identifier prefix
	return "encrypted:" + encrypted, nil
}

// decryptSensitiveData Decrypt sensitive data
func decryptSensitiveData(data string) (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	// Check if it has encryption identifier prefix
	if !strings.HasPrefix(data, "encrypted:") {
		// If no encryption identifier, return original data directly
		return data, nil
	}

	// Remove encryption identifier prefix
	encryptedData := strings.TrimPrefix(data, "encrypted:")

	// Use JWT secret as decryption key
	key := AppConfig.Security.JWTSecret

	// Decrypt data
	return crypto.Decrypt(encryptedData, key)
}

// SetInitialized Set initialization status
func SetInitialized(initialized bool) error {
	if AppConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	AppConfig.Initialized = initialized
	return SaveConfig(AppConfig)
}

// IsInitialized Check if initialized
func IsInitialized() bool {
	if AppConfig == nil {
		return false
	}
	return AppConfig.Initialized
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
