package auth

import (
	"strings"
	"testing"
)

func TestGeneratePKCEVerifier(t *testing.T) {
	verifier, err := GeneratePKCEVerifier()
	if err != nil {
		t.Fatalf("Failed to generate verifier: %v", err)
	}
	
	// Should be 43 characters (32 bytes base64url encoded)
	if len(verifier) != 43 {
		t.Errorf("Expected length 43, got %d", len(verifier))
	}
	
	// Should be URL-safe (no padding, no + or /)
	if strings.Contains(verifier, "+") || strings.Contains(verifier, "/") || strings.Contains(verifier, "=") {
		t.Error("Verifier should be URL-safe base64 (no +, /, or =)")
	}
	
	// Should be unique
	verifier2, _ := GeneratePKCEVerifier()
	if verifier == verifier2 {
		t.Error("Verifiers should be unique")
	}
}

func TestGeneratePKCEChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GeneratePKCEChallenge(verifier)
	
	// Known test vector from RFC 7636
	expected := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if challenge != expected {
		t.Errorf("Expected %s, got %s", expected, challenge)
	}
	
	// Should be 43 characters
	if len(challenge) != 43 {
		t.Errorf("Expected length 43, got %d", len(challenge))
	}
}

func TestGenerateState(t *testing.T) {
	state := GenerateState()
	
	// Should be 64 hex characters (32 bytes)
	if len(state) != 64 {
		t.Errorf("Expected length 64, got %d", len(state))
	}
	
	// Should be unique
	state2 := GenerateState()
	if state == state2 {
		t.Error("States should be unique")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce := GenerateNonce()
	
	// Should be 64 hex characters (32 bytes)
	if len(nonce) != 64 {
		t.Errorf("Expected length 64, got %d", len(nonce))
	}
	
	// Should be unique
	nonce2 := GenerateNonce()
	if nonce == nonce2 {
		t.Error("Nonces should be unique")
	}
}
