package auth

import (
	"strings"
	"testing"
)

func TestGetProviderConfig_Google(t *testing.T) {
	cfg, err := GetProviderConfig("google", "client123", "secret123", "http://localhost/callback", []string{"email"})
	
	if err != nil {
		t.Fatalf("Failed to get Google config: %v", err)
	}
	
	if cfg.Name != "google" {
		t.Errorf("Expected name 'google', got '%s'", cfg.Name)
	}
	
	if cfg.ClientID != "client123" {
		t.Errorf("Expected clientID 'client123', got '%s'", cfg.ClientID)
	}
	
	if !strings.Contains(cfg.AuthURL, "google") {
		t.Errorf("Expected Google auth URL, got '%s'", cfg.AuthURL)
	}
	
	if cfg.UserInfoURL == "" {
		t.Error("UserInfoURL should not be empty")
	}
}

func TestGetProviderConfig_GitHub(t *testing.T) {
	cfg, err := GetProviderConfig("github", "client456", "secret456", "http://localhost/callback", []string{"user:email"})
	
	if err != nil {
		t.Fatalf("Failed to get GitHub config: %v", err)
	}
	
	if cfg.Name != "github" {
		t.Errorf("Expected name 'github', got '%s'", cfg.Name)
	}
	
	if !strings.Contains(cfg.AuthURL, "github") {
		t.Errorf("Expected GitHub auth URL, got '%s'", cfg.AuthURL)
	}
}

func TestGetProviderConfig_Unknown(t *testing.T) {
	_, err := GetProviderConfig("facebook", "client", "secret", "http://localhost", []string{})
	
	if err == nil {
		t.Error("Expected error for unknown provider")
	}
	
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Errorf("Expected 'unsupported provider' error, got: %v", err)
	}
}

func TestGetAuthURL(t *testing.T) {
	cfg := &ProviderConfig{
		Name:        "google",
		ClientID:    "test-client",
		RedirectURL: "http://localhost/callback",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		Scopes:      []string{"email", "profile"},
	}
	
	authURL := cfg.GetAuthURL("test-state", "test-challenge")
	
	// Should contain all required parameters
	if !strings.Contains(authURL, "client_id=test-client") {
		t.Error("Auth URL missing client_id")
	}
	
	if !strings.Contains(authURL, "state=test-state") {
		t.Error("Auth URL missing state")
	}
	
	if !strings.Contains(authURL, "code_challenge=test-challenge") {
		t.Error("Auth URL missing code_challenge")
	}
	
	if !strings.Contains(authURL, "code_challenge_method=S256") {
		t.Error("Auth URL missing code_challenge_method")
	}
	
	if !strings.Contains(authURL, "response_type=code") {
		t.Error("Auth URL missing response_type")
	}
	
	// Google-specific
	if !strings.Contains(authURL, "access_type=offline") {
		t.Error("Google auth URL missing access_type=offline")
	}
}

func TestGetAuthURL_GitHub(t *testing.T) {
	cfg := &ProviderConfig{
		Name:        "github",
		ClientID:    "test-client",
		RedirectURL: "http://localhost/callback",
		AuthURL:     "https://github.com/login/oauth/authorize",
		Scopes:      []string{"user:email"},
	}
	
	authURL := cfg.GetAuthURL("test-state", "test-challenge")
	
	// Should NOT contain Google-specific params
	if strings.Contains(authURL, "access_type") {
		t.Error("GitHub auth URL should not have access_type")
	}
	
	// Should have standard OAuth2 params
	if !strings.Contains(authURL, "client_id") {
		t.Error("Auth URL missing client_id")
	}
}

func TestParseGoogleUserInfo(t *testing.T) {
	jsonData := []byte(`{
		"id": "123456789",
		"email": "user@example.com",
		"verified_email": true,
		"name": "John Doe",
		"picture": "https://example.com/photo.jpg"
	}`)
	
	info, err := parseGoogleUserInfo(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse Google user info: %v", err)
	}
	
	if info.ProviderUserID != "123456789" {
		t.Errorf("Expected ID '123456789', got '%s'", info.ProviderUserID)
	}
	
	if info.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", info.Email)
	}
	
	if !info.EmailVerified {
		t.Error("Expected email to be verified")
	}
	
	if info.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", info.Name)
	}
}

func TestParseGoogleUserInfo_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid json}`)
	
	_, err := parseGoogleUserInfo(jsonData)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
