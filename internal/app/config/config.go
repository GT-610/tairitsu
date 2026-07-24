package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/GT-610/tairitsu/internal/app/crypto"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// DatabaseConfig Database configuration
type DatabaseConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"` // Encrypted password
	Name string `json:"name"`
}

// ZeroTierConfig ZeroTier configuration
type ZeroTierConfig struct {
	URL       string `json:"url"`
	Token     string `json:"token"`     // Encrypted token
	TokenPath string `json:"tokenPath"` // Token file path
}

// ServerConfig Server configuration
type ServerConfig struct {
	Port int `json:"port"`
}

// SecurityConfig Security configuration
type SecurityConfig struct {
	JWTSecret string `json:"jwt_secret"`
}

type RegistrationConfig struct {
	AllowPublicRegistration *bool `json:"allow_public_registration,omitempty"`
}

// Config Application configuration structure
type Config struct {
	Initialized  bool               `json:"initialized"` // Initialization status flag
	Database     DatabaseConfig     `json:"database"`    // Database configuration
	ZeroTier     ZeroTierConfig     `json:"zerotier"`    // ZeroTier configuration
	Server       ServerConfig       `json:"server"`      // Server configuration
	Security     SecurityConfig     `json:"security"`    // Security configuration
	Registration RegistrationConfig `json:"registration"`
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
		generated, err := ensureJWTSecret(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		if generated {
			if err := SaveConfig(cfg); err != nil {
				return nil, fmt.Errorf("failed to save generated JWT secret: %w", err)
			}
		}
		AppConfig = cfg
		return cfg, nil
	}

	// If config.json doesn't exist or reading fails, create default configuration
	cfg = createDefaultConfig()

	// Try to load partial configuration from .env file or environment variables
	loadEnvConfig(cfg)
	generated, err := ensureJWTSecret(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	if generated && cfg.ZeroTier.TokenPath != "" && cfg.ZeroTier.Token == "" {
		// Loading an environment-provided token path may have failed before a
		// generated encryption key was available. Retry after creating the key.
		if err := LoadTokenFromPathInto(cfg, cfg.ZeroTier.TokenPath); err != nil {
			logger.Warn("failed to load environment-provided ZeroTier token after generating JWT secret; continuing without token", zap.Error(err))
		}
	}

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
		},
		Security: SecurityConfig{
			JWTSecret: "", // Generated and persisted by LoadConfig before service assembly.
		},
		Registration: RegistrationConfig{
			AllowPublicRegistration: boolPtr(true),
		},
	}
}

func ensureJWTSecret(cfg *Config) (bool, error) {
	if cfg.Security.JWTSecret != "" {
		return false, nil
	}

	token, err := decryptWithEmptyLegacyKey(cfg.ZeroTier.Token)
	if err != nil {
		return false, fmt.Errorf("failed to recover ZeroTier token: %w", err)
	}
	databasePassword, err := decryptWithEmptyLegacyKey(cfg.Database.Pass)
	if err != nil {
		return false, fmt.Errorf("failed to recover database password: %w", err)
	}

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return false, err
	}
	cfg.Security.JWTSecret = base64.URLEncoding.EncodeToString(secretBytes)
	if token != "" {
		if err := SetZTTokenOn(cfg, token); err != nil {
			return false, err
		}
	}
	if databasePassword != "" {
		if err := SetDatabasePasswordOn(cfg, databasePassword); err != nil {
			return false, err
		}
	}
	return true, nil
}

func decryptWithEmptyLegacyKey(value string) (string, error) {
	if value == "" || !strings.HasPrefix(value, "encrypted:") {
		return value, nil
	}

	plaintext, _, err := crypto.DecryptWithLegacy(strings.TrimPrefix(value, "encrypted:"), "")
	if err != nil {
		return "", err
	}
	return plaintext, nil
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

	if cfg.ZeroTier.Token != "" {
		plaintext, reEncrypted, err := decryptSensitiveDataWithConfig(cfg, cfg.ZeroTier.Token)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt ZeroTier token: %w", err)
		}
		if reEncrypted != "" {
			cfg.ZeroTier.Token = reEncrypted
			if saveErr := SaveConfig(cfg); saveErr != nil {
				return plaintext, fmt.Errorf("decrypted token OK but failed to persist re-encrypted config: %w", saveErr)
			}
		}
		return plaintext, nil
	}

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

	if err := LoadTokenFromPathInto(cfg, tokenPath); err != nil {
		return err
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
		plaintext, reEncrypted, err := decryptSensitiveDataWithConfig(cfg, cfg.Database.Pass)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt database password: %w", err)
		}
		if reEncrypted != "" {
			cfg.Database.Pass = reEncrypted
			if saveErr := SaveConfig(cfg); saveErr != nil {
				return plaintext, fmt.Errorf("decrypted password OK but failed to persist re-encrypted config: %w", saveErr)
			}
		}
		return plaintext, nil
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

// decryptSensitiveDataWithConfig decrypts a value, transparently migrating
// legacy zero-padding ciphertext to argon2id on first access.
// Returns the plaintext and a re-encrypted value if migration was needed.
func decryptSensitiveDataWithConfig(cfg *Config, data string) (plaintext string, reEncrypted string, err error) {
	if cfg == nil {
		return "", "", fmt.Errorf("configuration not loaded")
	}

	if !strings.HasPrefix(data, "encrypted:") {
		return data, "", nil
	}

	encryptedData := strings.TrimPrefix(data, "encrypted:")
	key := cfg.Security.JWTSecret
	plaintext, needsReEncrypt, err := crypto.DecryptWithLegacy(encryptedData, key)
	if err != nil {
		return "", "", err
	}

	if needsReEncrypt {
		re, encErr := crypto.Encrypt(plaintext, key)
		if encErr != nil {
			return "", "", fmt.Errorf("failed to re-encrypt with new key derivation: %w", encErr)
		}
		return plaintext, "encrypted:" + re, nil
	}

	return plaintext, "", nil
}

func AllowPublicRegistration(cfg *Config) bool {
	if cfg == nil || cfg.Registration.AllowPublicRegistration == nil {
		return true
	}

	return *cfg.Registration.AllowPublicRegistration
}

func SetAllowPublicRegistrationOn(cfg *Config, enabled bool) error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	cfg.Registration.AllowPublicRegistration = boolPtr(enabled)
	return SaveConfig(cfg)
}

func boolPtr(value bool) *bool {
	return &value
}

func ServerAddressFrom(cfg *Config) string {
	if cfg == nil {
		return ":8080"
	}
	return fmt.Sprintf(":%d", cfg.Server.Port)
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
