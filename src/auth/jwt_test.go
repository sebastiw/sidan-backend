package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateJWT(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long-12345678")
	memberNumber := int64(123)
	email := "test@example.com"
	scopes := []string{"write:email", "read:member"}
	provider := "google"

	// Generate JWT
	token, err := GenerateJWT(memberNumber, email, scopes, provider, secret)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Validate JWT
	claims, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	// Verify claims
	if claims.MemberNumber != memberNumber {
		t.Errorf("MemberNumber mismatch: got %d, want %d", claims.MemberNumber, memberNumber)
	}
	
	if claims.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", claims.Email, email)
	}
	
	if claims.Provider != provider {
		t.Errorf("Provider mismatch: got %s, want %s", claims.Provider, provider)
	}
	
	if len(claims.Scopes) != len(scopes) {
		t.Errorf("Scopes length mismatch: got %d, want %d", len(claims.Scopes), len(scopes))
	}
	
	// Verify expiry
	if time.Until(claims.ExpiresAt.Time) < 7*time.Hour {
		t.Error("Token expires too soon")
	}
}

func TestValidateJWT_InvalidSecret(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long-12345678")
	wrongSecret := []byte("wrong-secret-key-at-least-32-bytes-long-87654321")
	
	token, _ := GenerateJWT(123, "test@example.com", []string{}, "google", secret)
	
	_, err := ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("Expected error with wrong secret, got nil")
	}
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long-12345678")
	
	_, err := ValidateJWT("not.a.valid.jwt", secret)
	if err == nil {
		t.Error("Expected error with malformed token, got nil")
	}
}

func TestExtractBearer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Bearer abc123", "abc123"},
		{"Bearer ", ""},
		{"bearer abc123", ""},
		{"abc123", ""},
		{"", ""},
	}
	
	for _, tt := range tests {
		result := ExtractBearer(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractBearer(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
