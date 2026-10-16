package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigGeneratesAndPersistsEncryptionKey(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")

	first, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotEmpty(t, first.Security.EncryptionKey)

	configBytes, err := os.ReadFile(filepath.Join("data", "config.json"))
	require.NoError(t, err)
	var persisted config.Config
	require.NoError(t, json.Unmarshal(configBytes, &persisted))
	assert.Equal(t, first.Security.EncryptionKey, persisted.Security.EncryptionKey)

	second, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, first.Security.EncryptionKey, second.Security.EncryptionKey)
	assert.Empty(t, second.ZeroTier.Token)
}

func TestLoadConfigMigratesLegacyCredentialsEncryptedWithEmptyEncryptionKey(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")

	legacyConfig := &config.Config{
		ZeroTier: config.ZeroTierConfig{
			Token: "encrypted:" + encryptWithLegacyEmptyKey(t, "legacy-controller-token"),
		},
		Database: config.DatabaseConfig{
			Pass: "encrypted:" + encryptWithLegacyEmptyKey(t, "legacy-database-password"),
		},
	}
	require.NoError(t, config.SaveConfig(legacyConfig))

	repaired, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotEmpty(t, repaired.Security.EncryptionKey)

	decrypted, err := config.GetZTTokenFrom(repaired)
	require.NoError(t, err)
	assert.Equal(t, "legacy-controller-token", decrypted)
	databasePassword, err := config.GetDatabasePasswordFrom(repaired)
	require.NoError(t, err)
	assert.Equal(t, "legacy-database-password", databasePassword)

	reloaded, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, repaired.Security.EncryptionKey, reloaded.Security.EncryptionKey)
	decrypted, err = config.GetZTTokenFrom(reloaded)
	require.NoError(t, err)
	assert.Equal(t, "legacy-controller-token", decrypted)
	databasePassword, err = config.GetDatabasePasswordFrom(reloaded)
	require.NoError(t, err)
	assert.Equal(t, "legacy-database-password", databasePassword)
}

func TestLoadConfigContinuesWhenTokenRetryFails(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("ZT_TOKEN_PATH", filepath.Join(t.TempDir(), "missing-authtoken.secret"))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Security.EncryptionKey)
	assert.Empty(t, cfg.ZeroTier.Token)
}

// The encryption key used to be called jwt_secret. Adopting it matters because
// the key protects credentials that are already stored: generating a fresh one
// on upgrade would lose them.
func TestLoadConfigAdoptsTheLegacyEncryptionKeyName(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")

	require.NoError(t, os.MkdirAll("data", 0755))
	legacy := map[string]any{
		"initialized": true,
		"security":    map[string]any{"jwt_secret": "existing-secret"},
	}
	configBytes, err := json.Marshal(legacy)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join("data", "config.json"), configBytes, 0600))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "existing-secret", cfg.Security.EncryptionKey)

	// Saving rewrites it under the current name, so the old one disappears.
	require.NoError(t, config.SaveConfig(cfg))
	rewritten, err := os.ReadFile(filepath.Join("data", "config.json"))
	require.NoError(t, err)

	var persisted struct {
		Security map[string]any `json:"security"`
	}
	require.NoError(t, json.Unmarshal(rewritten, &persisted))
	assert.Equal(t, "existing-secret", persisted.Security["encryption_key"])
	assert.NotContains(t, persisted.Security, "jwt_secret")
}

func TestLoadConfigReadsTheEncryptionKeyFromTheEnvironment(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "from-the-environment")
	t.Setenv("JWT_SECRET", "from-the-old-name")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "from-the-environment", cfg.Security.EncryptionKey)
}

// An installation that still sets the old variable must keep working rather
// than silently generating a new key.
func TestLoadConfigFallsBackToTheDeprecatedEncryptionKeyVariable(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "from-the-old-name")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "from-the-old-name", cfg.Security.EncryptionKey)
}

// The console port is fixed. An old config.json still carries a "server" block
// that described the separate backend process, and its value must not influence
// the port the console listens on.
func TestLoadConfigIgnoresLegacyServerPortFromExistingConfig(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")

	require.NoError(t, os.MkdirAll("data", 0755))
	legacy := map[string]any{
		"initialized": true,
		"server":      map[string]any{"port": 8080},
		"security":    map[string]any{"jwt_secret": "existing-secret"},
		"registration": map[string]any{
			"allow_public_registration": true,
		},
	}
	configBytes, err := json.Marshal(legacy)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join("data", "config.json"), configBytes, 0600))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, config.ConsolePort, 3000)
	assert.Equal(t, ":3000", config.ServerAddress())
}

// A stale SERVER_PORT in the environment must not move the console, since the
// port is no longer configurable and the operator maps it from the host.
func TestLoadConfigIgnoresServerPortEnvironment(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("SERVER_PORT", "9090")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 3000, config.ConsolePort)
	assert.Equal(t, ":3000", config.ServerAddress())
}

// A legacy "server" block must not be reintroduced when the config is rewritten,
// so an upgraded deployment does not keep carrying a field that has no meaning.
func TestSaveConfigDoesNotWriteLegacyServerBlock(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("TAIRITSU_ENCRYPTION_KEY", "")
	t.Setenv("JWT_SECRET", "")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NoError(t, config.SaveConfig(cfg))

	configBytes, err := os.ReadFile(filepath.Join("data", "config.json"))
	require.NoError(t, err)

	var persisted map[string]any
	require.NoError(t, json.Unmarshal(configBytes, &persisted))
	assert.NotContains(t, persisted, "server")
}

func useTemporaryWorkingDirectory(t *testing.T) {
	t.Helper()
	originalWorkingDirectory, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWorkingDirectory))
	})
}

func encryptWithLegacyEmptyKey(t *testing.T, plaintext string) string {
	t.Helper()
	block, err := aes.NewCipher([]byte("00000000000000000000000000000000"))
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil))
}

func TestTempSettings(t *testing.T) {
	// Arrange
	testKey := "test-key"
	testValue := "test-value"

	// Act - Set temp setting
	config.SetTempSetting(testKey, testValue)

	// Assert - Get temp setting
	retrievedValue := config.GetTempSetting(testKey)
	assert.Equal(t, testValue, retrievedValue)

	// Act - Set another value for the same key
	newValue := "new-test-value"
	config.SetTempSetting(testKey, newValue)

	// Assert - Get updated temp setting
	retrievedValue = config.GetTempSetting(testKey)
	assert.Equal(t, newValue, retrievedValue)

	// Act - Get non-existent key
	nonExistentValue := config.GetTempSetting("non-existent-key")

	// Assert - Should return empty string for non-existent key
	assert.Empty(t, nonExistentValue)
}
