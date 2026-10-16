package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GT-610/tairitsu/internal/app/crypto"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseConfig describes where Tairitsu keeps its data.
//
// Only SQLite is implemented: the setup wizard rejects every other type, so the
// host, port, user, password and database-name fields that a MySQL or
// PostgreSQL backend would have needed are gone rather than carried as dead
// configuration.
type DatabaseConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// ZeroTierConfig ZeroTier configuration
type ZeroTierConfig struct {
	URL       string `json:"url"`
	Token     string `json:"token"`     // Encrypted token
	TokenPath string `json:"tokenPath"` // Token file path
}

// SecurityConfig holds the key that protects stored credentials.
type SecurityConfig struct {
	// EncryptionKey encrypts the ZeroTier token and, before those backends were
	// removed, the database password. It is generated on first start and stored
	// beside the ciphertext, so it protects against a leaked copy of one file
	// rather than against an attacker who can read the data directory.
	EncryptionKey string `json:"encryption_key"`
}

// EncryptionKeyEnv seeds the encryption key on a fresh instance.
const EncryptionKeyEnv = "TAIRITSU_ENCRYPTION_KEY"

// legacyEncryptionKeyEnv is the name this variable had while the key was also
// used to sign JSON Web Tokens. It is still read so that an existing
// deployment does not silently generate a new key, and can be dropped once no
// installation sets it.
const legacyEncryptionKeyEnv = "JWT_SECRET"

type RegistrationConfig struct {
	AllowPublicRegistration *bool `json:"allow_public_registration,omitempty"`
}

// ConsolePort is the port the web console and the API listen on. It is fixed
// rather than configurable: the container is the supported way to run Tairitsu,
// so the port a request arrives on inside the container is an implementation
// detail, and the port users reach is decided by their own mapping (for example
// `-p 8080:3000`). Keeping it constant also means an existing `-p 3000:3000`
// deployment keeps working, and no stored setting can silently move the console
// to a port the operator is not publishing.
const ConsolePort = 3000

// Config Application configuration structure
//
// A legacy `server` block may still be present in an existing config.json. It
// described the separate backend process that the bundled web server proxied
// to. That process no longer exists, and the field is not read: unknown JSON
// members are ignored on load, so old files keep working untouched.
type Config struct {
	Initialized  bool               `json:"initialized"` // Initialization status flag
	Database     DatabaseConfig     `json:"database"`    // Database configuration
	ZeroTier     ZeroTierConfig     `json:"zerotier"`    // ZeroTier configuration
	Security     SecurityConfig     `json:"security"`    // Security configuration
	Registration RegistrationConfig `json:"registration"`
}

// AppConfig Global configuration instance
var AppConfig *Config

// tempSettings In-memory temporary settings for non-persistent configuration
var tempSettings = make(map[string]string)
var tempSettingsMutex sync.RWMutex

// dataDirName holds Tairitsu's persistent state: the configuration file and the
// SQLite database created by the setup wizard.
const dataDirName = "data"

// DataDirEnv relocates dataDirName.
const DataDirEnv = "TAIRITSU_DATA_DIR"

// configFileName is the persisted configuration inside the data directory.
const configFileName = "config.json"

// defaultDatabaseFileName is the SQLite file used when setup does not name one.
const defaultDatabaseFileName = "tairitsu.db"

// DataRoot returns the absolute path of Tairitsu's data directory.
//
// Everything a request may ask the server to create or delete lives here, so
// the directory the operator mounts is also the boundary the server enforces.
// It is read on every call rather than cached, which keeps it correct for a
// process whose working directory or environment changes under a test.
func DataRoot() string {
	root := Env(DataDirEnv)
	if root == "" {
		root = dataDirName
	}

	absolute, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return absolute
}

// ConfigFilePath returns the absolute path of the persisted configuration.
func ConfigFilePath() string {
	return filepath.Join(DataRoot(), configFileName)
}

// ResolveDataPath resolves candidate to an absolute path and refuses anything
// that escapes Tairitsu's data directory. An empty candidate selects the
// default database file.
//
// Setup input decides where the SQLite file lives and a later reset deletes
// whatever that path names, so a request must not be able to point the server
// at an arbitrary file. Containment is decided on the symlink-resolved location
// while the returned path keeps the name the caller used, so a stored setting
// is not rewritten into a different spelling of the same directory.
func ResolveDataPath(candidate string) (string, error) {
	resolvedRoot := DataRoot()
	if evaluated, err := filepath.EvalSymlinks(resolvedRoot); err == nil {
		resolvedRoot = evaluated
	} else if err := os.MkdirAll(resolvedRoot, 0755); err == nil {
		if evaluated, evalErr := filepath.EvalSymlinks(resolvedRoot); evalErr == nil {
			resolvedRoot = evaluated
		}
	}

	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		trimmed = filepath.Join(DataRoot(), defaultDatabaseFileName)
	}
	absolute, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid path %q", candidate)
	}

	// Both sides are resolved the same way, so a path that does not exist yet is
	// compared against the location it would really live in rather than against
	// its unresolved spelling.
	relative, err := filepath.Rel(resolveExistingPrefix(resolvedRoot), resolveExistingPrefix(absolute))
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside the Tairitsu data directory (%s)", candidate, resolvedRoot)
	}
	if relative == "." {
		return "", fmt.Errorf("path %q is the Tairitsu data directory itself", candidate)
	}

	return absolute, nil
}

// resolveExistingPrefix resolves symlinks in the longest existing ancestor of
// path and re-appends the components that do not exist yet.
//
// A plain EvalSymlinks only works on a path that exists, which leaves a file
// that is about to be created with no resolved location to compare.
func resolveExistingPrefix(path string) string {
	var missing []string
	current := path

	for {
		if evaluated, err := filepath.EvalSymlinks(current); err == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				evaluated = filepath.Join(evaluated, missing[index])
			}
			return evaluated
		}

		parent := filepath.Dir(current)
		if parent == current {
			return path
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

// LoadConfig Load configuration (from config.json)
func LoadConfig() (*Config, error) {
	// Ensure data directory exists
	dataDir := DataRoot()
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// First try to load from config.json
	cfg, err := loadConfigFromJSON()
	if err == nil {
		generated, err := ensureEncryptionKey(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate the configuration encryption key: %w", err)
		}
		if generated {
			if err := SaveConfig(cfg); err != nil {
				return nil, fmt.Errorf("failed to save the generated configuration encryption key: %w", err)
			}
		}
		AppConfig = cfg
		return cfg, nil
	}

	// If config.json doesn't exist or reading fails, create default configuration
	cfg = createDefaultConfig()

	// Try to load partial configuration from .env file or environment variables
	loadEnvConfig(cfg)
	generated, err := ensureEncryptionKey(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the configuration encryption key: %w", err)
	}
	if generated && cfg.ZeroTier.TokenPath != "" && cfg.ZeroTier.Token == "" {
		// Loading an environment-provided token path may have failed before a
		// generated encryption key was available. Retry after creating the key.
		if err := LoadTokenFromPathInto(cfg, cfg.ZeroTier.TokenPath); err != nil {
			logger.Warn("failed to load environment-provided ZeroTier token after generating the encryption key; continuing without token", zap.Error(err))
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
	configFile := ConfigFilePath()

	// Check if configuration file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found")
	}

	// Read configuration file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse JSON
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	adoptLegacySecurityKey(cfg, data)

	return cfg, nil
}

// adoptLegacySecurityKey reads the encryption key from the name it had before
// the JWT service was removed.
//
// The field protects credentials that are already stored, so losing it would
// lose them; the value is carried over and written back under the new name the
// next time the configuration is saved.
func adoptLegacySecurityKey(cfg *Config, raw []byte) {
	if cfg == nil || cfg.Security.EncryptionKey != "" {
		return
	}

	var legacy struct {
		Security struct {
			Key string `json:"jwt_secret"`
		} `json:"security"`
	}
	if err := json.Unmarshal(raw, &legacy); err != nil || legacy.Security.Key == "" {
		return
	}

	logger.Warn("configuration stores its encryption key under the old jwt_secret name; it will be rewritten as encryption_key on the next save")
	cfg.Security.EncryptionKey = legacy.Security.Key
}

// SaveConfig Save configuration to JSON file
func SaveConfig(cfg *Config) error {
	// Serialize configuration
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Ensure data directory exists
	dataDir := DataRoot()
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Write to file with appropriate permissions (owner read/write only)
	return os.WriteFile(ConfigFilePath(), data, 0600)
}

// createDefaultConfig Create default configuration
func createDefaultConfig() *Config {
	return &Config{
		Initialized: false,
		Database:    DatabaseConfig{},
		ZeroTier:    ZeroTierConfig{},
		Security: SecurityConfig{
			EncryptionKey: "", // Generated and persisted by LoadConfig before service assembly.
		},
		Registration: RegistrationConfig{
			AllowPublicRegistration: boolPtr(true),
		},
	}
}

func ensureEncryptionKey(cfg *Config) (bool, error) {
	if cfg.Security.EncryptionKey != "" {
		return false, nil
	}

	token, err := decryptWithEmptyLegacyKey(cfg.ZeroTier.Token)
	if err != nil {
		return false, fmt.Errorf("failed to recover ZeroTier token: %w", err)
	}

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return false, err
	}
	cfg.Security.EncryptionKey = base64.URLEncoding.EncodeToString(secretBytes)
	if token != "" {
		if err := SetZTTokenOn(cfg, token); err != nil {
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

// loadEnvConfig seeds a configuration that does not exist yet from the
// environment, so a container can be started without a config.json.
//
// It runs only when config.json is absent: once the file exists it is the
// source of truth, and the setup wizard writes to it. A value that was seeded
// here is therefore persisted by the caller and is not re-read on later starts.
func loadEnvConfig(cfg *Config) {
	if url := Env("ZT_CONTROLLER_URL"); url != "" {
		cfg.ZeroTier.URL = url
	}
	if key := Env(EncryptionKeyEnv); key != "" {
		cfg.Security.EncryptionKey = key
	} else if legacy := Env(legacyEncryptionKeyEnv); legacy != "" {
		logger.Warn("ignoring the deprecated environment variable; rename it",
			zap.String("deprecated", legacyEncryptionKeyEnv), zap.String("replacement", EncryptionKeyEnv))
		cfg.Security.EncryptionKey = legacy
	}

	// Read ZT_TOKEN_PATH and try to read token from file
	if tokenPath := Env("ZT_TOKEN_PATH"); tokenPath != "" {
		cfg.ZeroTier.TokenPath = tokenPath
		// Try to read token file
		_ = LoadTokenFromPathInto(cfg, tokenPath)
	}
}

// GetZTTokenFrom retrieves the ZeroTier token from the configuration
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

	if token := Env("ZT_TOKEN"); token != "" {
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

// tokenFileSizeLimit caps how much of a token file is read. A ZeroTier
// authtoken is a few dozen bytes, and the path comes from a setup request, so
// reading whatever size the named file happens to be would be a denial of
// service.
const tokenFileSizeLimit = 4096

// LoadTokenFromPathInto reads the controller token from path and stores it
// encrypted in cfg.
//
// The path is chosen by the operator, so it cannot be confined to a directory
// the way the database path is: a ZeroTier data directory lives wherever the
// controller was installed. Instead the read is refused unless the target is a
// regular file of a plausible size.
func LoadTokenFromPathInto(cfg *Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return fmt.Errorf("token file path is empty")
	}

	info, err := os.Stat(trimmed)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("token file %q is not a regular file", trimmed)
	}

	file, err := os.Open(trimmed)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, tokenFileSizeLimit+1))
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}
	if len(content) > tokenFileSizeLimit {
		return fmt.Errorf("token file %q is larger than %d bytes", trimmed, tokenFileSizeLimit)
	}

	// Successfully read file, remove newline and set token
	token := strings.TrimSpace(string(content))
	if token == "" {
		return fmt.Errorf("token file is empty")
	}

	return SetZTTokenOn(cfg, token)
}

func encryptSensitiveDataWithConfig(cfg *Config, data string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	key := cfg.Security.EncryptionKey
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
	key := cfg.Security.EncryptionKey
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

// ServerAddress returns the listen address for the web console.
func ServerAddress() string {
	return fmt.Sprintf(":%d", ConsolePort)
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
