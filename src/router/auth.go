package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
)

// NewAuthHandler creates auth handlers with database access
func NewAuthHandler(db data.Database, crypto auth.TokenCrypto) *AuthHandler {
	return &AuthHandler{db: db, crypto: crypto}
}

type AuthHandler struct {
	db     data.Database
	crypto auth.TokenCrypto
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
	
	// First, try to find existing member by provider link
	member, err := h.db.GetMemberByProviderEmail(authState.Provider, userInfo.Email)
	if err != nil {
		// No provider link exists yet - check if member exists by email in members table
		slog.Info("no provider link found, checking member table", "email", userInfo.Email)
		
		// Query members table directly
		members, err := h.db.ReadMembers(true) // Get only valid members
		if err != nil {
			slog.Error("failed to read members", "error", err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		
		// Find member by email
		var foundMember *models.Member
		for _, m := range members {
			if m.Email != nil && *m.Email == userInfo.Email {
				foundMember = &m
				break
			}
		}
		
		if foundMember == nil {
			slog.Warn("email not registered in members table", "provider", authState.Provider, "email", userInfo.Email)
			http.Error(w, "email not registered - please contact admin", http.StatusForbidden)
			return
		}
		
		member = foundMember
		slog.Info("member found by email", "member_id", member.Id, "email", userInfo.Email)
	}
	
	// Encrypt and store token
	encryptedAccess, err := h.crypto.Encrypt(token.AccessToken)
	if err != nil {
		slog.Error("failed to encrypt access token", "error", err)
		http.Error(w, "encryption error", http.StatusInternalServerError)
		return
	}
	
	var encryptedRefresh *string
	if token.RefreshToken != "" {
		encrypted, err := h.crypto.Encrypt(token.RefreshToken)
		if err != nil {
			slog.Error("failed to encrypt refresh token", "error", err)
			http.Error(w, "encryption error", http.StatusInternalServerError)
			return
		}
		encryptedRefresh = &encrypted
	}
	
	authToken := &models.AuthToken{
		MemberID:     member.Id,
		Provider:     authState.Provider,
		AccessToken:  encryptedAccess,
		RefreshToken: encryptedRefresh,
		TokenType:    token.TokenType,
		ExpiresAt:    &token.Expiry,
		Scopes:       oauth2Cfg.Scopes,
	}
	
	// Create or update token
	existing, _ := h.db.GetAuthToken(member.Id, authState.Provider)
	if existing != nil {
		authToken.ID = existing.ID
		h.db.UpdateAuthToken(authToken)
	} else {
		h.db.CreateAuthToken(authToken)
	}
	
	// Create or update provider link
	link, _ := h.db.GetAuthProviderLink(authState.Provider, userInfo.ProviderUserID)
	if link == nil {
		h.db.CreateAuthProviderLink(&models.AuthProviderLink{
			MemberID:       member.Id,
			Provider:       authState.Provider,
			ProviderUserID: userInfo.ProviderUserID,
			ProviderEmail:  userInfo.Email,
			EmailVerified:  userInfo.EmailVerified,
			LinkedAt:       time.Now(),
		})
	}
	
	// Determine scopes based on member type
	scopes := getScopesForMemberType(member)
	
	// Create session
	now := time.Now()
	sessionID := auth.GenerateState() // Reuse state generator for session ID
	session := &models.AuthSession{
		ID:           sessionID,
		MemberID:     member.Id,
		Data:         &models.SessionData{
			Scopes:   scopes,
			Provider: authState.Provider,
		},
		CreatedAt:    now,
		ExpiresAt:    now.Add(8 * time.Hour),
		LastActivity: now,
	}
	
	if err := h.db.CreateAuthSession(session); err != nil {
		slog.Error("failed to create session", "error", err)
		http.Error(w, "session creation failed", http.StatusInternalServerError)
		return
	}
	
	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   8 * 60 * 60, // 8 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
	
	slog.Info("login successful", "provider", authState.Provider, "member", member.Id, "email", userInfo.Email)
	
	// Redirect to client or show success page
	if authState.RedirectURI != "" {
		http.Redirect(w, r, authState.RedirectURI, http.StatusTemporaryRedirect)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"member":  map[string]interface{}{
				"id":    member.Id,
				"email": userInfo.Email,
				"name":  userInfo.Name,
			},
		})
	}
}

// GetSession returns current session info
// GET /auth/session
func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionIDFromRequest(r)
	if err != nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	
	session, err := h.db.GetAuthSession(sessionID)
	if err != nil {
		http.Error(w, "session not found or expired", http.StatusUnauthorized)
		return
	}
	
	// Update last activity
	h.db.TouchAuthSession(sessionID)
	
	// Get member info
	member, err := h.db.ReadMember(session.MemberID)
	if err != nil {
		http.Error(w, "member not found", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"member": map[string]interface{}{
			"id":     member.Id,
			"number": member.Number,
			"name":   member.Name,
			"email":  member.Email,
		},
		"scopes":     session.Data.Scopes,
		"provider":   session.Data.Provider,
		"expires_at": session.ExpiresAt,
	})
}

// Logout ends the session
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionIDFromRequest(r)
	if err != nil {
		http.Error(w, "no session", http.StatusBadRequest)
		return
	}
	
	if err := h.db.DeleteAuthSession(sessionID); err != nil {
		slog.Warn("failed to delete session", "session", sessionID[:8]+"...", "error", err)
	}
	
	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Refresh refreshes OAuth2 access token
// POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionIDFromRequest(r)
	if err != nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	
	session, err := h.db.GetAuthSession(sessionID)
	if err != nil {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}
	
	if session.Data == nil {
		http.Error(w, "invalid session data", http.StatusInternalServerError)
		return
	}
	
	// Create middleware to use refresh logic
	middleware := auth.NewMiddleware(h.db)
	err = middleware.RefreshTokenIfNeeded(session.MemberID, session.Data.Provider, h.crypto)
	if err != nil {
		slog.Error("token refresh failed", 
			slog.Int64("member_id", session.MemberID),
			slog.String("provider", session.Data.Provider),
			slog.String("error", err.Error()))
		http.Error(w, "refresh failed", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Helper functions

func getSessionIDFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func getScopesForMemberType(member *models.Member) []string {
	// All valid members get basic scopes
	if member.Isvalid != nil && *member.Isvalid {
		return []string{"write:email", "write:image", "write:member", "read:member", "modify:entry"}
	}
	// Inactive members get limited access
	return []string{"read:member"}
}
