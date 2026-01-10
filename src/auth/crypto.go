package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// TokenCrypto provides encryption/decryption for OAuth2 tokens at rest
type TokenCrypto interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESTokenCrypto implements TokenCrypto using AES-256-GCM
type AESTokenCrypto struct {
	key []byte
}

// NewTokenCrypto creates a new token crypto instance
// key must be 32 bytes (256 bits) for AES-256
func NewTokenCrypto(hexKey string) (TokenCrypto, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, errors.New("invalid encryption key: must be hex-encoded")
	}
	
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes (64 hex characters)")
	}
	
	return &AESTokenCrypto{key: key}, nil
}

// GenerateKey generates a random 32-byte key for AES-256
func GenerateKey() string {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return hex.EncodeToString(key)
}

// Encrypt encrypts plaintext using AES-256-GCM
// Returns base64-encoded ciphertext
func (c *AESTokenCrypto) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
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
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (c *AESTokenCrypto) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	
	data, err := base64.StdEncoding.DecodeString(ciphertext)
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
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	
	nonce, cipherbytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherbytes, nil)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}
