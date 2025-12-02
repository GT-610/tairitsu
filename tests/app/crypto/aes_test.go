package crypto

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/crypto"
	"github.com/stretchr/testify/assert"
)

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
	assert.Empty(t, decrypted)
}
