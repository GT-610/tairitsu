package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encryptLegacy simulates the v0.x zero-padding AES-GCM encryption.
func encryptLegacy(t *testing.T, plaintext, key string) string {
	t.Helper()
	for len(key) < 32 {
		key += "0"
	}
	if len(key) > 32 {
		key = key[:32]
	}
	block, err := aes.NewCipher([]byte(key))
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func TestAES_EncryptDecrypt(t *testing.T) {
	// Arrange
	key := "test-secret-key-32-characters-long"
	plaintext := "test-sensitive-data"

	// Act
	encrypted, err := crypto.Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := crypto.Decrypt(encrypted, key)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, plaintext, decrypted)
}

func TestAES_EncryptDecrypt_ShortKey(t *testing.T) {
	// Arrange
	key := "short-key"
	plaintext := "test-sensitive-data"

	// Act
	encrypted, err := crypto.Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := crypto.Decrypt(encrypted, key)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, plaintext, decrypted)
}

func TestAES_EncryptDecrypt_LongKey(t *testing.T) {
	// Arrange
	key := "this-is-a-very-long-key-that-exceeds-thirty-two-characters"
	plaintext := "test-sensitive-data"

	// Act
	encrypted, err := crypto.Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := crypto.Decrypt(encrypted, key)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, plaintext, decrypted)
}

func TestAES_Decrypt_InvalidCiphertext(t *testing.T) {
	// Arrange
	key := "test-secret-key-32-characters-long"
	invalidCiphertext := "invalid-base64-string"

	// Act
	decrypted, err := crypto.Decrypt(invalidCiphertext, key)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, crypto.ErrInvalidCiphertext))
	assert.Empty(t, decrypted)
}

func TestAES_Decrypt_WrongKey(t *testing.T) {
	// Arrange
	key1 := "test-secret-key-32-characters-long"
	key2 := "different-secret-key-32-characters"
	plaintext := "test-sensitive-data"

	// Act
	encrypted, err := crypto.Encrypt(plaintext, key1)
	assert.NoError(t, err)

	decrypted, err := crypto.Decrypt(encrypted, key2)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, crypto.ErrDecryptFailed))
	assert.Empty(t, decrypted)
}

func TestAES_Decrypt_EmptyCiphertext(t *testing.T) {
	// Arrange
	key := "test-secret-key-32-characters-long"
	emptyCiphertext := ""

	// Act
	decrypted, err := crypto.Decrypt(emptyCiphertext, key)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, crypto.ErrInvalidCiphertext))
	assert.Empty(t, decrypted)
}

func TestAES_ErrorSentinels(t *testing.T) {
	assert.True(t, errors.Is(fmt.Errorf("%w: detail", crypto.ErrEncryptFailed), crypto.ErrEncryptFailed))
	assert.True(t, errors.Is(fmt.Errorf("%w: detail", crypto.ErrDecryptFailed), crypto.ErrDecryptFailed))
	assert.True(t, errors.Is(fmt.Errorf("%w: detail", crypto.ErrInvalidCiphertext), crypto.ErrInvalidCiphertext))
}

func TestDecryptWithLegacy_UpgradesFromZeroPadding(t *testing.T) {
	key := "short-key"
	plaintext := "my-secret-token"

	// Encrypt with the old zero-padding method
	legacyEncrypted := encryptLegacy(t, plaintext, key)
	require.NotEmpty(t, legacyEncrypted)

	// DecryptWithLegacy should succeed and flag for re-encryption
	got, needsReEncrypt, err := crypto.DecryptWithLegacy(legacyEncrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
	assert.True(t, needsReEncrypt, "legacy ciphertext should trigger re-encryption")

	// After re-encrypting with the new method, plain Decrypt should work
	reEncrypted, err := crypto.Encrypt(plaintext, key)
	require.NoError(t, err)

	got2, err := crypto.Decrypt(reEncrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got2)
}

func TestDecryptWithLegacy_NoMigrationForNewCiphertext(t *testing.T) {
	key := "short-key"
	plaintext := "my-secret-token"

	// Encrypt with the new argon2id method
	encrypted, err := crypto.Encrypt(plaintext, key)
	require.NoError(t, err)

	// DecryptWithLegacy should succeed without needing migration
	got, needsReEncrypt, err := crypto.DecryptWithLegacy(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
	assert.False(t, needsReEncrypt, "new ciphertext should not trigger re-encryption")
}
