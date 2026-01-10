package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

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
	claimsKey contextKey = "claims"
	memberKey contextKey = "member"
	scopesKey contextKey = "scopes"
)

// Middleware is a wrapper that provides auth functionality
type Middleware struct {
	db data.Database
}

// NewMiddleware creates auth middleware
func NewMiddleware(db data.Database) *Middleware {
	return &Middleware{db: db}
}

// RequireAuth validates JWT Bearer token and injects member into context
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		token := ExtractBearer(authHeader)
		if token == "" {
			http.Error(w, `{"error":"invalid authorization format, use: Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		// Validate JWT
		claims, err := ValidateJWT(token, config.GetJWTSecret())
		if err != nil {
			if err == ErrExpiredToken {
				http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
			} else {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			}
			return
		}

		// Get member from database
		member, err := m.db.ReadMember(claims.MemberID)
		if err != nil {
			http.Error(w, `{"error":"member not found"}`, http.StatusUnauthorized)
			return
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		ctx = context.WithValue(ctx, memberKey, member)
		ctx = context.WithValue(ctx, scopesKey, claims.Scopes)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope checks if JWT has required scope
func (m *Middleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes := GetScopes(r)
			if scopes == nil {
				http.Error(w, `{"error":"no scopes in context"}`, http.StatusUnauthorized)
				return
			}

			// Check if scope exists
			hasScope := false
			for _, s := range scopes {
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

// OptionalAuth tries to load JWT but doesn't fail if missing
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth, continue without it
			next.ServeHTTP(w, r)
			return
		}

		token := ExtractBearer(authHeader)
		if token == "" {
			// Invalid format, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		claims, err := ValidateJWT(token, config.GetJWTSecret())
		if err != nil {
			// Invalid token, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Get member
		member, err := m.db.ReadMember(claims.MemberID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		ctx = context.WithValue(ctx, memberKey, member)
		ctx = context.WithValue(ctx, scopesKey, claims.Scopes)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaims retrieves JWT claims from request context
func GetClaims(r *http.Request) *JWTClaims {
	val := r.Context().Value(claimsKey)
	if val == nil {
		return nil
	}
	claims, _ := val.(*JWTClaims)
	return claims
}

// GetScopes retrieves scopes from request context
func GetScopes(r *http.Request) []string {
	val := r.Context().Value(scopesKey)
	if val == nil {
		return nil
	}
	scopes, _ := val.([]string)
	return scopes
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

// CleanupExpired removes expired auth states
func CleanupExpired(db data.Database) error {
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
