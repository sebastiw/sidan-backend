package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
)

// Scope constants for authorization
const (
	WriteEmailScope  = "write:email"
	WriteImageScope  = "write:image"
	WriteMemberScope = "write:member"
	ReadMemberScope  = "read:member"
	ModifyEntryScope = "modify:entry"
)

// Context keys for storing auth data in request context
type contextKey string

const (
	sessionKey contextKey = "session"
	memberKey  contextKey = "member"
)

// Middleware is a wrapper that provides auth functionality
type Middleware struct {
	db data.Database
}

// NewMiddleware creates auth middleware
func NewMiddleware(db data.Database) *Middleware {
	return &Middleware{db: db}
}

// RequireAuth validates session and injects member into context
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, `{"error":"no session"}`, http.StatusUnauthorized)
			return
		}

		// Validate session
		session, err := m.db.GetAuthSession(cookie.Value)
		if err != nil {
			http.Error(w, `{"error":"invalid session"}`, http.StatusUnauthorized)
			return
		}

		// Check expiry
		if time.Now().After(session.ExpiresAt) {
			m.db.DeleteAuthSession(session.ID)
			http.Error(w, `{"error":"session expired"}`, http.StatusUnauthorized)
			return
		}

		// Update last activity
		m.db.TouchAuthSession(session.ID)

		// Get member
		member, err := m.db.ReadMember(session.MemberID)
		if err != nil {
			http.Error(w, `{"error":"member not found"}`, http.StatusUnauthorized)
			return
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), sessionKey, session)
		ctx = context.WithValue(ctx, memberKey, member)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope checks if session has required scope
func (m *Middleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSession(r)
			if session == nil || session.Data == nil {
				http.Error(w, `{"error":"no session in context"}`, http.StatusUnauthorized)
				return
			}

			// Check if scope exists in session scopes
			hasScope := false
			for _, s := range session.Data.Scopes {
				if s == scope {
					hasScope = true
					break
				}
			}

			if !hasScope {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth tries to load session but doesn't fail if missing
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			// No session, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		session, err := m.db.GetAuthSession(cookie.Value)
		if err != nil || time.Now().After(session.ExpiresAt) {
			// Invalid/expired session, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Update last activity
		m.db.TouchAuthSession(session.ID)

		// Get member
		member, err := m.db.ReadMember(session.MemberID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), sessionKey, session)
		ctx = context.WithValue(ctx, memberKey, member)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSession retrieves session from request context
func GetSession(r *http.Request) *models.AuthSession {
	val := r.Context().Value(sessionKey)
	if val == nil {
		return nil
	}
	session, _ := val.(*models.AuthSession)
	return session
}

// GetMember retrieves member from request context
func GetMember(r *http.Request) *models.Member {
	val := r.Context().Value(memberKey)
	if val == nil {
		return nil
	}
	member, _ := val.(*models.Member)
	return member
}

// RefreshTokenIfNeeded checks if token is expiring soon and refreshes it
func (m *Middleware) RefreshTokenIfNeeded(memberID int64, provider string, crypto TokenCrypto) error {
	token, err := m.db.GetAuthToken(memberID, provider)
	if err != nil {
		return err
	}

	// If no expiry set, nothing to refresh
	if token.ExpiresAt == nil {
		return nil
	}

	// Refresh if less than 5 minutes remaining
	timeUntilExpiry := time.Until(*token.ExpiresAt)
	if timeUntilExpiry > 5*time.Minute {
		return nil // Still valid
	}

	// Decrypt refresh token
	if token.RefreshToken == nil {
		slog.Warn("token expiring but no refresh token available", 
			slog.Int64("member_id", memberID), 
			slog.String("provider", provider))
		return nil
	}

	refreshToken, err := crypto.Decrypt(*token.RefreshToken)
	if err != nil {
		return err
	}

	// Get provider OAuth2 config from app config
	oauth2Configs := config.Get().OAuth2
	providerCfg, ok := oauth2Configs[provider]
	if !ok {
		return err
	}

	// Build OAuth2 config for token refresh
	providerConfig, err := GetProviderConfig(provider, providerCfg.ClientID, providerCfg.ClientSecret, providerCfg.RedirectURL, providerCfg.Scopes)
	if err != nil {
		return err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: providerConfig.TokenURL,
		},
	}

	newToken, err := oauth2Config.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	}).Token()
	if err != nil {
		return err
	}

	// Encrypt and update
	encryptedAccess, err := crypto.Encrypt(newToken.AccessToken)
	if err != nil {
		return err
	}

	var encryptedRefresh *string
	if newToken.RefreshToken != "" {
		encrypted, err := crypto.Encrypt(newToken.RefreshToken)
		if err != nil {
			return err
		}
		encryptedRefresh = &encrypted
	}

	token.AccessToken = encryptedAccess
	token.RefreshToken = encryptedRefresh
	token.ExpiresAt = &newToken.Expiry
	err = m.db.UpdateAuthToken(token)
	if err != nil {
		return err
	}

	slog.Info("token refreshed successfully",
		slog.Int64("member_id", memberID),
		slog.String("provider", provider))

	return nil
}

// CleanupExpired removes expired sessions and states
func CleanupExpired(db data.Database) error {
	// Delete expired sessions
	if err := db.CleanupExpiredAuthSessions(); err != nil {
		slog.Error("failed to delete expired sessions", slog.String("error", err.Error()))
		return err
	}

	// Delete expired states
	if err := db.CleanupExpiredAuthStates(); err != nil {
		slog.Error("failed to delete expired states", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("cleaned up expired auth data")
	return nil
}

// StartCleanupJob runs cleanup in background
func StartCleanupJob(db data.Database, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			CleanupExpired(db)
		}
	}()
	slog.Info("cleanup job started", slog.Duration("interval", interval))
}
