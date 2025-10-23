package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// CryptoService provides symmetric encryption utilities using AES-GCM
type CryptoService struct {
	key []byte
}

var globalCryptoService *CryptoService

// SetGlobalCryptoService sets the global crypto service instance
func SetGlobalCryptoService(c *CryptoService) {
	globalCryptoService = c
}

// GetGlobalCryptoService returns the global crypto service instance
func GetGlobalCryptoService() *CryptoService {
	return globalCryptoService
}

// NewCryptoServiceFromEnv initializes CryptoService from CONFIG_ENCRYPTION_KEY
func NewCryptoServiceFromEnv() (*CryptoService, error) {
	secret := os.Getenv("CONFIG_ENCRYPTION_KEY")
	if secret == "" {
		return nil, errors.New("CONFIG_ENCRYPTION_KEY is required for secure settings")
	}
	// Derive 32-byte key from provided secret using SHA-256
	sum := sha256.Sum256([]byte(secret))
	return &CryptoService{key: sum[:]}, nil
}

// EncryptString encrypts plaintext using AES-GCM and returns base64 string (nonce|ciphertext)
func (c *CryptoService) EncryptString(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	combined := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptString decrypts a base64 (nonce|ciphertext) string using AES-GCM
func (c *CryptoService) DecryptString(encryptedB64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted payload")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
