package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

var (
	ErrEncryptFailed     = errors.New("crypto.encrypt_failed")
	ErrDecryptFailed     = errors.New("crypto.decrypt_failed")
	ErrInvalidCiphertext = errors.New("crypto.invalid_ciphertext")
	ErrEmptyKey          = errors.New("crypto.empty_key")
)

// deriveKey uses argon2id to derive a 32-byte AES-256 key from an arbitrary-length secret.
func deriveKey(secret string) ([]byte, error) {
	if secret == "" {
		return nil, ErrEmptyKey
	}
	salt := sha256.Sum256([]byte("tairitsu-config-encryption-salt-v1"))
	return argon2.IDKey([]byte(secret), salt[:], 1, 64*1024, 4, 32), nil
}

// deriveKeyLegacy pads the secret with ASCII '0' to 32 bytes (v0.x behavior).
func deriveKeyLegacy(secret string) []byte {
	key := secret
	if len(key) < 32 {
		for len(key) < 32 {
			key += "0"
		}
	} else if len(key) > 32 {
		key = key[:32]
	}
	return []byte(key)
}

func encryptWithKey(text string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("%w: create AES cipher: %v", ErrEncryptFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: create GCM mode: %v", ErrEncryptFailed, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: generate nonce: %v", ErrEncryptFailed, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptWithKey(encryptedText string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("%w: base64 decode: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("%w: create AES cipher: %v", ErrDecryptFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: create GCM mode: %v", ErrDecryptFailed, err)
	}

	if len(data) < gcm.NonceSize() {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return string(plaintext), nil
}

// Encrypt encrypts sensitive data using AES-256-GCM with a key derived via argon2id.
func Encrypt(text, key string) (string, error) {
	derived, err := deriveKey(key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}
	return encryptWithKey(text, derived)
}

// Decrypt decrypts sensitive data using AES-256-GCM with a key derived via argon2id.
func Decrypt(encryptedText, key string) (string, error) {
	derived, err := deriveKey(key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}
	return decryptWithKey(encryptedText, derived)
}

// DecryptWithLegacy tries the current argon2id derivation first; if that fails
// it falls back to the legacy zero-padding key used in v0.x.  When the legacy
// path succeeds it returns needsReEncrypt=true so the caller can transparently
// upgrade the stored ciphertext.
func DecryptWithLegacy(encryptedText, key string) (plaintext string, needsReEncrypt bool, err error) {
	plaintext, err = Decrypt(encryptedText, key)
	if err == nil {
		return plaintext, false, nil
	}

	// Fall back to legacy zero-padding key derivation
	legacyKey := deriveKeyLegacy(key)
	plaintext, legacyErr := decryptWithKey(encryptedText, legacyKey)
	if legacyErr != nil {
		return "", false, err // return the original argon2id error
	}

	return plaintext, true, nil
}
