package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
)

// NewAuthHandler creates auth handlers with database access
func NewAuthHandler(db data.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

type AuthHandler struct {
	db data.Database
}

// Login initiates OAuth2 flow
// GET /auth/login?provider=google&redirect_uri=https://...
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	redirectURI := r.URL.Query().Get("redirect_uri")
	
	if provider == "" {
		http.Error(w, "provider required", http.StatusBadRequest)
		return
	}
	
	// Get OAuth2 config for provider
	oauth2Cfg, exists := config.Get().OAuth2[provider]
	if !exists {
		http.Error(w, "unknown provider", http.StatusBadRequest)
		return
	}
	
	providerCfg, err := auth.GetProviderConfig(provider, oauth2Cfg.ClientID, 
		oauth2Cfg.ClientSecret, oauth2Cfg.RedirectURL, oauth2Cfg.Scopes)
	if err != nil {
		slog.Error("provider config failed", "provider", provider, "error", err)
		http.Error(w, "provider configuration error", http.StatusInternalServerError)
		return
	}
	
	// Generate PKCE
	verifier, err := auth.GeneratePKCEVerifier()
	if err != nil {
		slog.Error("PKCE verifier generation failed", "error", err)
		http.Error(w, "crypto error", http.StatusInternalServerError)
		return
	}
	challenge := auth.GeneratePKCEChallenge(verifier)
	
	// Generate state and nonce
	state := auth.GenerateState()
	nonce := auth.GenerateNonce()
	
	// Store state in database (10 min TTL)
	authState := &models.AuthState{
		ID:           state,
		Provider:     provider,
		Nonce:        nonce,
		PKCEVerifier: verifier,
		RedirectURI:  redirectURI,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}
	
	if err := h.db.CreateAuthState(authState); err != nil {
		slog.Error("failed to store auth state", "error", err)
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	
	// Build authorization URL
	authURL := providerCfg.GetAuthURL(state, challenge)
	
	slog.Info("oauth2 login initiated", "provider", provider, "state", state[:8]+"...")
	
	// Redirect to provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles OAuth2 callback
// GET /auth/callback?state=...&code=...
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateParam := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	
	if stateParam == "" || code == "" {
		http.Error(w, "missing state or code", http.StatusBadRequest)
		return
	}
	
	// Get and validate state from database
	authState, err := h.db.GetAuthState(stateParam)
	if err != nil {
		slog.Warn("invalid or expired state", "state", stateParam[:8]+"...", "error", err)
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}
	
	// Delete state (one-time use)
	h.db.DeleteAuthState(stateParam)
	
	// Get provider config
	oauth2Cfg := config.Get().OAuth2[authState.Provider]
	providerCfg, err := auth.GetProviderConfig(authState.Provider, oauth2Cfg.ClientID,
		oauth2Cfg.ClientSecret, oauth2Cfg.RedirectURL, oauth2Cfg.Scopes)
	if err != nil {
		http.Error(w, "provider configuration error", http.StatusInternalServerError)
		return
	}
	
	// Exchange code for token using PKCE verifier
	token, err := providerCfg.ExchangeCode(code, authState.PKCEVerifier)
	if err != nil {
		slog.Error("token exchange failed", "provider", authState.Provider, "error", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}
	
	// Get user info from provider
	userInfo, err := providerCfg.GetUserInfo(token.AccessToken)
	if err != nil {
		slog.Error("failed to get user info", "provider", authState.Provider, "error", err)
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	
	if !userInfo.EmailVerified {
		http.Error(w, "email not verified with provider", http.StatusForbidden)
		return
	}
	
	// Find member by email in cl2007_members table
	members, err := h.db.ReadMembers(true) // Get only valid members
	if err != nil {
		slog.Error("failed to read members", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	
	// Find member by email
	var member *models.Member
	for _, m := range members {
		if m.Email != nil && *m.Email == userInfo.Email {
			member = &m
			break
		}
	}
	
	if member == nil {
		slog.Warn("email not registered in members table", "provider", authState.Provider, "email", userInfo.Email)
		http.Error(w, "email not registered - please contact admin", http.StatusForbidden)
		return
	}
	
	slog.Info("member authenticated", "member_number", member.Number, "email", userInfo.Email, "provider", authState.Provider)

	// Determine scopes based on member type
	scopes := getScopesForMemberType(member)

	// Generate JWT token using member number
	jwtToken, err := auth.GenerateJWT(member.Number, userInfo.Email, scopes, authState.Provider, config.GetJWTSecret())
	if err != nil {
		slog.Error("JWT generation failed", "error", err)
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}
	
	slog.Info("login successful", "provider", authState.Provider, "member", member.Id, "email", userInfo.Email)

	// If redirect_uri was provided, redirect with token
	if authState.RedirectURI != "" {
		// Build token JSON
		tokenData := map[string]interface{}{
			"access_token": jwtToken,
			"token_type":   "Bearer",
			"expires_in":   28800, // 8 hours in seconds
			"member": map[string]interface{}{
				"number": member.Number,
				"email":  userInfo.Email,
				"name":   userInfo.Name,
			},
			"scopes": scopes,
		}

		// Encode as JSON
		tokenJSON, err := json.Marshal(tokenData)
		if err != nil {
			slog.Error("failed to encode token JSON", "error", err)
			http.Error(w, "encoding error", http.StatusInternalServerError)
			return
		}

		// URL-encode the JSON
		tokenParam := url.QueryEscape(string(tokenJSON))

		// Build redirect URL
		redirectURL := authState.RedirectURI + "?token=" + tokenParam

		slog.Info("redirecting to app", "redirect_uri", authState.RedirectURI)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// No redirect_uri - return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": jwtToken,
		"token_type":   "Bearer",
		"expires_in":   28800, // 8 hours in seconds
		"member": map[string]interface{}{
			"number": member.Number,
			"email":  userInfo.Email,
			"name":   userInfo.Name,
		},
		"scopes": scopes,
	})
}

// GetSession returns current JWT claims and member info
// GET /auth/session
// Authorization: Bearer <token>
func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Get claims from context (injected by RequireAuth middleware)
	claims := auth.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"no authentication"}`, http.StatusUnauthorized)
		return
	}
	
	member := auth.GetMember(r)
	if member == nil {
		http.Error(w, `{"error":"member not found"}`, http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"member": map[string]interface{}{
			"number": member.Number,
			"name":   member.Name,
			"email":  member.Email,
		},
		"scopes":     claims.Scopes,
		"provider":   claims.Provider,
		"expires_at": claims.ExpiresAt.Time,
		"issued_at":  claims.IssuedAt.Time,
	})
}

// Logout ends authentication (JWT remains valid until expiry)
// POST /auth/logout
// Authorization: Bearer <token>
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get claims from context
	claims := auth.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"no authentication"}`, http.StatusUnauthorized)
		return
	}
	
	slog.Info("logout successful", "member_number", claims.MemberNumber, "email", claims.Email)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Refresh generates a new JWT token
// POST /auth/refresh
// Authorization: Bearer <token>
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Get claims from context
	claims := auth.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"no authentication"}`, http.StatusUnauthorized)
		return
	}
	
	member := auth.GetMember(r)
	if member == nil {
		http.Error(w, `{"error":"member not found"}`, http.StatusInternalServerError)
		return
	}
	
	// Generate new JWT with same scopes
	newToken, err := auth.GenerateJWT(claims.MemberNumber, claims.Email, claims.Scopes, claims.Provider, config.GetJWTSecret())
	if err != nil {
		slog.Error("JWT refresh failed", "error", err)
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": newToken,
		"token_type":   "Bearer",
		"expires_in":   28800, // 8 hours
	})
}

// Helper functions

func getScopesForMemberType(member *models.Member) []string {
	// All valid members get basic scopes
	if member.Isvalid != nil && *member.Isvalid {
		return []string{"write:email", "write:image", "write:member", "read:member", "modify:entry", "write:arr"}
	}
	// Inactive members get limited access
	return []string{"read:member"}
}
