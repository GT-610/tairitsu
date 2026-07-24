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

func TestLoadConfigGeneratesAndPersistsJWTSecret(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("JWT_SECRET", "")

	first, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotEmpty(t, first.Security.JWTSecret)

	configBytes, err := os.ReadFile(filepath.Join("data", "config.json"))
	require.NoError(t, err)
	var persisted config.Config
	require.NoError(t, json.Unmarshal(configBytes, &persisted))
	assert.Equal(t, first.Security.JWTSecret, persisted.Security.JWTSecret)

	second, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, first.Security.JWTSecret, second.Security.JWTSecret)
	assert.Empty(t, second.ZeroTier.Token)
}

func TestLoadConfigMigratesLegacyCredentialsEncryptedWithEmptyJWTSecret(t *testing.T) {
	useTemporaryWorkingDirectory(t)
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
	require.NotEmpty(t, repaired.Security.JWTSecret)

	decrypted, err := config.GetZTTokenFrom(repaired)
	require.NoError(t, err)
	assert.Equal(t, "legacy-controller-token", decrypted)
	databasePassword, err := config.GetDatabasePasswordFrom(repaired)
	require.NoError(t, err)
	assert.Equal(t, "legacy-database-password", databasePassword)

	reloaded, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, repaired.Security.JWTSecret, reloaded.Security.JWTSecret)
	decrypted, err = config.GetZTTokenFrom(reloaded)
	require.NoError(t, err)
	assert.Equal(t, "legacy-controller-token", decrypted)
	databasePassword, err = config.GetDatabasePasswordFrom(reloaded)
	require.NoError(t, err)
	assert.Equal(t, "legacy-database-password", databasePassword)
}

func TestLoadConfigContinuesWhenTokenRetryFails(t *testing.T) {
	useTemporaryWorkingDirectory(t)
	t.Setenv("JWT_SECRET", "")
	t.Setenv("ZT_TOKEN_PATH", filepath.Join(t.TempDir(), "missing-authtoken.secret"))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Security.JWTSecret)
	assert.Empty(t, cfg.ZeroTier.Token)
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
