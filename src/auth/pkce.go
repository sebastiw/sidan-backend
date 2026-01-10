package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GeneratePKCEVerifier generates a random code verifier for PKCE
// Returns a 43-character URL-safe string (32 random bytes base64-encoded)
func GeneratePKCEVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Base64 URL encoding without padding as per RFC 7636
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GeneratePKCEChallenge generates the code challenge from a verifier
// Uses S256 method (SHA256 hash, base64-encoded)
func GeneratePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateState generates a random state parameter for CSRF protection
func GenerateState() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateNonce generates a random nonce
func GenerateNonce() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
