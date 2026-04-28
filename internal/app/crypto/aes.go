package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrEncryptFailed     = errors.New("crypto.encrypt_failed")
	ErrDecryptFailed     = errors.New("crypto.decrypt_failed")
	ErrInvalidCiphertext = errors.New("crypto.invalid_ciphertext")
)

// Encrypt 使用AES-GCM加密敏感信息
func Encrypt(text, key string) (string, error) {
	// 确保key长度为32字节（256位）
	if len(key) < 32 {
		// 填充key到32字节
		for len(key) < 32 {
			key += "0"
		}
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher([]byte(key))
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

// Decrypt 解密敏感信息
func Decrypt(encryptedText, key string) (string, error) {
	// 确保key长度为32字节（256位）
	if len(key) < 32 {
		// 填充key到32字节
		for len(key) < 32 {
			key += "0"
		}
	} else if len(key) > 32 {
		key = key[:32]
	}

	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("%w: base64 decode: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher([]byte(key))
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
