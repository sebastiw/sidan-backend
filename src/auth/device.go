package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

// GenerateDeviceCode generates a random device code
func GenerateDeviceCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// GenerateUserCode generates an 8-character user code (e.g., ABCD-1234)
func GenerateUserCode() (string, error) {
	// Generate 5 random bytes (40 bits) for base32 encoding = 8 chars
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	
	// Base32 encode and take first 8 chars
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	code := encoded[:8]
	
	// Format as XXXX-XXXX
	return fmt.Sprintf("%s-%s", code[:4], code[4:8]), nil
}

// ValidateUserCode validates user code format
func ValidateUserCode(code string) bool {
	// Remove dash
	cleaned := strings.ReplaceAll(code, "-", "")
	
	// Check length
	if len(cleaned) != 8 {
		return false
	}
	
	// Check if all characters are valid base32 (A-Z, 2-7)
	for _, c := range cleaned {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')) {
			return false
		}
	}
	
	return true
}
