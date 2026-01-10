package auth

import (
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key := GenerateKey()
	
	// Should be 64 hex characters (32 bytes)
	if len(key) != 64 {
		t.Errorf("Expected key length 64, got %d", len(key))
	}
	
	// Should be unique each time
	key2 := GenerateKey()
	if key == key2 {
		t.Error("Generated keys should be unique")
	}
}

func TestNewTokenCrypto_ValidKey(t *testing.T) {
	key := GenerateKey()
	crypto, err := NewTokenCrypto(key)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if crypto == nil {
		t.Error("Expected crypto instance, got nil")
	}
}

func TestNewTokenCrypto_InvalidHex(t *testing.T) {
	_, err := NewTokenCrypto("not-hex")
	
	if err == nil {
		t.Error("Expected error for invalid hex, got nil")
	}
	
	if !strings.Contains(err.Error(), "invalid encryption key") {
		t.Errorf("Expected 'invalid encryption key' error, got %v", err)
	}
}

func TestNewTokenCrypto_WrongLength(t *testing.T) {
	// 16 bytes instead of 32
	shortKey := "0123456789abcdef0123456789abcdef"
	_, err := NewTokenCrypto(shortKey)
	
	if err == nil {
		t.Error("Expected error for wrong key length, got nil")
	}
	
	if !strings.Contains(err.Error(), "32 bytes") {
		t.Errorf("Expected '32 bytes' error, got %v", err)
	}
}

func TestEncryptDecrypt_Success(t *testing.T) {
	key := GenerateKey()
	crypto, _ := NewTokenCrypto(key)
	
	plaintext := "my-secret-oauth2-token-12345"
	
	// Encrypt
	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Errorf("Encrypt failed: %v", err)
	}
	
	if ciphertext == "" {
		t.Error("Expected non-empty ciphertext")
	}
	
	if ciphertext == plaintext {
		t.Error("Ciphertext should not equal plaintext")
	}
	
	// Decrypt
	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
	}
	
	if decrypted != plaintext {
		t.Errorf("Expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestEncryptDecrypt_EmptyString(t *testing.T) {
	key := GenerateKey()
	crypto, _ := NewTokenCrypto(key)
	
	ciphertext, err := crypto.Encrypt("")
	if err != nil {
		t.Errorf("Encrypt empty string failed: %v", err)
	}
	
	if ciphertext != "" {
		t.Error("Expected empty ciphertext for empty plaintext")
	}
	
	decrypted, err := crypto.Decrypt("")
	if err != nil {
		t.Errorf("Decrypt empty string failed: %v", err)
	}
	
	if decrypted != "" {
		t.Error("Expected empty plaintext for empty ciphertext")
	}
}

func TestEncryptDecrypt_LongString(t *testing.T) {
	key := GenerateKey()
	crypto, _ := NewTokenCrypto(key)
	
	// Long OAuth2 token
	plaintext := strings.Repeat("a", 1000)
	
	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Errorf("Encrypt failed: %v", err)
	}
	
	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
	}
	
	if decrypted != plaintext {
		t.Error("Decrypted text does not match original")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	key := GenerateKey()
	crypto, _ := NewTokenCrypto(key)
	
	// Try to decrypt garbage
	_, err := crypto.Decrypt("not-valid-base64-!@#$")
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := GenerateKey()
	key2 := GenerateKey()
	
	crypto1, _ := NewTokenCrypto(key1)
	crypto2, _ := NewTokenCrypto(key2)
	
	plaintext := "secret-token"
	ciphertext, _ := crypto1.Encrypt(plaintext)
	
	// Try to decrypt with different key
	_, err := crypto2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key, got nil")
	}
}

func TestEncrypt_UniqueOutputs(t *testing.T) {
	key := GenerateKey()
	crypto, _ := NewTokenCrypto(key)
	
	plaintext := "same-plaintext"
	
	// Encrypt same plaintext twice
	cipher1, _ := crypto.Encrypt(plaintext)
	cipher2, _ := crypto.Encrypt(plaintext)
	
	// Should produce different ciphertexts due to random nonce
	if cipher1 == cipher2 {
		t.Error("Expected different ciphertexts for same plaintext (nonce should be random)")
	}
	
	// But both should decrypt to same plaintext
	decrypted1, _ := crypto.Decrypt(cipher1)
	decrypted2, _ := crypto.Decrypt(cipher2)
	
	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}
