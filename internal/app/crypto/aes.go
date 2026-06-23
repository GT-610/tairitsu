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
// The application-specific salt prevents cross-application key reuse.
func deriveKey(secret string) ([]byte, error) {
	if secret == "" {
		return nil, ErrEmptyKey
	}
	salt := sha256.Sum256([]byte("tairitsu-config-encryption-salt-v1"))
	return argon2.IDKey([]byte(secret), salt[:], 1, 64*1024, 4, 32), nil
}

// Encrypt encrypts sensitive data using AES-256-GCM with a key derived via argon2id.
func Encrypt(text, key string) (string, error) {
	derived, err := deriveKey(key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}

	block, err := aes.NewCipher(derived)
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

// Decrypt decrypts sensitive data using AES-256-GCM with a key derived via argon2id.
func Decrypt(encryptedText, key string) (string, error) {
	derived, err := deriveKey(key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("%w: base64 decode: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(derived)
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
