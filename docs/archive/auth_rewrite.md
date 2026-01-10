# Authentication System Rewrite

## Current State Analysis

**Problems**:
- Tokens stored directly in sessions (no database persistence across restarts)
- No token refresh mechanism
- State parameter only validated once, not stored properly
- Mixed concerns: OAuth2 flow, session management, and user lookup in single handlers
- No PKCE support (security risk for public clients)
- No logout mechanism
- Session keys regenerated on restart, invalidating all sessions
- Custom "Sidan" provider implementation is half-baked OAuth2 simulation
- No audit trail of authentication events
- Scopes hardcoded by user type, not configurable
- No provider account unlinking
- Token expiry checking uses hardcoded 1998 date hack

**What Works**:
- Email verification against member database
- Multiple OAuth2 provider support
- Scope-based authorization middleware
- CORS configuration

## Goals

1. Production-ready OAuth2 implementation with proper token lifecycle
2. Extensible provider system (easy to add new providers)
3. Member authentication via verified email from OAuth2 providers
4. Persistent sessions across server restarts
5. Security best practices (PKCE, CSRF protection, secure token storage)
6. Remove legacy password authentication system entirely

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ 1. GET /auth/login?provider=google
       ▼
┌─────────────────┐
│  Auth Service   │─────┐ 2. Redirect to provider
│  (Go Handlers)  │     │
└────────┬────────┘     │
         │ 5. Callback  │
         ▼              ▼
┌─────────────────┐  ┌──────────────┐
│  Token Manager  │  │   Provider   │
│  (persistence)  │  │ (Google/GH)  │
└────────┬────────┘  └──────────────┘
         │ 6. Store token
         ▼
┌─────────────────┐
│   Database      │
│ - auth_states   │
│ - auth_tokens   │
│ - auth_links    │
└─────────────────┘
```

## Phase 1: Database Schema & Token Storage (Week 1)

**Goal**: Persistent token storage and state management

### Database Tables

```sql
-- OAuth2 state tracking (CSRF protection)
CREATE TABLE auth_states (
    id VARCHAR(64) PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    nonce VARCHAR(64) NOT NULL,
    pkce_verifier VARCHAR(128),
    redirect_uri TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    INDEX idx_expires (expires_at)
);

-- OAuth2 tokens (encrypted at rest)
CREATE TABLE auth_tokens (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    provider VARCHAR(32) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type VARCHAR(32) DEFAULT 'Bearer',
    expires_at TIMESTAMP,
    scopes JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES cl2007_members(id) ON DELETE CASCADE,
    UNIQUE KEY unique_member_provider (member_id, provider),
    INDEX idx_expires (expires_at)
);

-- Provider account linking (which emails belong to which member)
CREATE TABLE auth_provider_links (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    provider VARCHAR(32) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    linked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES cl2007_members(id) ON DELETE CASCADE,
    UNIQUE KEY unique_provider_user (provider, provider_user_id),
    INDEX idx_member (member_id),
    INDEX idx_provider_email (provider, provider_email)
);

-- Session management (stateless JWT or DB-backed)
CREATE TABLE auth_sessions (
    id VARCHAR(128) PRIMARY KEY,
    member_id BIGINT NOT NULL,
    data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES cl2007_members(id) ON DELETE CASCADE,
    INDEX idx_member (member_id),
    INDEX idx_expires (expires_at)
);
```

### Go Models

```go
// src/models/auth.go
type AuthState struct {
    ID            string    `gorm:"primaryKey"`
    Provider      string    `gorm:"size:32;not null"`
    Nonce         string    `gorm:"size:64;not null"`
    PKCEVerifier  string    `gorm:"size:128"`
    RedirectURI   string    `gorm:"type:text"`
    CreatedAt     time.Time
    ExpiresAt     time.Time `gorm:"not null;index"`
}

type AuthToken struct {
    ID           int64     `gorm:"primaryKey"`
    MemberID     int64     `gorm:"not null;uniqueIndex:unique_member_provider"`
    Provider     string    `gorm:"size:32;not null;uniqueIndex:unique_member_provider"`
    AccessToken  string    `gorm:"type:text;not null"` // encrypted
    RefreshToken *string   `gorm:"type:text"`          // encrypted
    TokenType    string    `gorm:"size:32;default:Bearer"`
    ExpiresAt    *time.Time `gorm:"index"`
    Scopes       []string  `gorm:"type:json"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type AuthProviderLink struct {
    ID              int64     `gorm:"primaryKey"`
    MemberID        int64     `gorm:"not null;index"`
    Provider        string    `gorm:"size:32;not null;uniqueIndex:unique_provider_user"`
    ProviderUserID  string    `gorm:"size:255;not null;uniqueIndex:unique_provider_user"`
    ProviderEmail   string    `gorm:"size:255;not null;index:idx_provider_email"`
    EmailVerified   bool      `gorm:"default:false"`
    LinkedAt        time.Time
}

type AuthSession struct {
    ID           string    `gorm:"primaryKey"`
    MemberID     int64     `gorm:"not null;index"`
    Data         string    `gorm:"type:json"`
    CreatedAt    time.Time
    ExpiresAt    time.Time `gorm:"not null;index"`
    LastActivity time.Time
}
```

### Token Encryption

```go
// src/auth/crypto.go
// Use AES-256-GCM for token encryption at rest
// Key derived from environment variable AUTH_ENCRYPTION_KEY
type TokenCrypto interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}
```

**Deliverables**:
- Migration SQL file: `db/2026-01-10-auth-tables.sql`
- Models: `src/models/auth.go`
- Token encryption: `src/auth/crypto.go`
- Database operations: `src/data/commondb/auth.go`

---

## Phase 2: Provider Abstraction Layer (Week 1-2)

**Goal**: Clean, extensible provider interface

### Provider Interface

```go
// src/auth/provider/provider.go
type Provider interface {
    Name() string
    GetAuthURL(state, nonce, pkceChallenge string) string
    GetTokenURL() string
    ExchangeToken(code, pkceVerifier string) (*oauth2.Token, error)
    GetUserInfo(token *oauth2.Token) (*UserInfo, error)
    RefreshToken(refreshToken string) (*oauth2.Token, error)
    SupportsRefresh() bool
}

type UserInfo struct {
    ProviderUserID string
    Email          string
    EmailVerified  bool
    Name           string
    Picture        string
    AdditionalEmails []string // for GitHub multiple emails
}

type ProviderConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}
```

### Concrete Providers

```go
// src/auth/provider/google.go
type GoogleProvider struct {
    config ProviderConfig
    oauth  *oauth2.Config
}

// src/auth/provider/github.go
type GithubProvider struct {
    config ProviderConfig
    oauth  *oauth2.Config
}

// Future: Microsoft, Auth0, etc.
```

### Provider Registry

```go
// src/auth/provider/registry.go
type Registry struct {
    providers map[string]Provider
}

func NewRegistry(cfg *config.Configuration) *Registry {
    r := &Registry{providers: make(map[string]Provider)}
    
    for name, providerCfg := range cfg.OAuth2 {
        switch name {
        case "google":
            r.Register(NewGoogleProvider(providerCfg))
        case "github":
            r.Register(NewGithubProvider(providerCfg))
        }
    }
    return r
}

func (r *Registry) Get(name string) (Provider, error)
func (r *Registry) List() []string
func (r *Registry) Register(provider Provider)
```

**Deliverables**:
- Provider interface: `src/auth/provider/provider.go`
- Registry: `src/auth/provider/registry.go`
- Google impl: `src/auth/provider/google.go`
- GitHub impl: `src/auth/provider/github.go`

---

## Phase 3: OAuth2 Flow Handlers (Week 2)

**Goal**: Secure, standards-compliant OAuth2 implementation

### Service Layer

```go
// src/auth/service.go
type Service struct {
    db          data.Database
    providers   *provider.Registry
    crypto      TokenCrypto
    sessionTTL  time.Duration
}

func (s *Service) InitiateAuth(providerName, redirectURI string) (*AuthInitiation, error)
func (s *Service) HandleCallback(state, code string) (*AuthResult, error)
func (s *Service) RefreshToken(memberID int64, providerName string) error
func (s *Service) RevokeToken(memberID int64, providerName string) error
func (s *Service) GetMemberFromSession(sessionID string) (*models.Member, error)
func (s *Service) ValidateSession(sessionID string) (bool, error)
func (s *Service) Logout(sessionID string) error
```

### HTTP Handlers

```go
// src/router/auth.go (new file, replaces old auth handlers)
type AuthHandler struct {
    service *auth.Service
}

// GET /auth/login?provider=google&redirect_uri=https://...
func (h *AuthHandler) InitiateLogin(w http.ResponseWriter, r *http.Request)

// GET /auth/callback?state=...&code=...
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request)

// POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request)

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request)

// GET /auth/session
func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request)

// GET /auth/providers
func (h *AuthHandler) ListProviders(w http.ResponseWriter, r *http.Request)
```

### PKCE Implementation

```go
// src/auth/pkce.go
func GenerateCodeVerifier() string
func GenerateCodeChallenge(verifier string) string
```

### Flow Steps

1. **Initiate** (`/auth/login`):
   - Generate state (CSRF token) + nonce
   - Generate PKCE verifier + challenge
   - Store in `auth_states` with 10min TTL
   - Build provider auth URL with PKCE challenge
   - Redirect to provider

2. **Callback** (`/auth/callback`):
   - Validate state exists and not expired
   - Exchange code for token using PKCE verifier
   - Fetch user info from provider
   - Find or create member by verified email
   - Create/update `auth_provider_links`
   - Encrypt and store tokens in `auth_tokens`
   - Create session in `auth_sessions`
   - Return session cookie (HttpOnly, Secure, SameSite=Lax)

3. **Refresh** (`/auth/refresh`):
   - Validate session
   - Fetch refresh token from DB
   - Call provider refresh endpoint
   - Update stored tokens
   - Extend session

**Deliverables**:
- Service layer: `src/auth/service.go`
- HTTP handlers: `src/router/auth.go`
- PKCE utils: `src/auth/pkce.go`
- Middleware update: `src/auth/middleware.go`

---

## Phase 4: Middleware & Session Management (Week 2-3)

**Goal**: Clean auth middleware for protected endpoints

### Session Strategy

Use JWT for stateless sessions with DB fallback for revocation:

```go
// src/auth/session.go
type SessionManager interface {
    Create(memberID int64, scopes []string) (*Session, error)
    Validate(sessionID string) (*Session, error)
    Refresh(sessionID string) error
    Revoke(sessionID string) error
}

type JWTSessionManager struct {
    signingKey []byte
    db         data.Database
}

type Session struct {
    ID        string
    MemberID  int64
    Scopes    []string
    ExpiresAt time.Time
}
```

### Middleware

```go
// src/auth/middleware.go (rewrite existing)
func RequireAuth(next http.Handler) http.Handler
func RequireScope(scope string) func(http.Handler) http.Handler
func RequireAnyScope(scopes ...string) func(http.Handler) http.Handler
func OptionalAuth(next http.Handler) http.Handler // for dual authed/unauthed endpoints
```

### Context Injection

```go
// src/auth/context.go
type contextKey string

const (
    SessionKey contextKey = "session"
    MemberKey  contextKey = "member"
)

func GetSession(r *http.Request) (*Session, error)
func GetMember(r *http.Request) (*models.Member, error)
func SetSession(r *http.Request, session *Session) *http.Request
```

**Deliverables**:
- Session manager: `src/auth/session.go`
- Middleware: `src/auth/middleware.go`
- Context utils: `src/auth/context.go`

---

## Phase 5: Migration & Cleanup (Week 3)

**Goal**: Remove old code, migrate existing users

### Migration Strategy

```go
// src/auth/migration/migrate.go
// One-time migration script (can be run via CLI flag)
func MigrateExistingSessions(db data.Database) error {
    // Since old sessions are in memory, this is a no-op
    // Just document that users need to re-login after upgrade
}

func MigratePasswordHashesToProviderLinks(db data.Database) error {
    // Optional: if keeping password field for reference
    // Mark which members need to link OAuth2 accounts
}
```

### Deletion Checklist

Remove files:
- `src/auth/auth.go` → replaced by `src/auth/service.go`
- `src/auth/auth_handlers.go` → replaced by `src/router/auth.go`
- `src/auth/sidan_auth_handler.go` → deleted (no custom provider)
- `src/auth/login.html` → deleted
- `src/auth/close.html` → deleted

Update files:
- `src/router/router.go` - replace auth route registration
- `src/config/config.go` - remove OAuth2 map, use structured provider configs
- `src/models/members.go` - deprecate password fields (don't delete for audit)
- `config/local.yaml` - update OAuth2 config structure

### Database Cleanup

```sql
-- DO NOT DELETE password fields (audit trail)
-- Just mark as deprecated in comments
ALTER TABLE cl2007_members 
    MODIFY COLUMN password VARCHAR(255) COMMENT 'DEPRECATED: Use auth_tokens table',
    MODIFY COLUMN password_classic VARCHAR(255) COMMENT 'DEPRECATED: Use auth_tokens table';
```

**Deliverables**:
- Migration script: `src/auth/migration/migrate.go`
- Updated router: `src/router/router.go`
- Cleanup documentation: `docs/auth_migration_guide.md`

---

## Phase 6: Testing & Documentation (Week 3-4)

**Goal**: Production readiness

### Test Coverage

```go
// src/auth/service_test.go
- TestInitiateAuth_ValidProvider
- TestInitiateAuth_InvalidProvider
- TestHandleCallback_ValidState
- TestHandleCallback_ExpiredState
- TestHandleCallback_InvalidState
- TestRefreshToken_Success
- TestRefreshToken_NoRefreshToken

// src/auth/provider/google_test.go
- TestGoogleProvider_GetUserInfo
- TestGoogleProvider_RefreshToken

// src/auth/middleware_test.go
- TestRequireAuth_ValidSession
- TestRequireAuth_ExpiredSession
- TestRequireScope_HasScope
- TestRequireScope_MissingScope
```

### Integration Tests

```go
// src/auth/integration_test.go
- TestFullOAuth2Flow_Google
- TestFullOAuth2Flow_GitHub
- TestTokenRefresh_BeforeExpiry
- TestLogout_RevokesSession
- TestMultipleProviderLinks_SameMember
```

### Documentation

1. **API Documentation** (`docs/api/auth.md`):
   - All endpoint specs with examples
   - Error codes and meanings
   - Rate limiting (if implemented)

2. **Developer Guide** (`docs/auth_developer_guide.md`):
   - How to add a new provider
   - How to test auth flows locally
   - How to configure OAuth2 apps

3. **User Migration Guide** (`docs/auth_migration_guide.md`):
   - Breaking changes
   - User impact (need to re-login)
   - Rollback plan

4. **Security Review** (`docs/auth_security.md`):
   - Threat model
   - Security controls
   - Known limitations

**Deliverables**:
- Test suite with >80% coverage
- Complete API documentation
- Developer guide
- Security documentation

---

## Phase 7: Optional Enhancements (Week 4+)

### 7.1 Token Rotation
- Implement refresh token rotation per OAuth2 best practices
- Automatic token refresh before expiry

### 7.2 Rate Limiting
- Per-IP rate limiting on auth endpoints
- Exponential backoff on failed attempts

### 7.3 Audit Logging
```sql
CREATE TABLE auth_audit_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT,
    event_type VARCHAR(50) NOT NULL, -- login, logout, refresh, link, unlink
    provider VARCHAR(32),
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN,
    details JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_member (member_id),
    INDEX idx_created (created_at)
);
```

### 7.4 Multi-Factor Authentication (MFA)
- TOTP support
- Backup codes
- Recovery flow

### 7.5 Admin Dashboard
- View linked accounts per member
- Revoke tokens
- View audit logs

---

## Security Considerations

### Implemented
- PKCE for all OAuth2 flows (protects against code interception)
- State parameter for CSRF protection
- Tokens encrypted at rest (AES-256-GCM)
- HttpOnly, Secure, SameSite cookies
- Token expiry validation
- HTTPS required in production
- Session TTL with activity tracking

### Configuration Requirements
```yaml
# config/production.yaml
auth:
  encryption_key: ${AUTH_ENCRYPTION_KEY} # 32-byte hex, from env
  session_ttl: 8h
  state_ttl: 10m
  cookie_secure: true
  cookie_domain: ".chalmerslosers.com"
  
  providers:
    google:
      client_id: ${GOOGLE_CLIENT_ID}
      client_secret: ${GOOGLE_CLIENT_SECRET}
      redirect_url: "https://api.chalmerslosers.com/auth/callback"
      scopes: ["openid", "email", "profile"]
    github:
      client_id: ${GITHUB_CLIENT_ID}
      client_secret: ${GITHUB_CLIENT_SECRET}
      redirect_url: "https://api.chalmerslosers.com/auth/callback"
      scopes: ["user:email"]
```

### Environment Variables
```bash
AUTH_ENCRYPTION_KEY=<64-char-hex>  # openssl rand -hex 32
GOOGLE_CLIENT_ID=<from-gcp-console>
GOOGLE_CLIENT_SECRET=<from-gcp-console>
GITHUB_CLIENT_ID=<from-github-oauth-app>
GITHUB_CLIENT_SECRET=<from-github-oauth-app>
```

---

## API Changes

### Removed Endpoints
- `GET /auth/{provider}` → `GET /auth/login?provider={provider}`
- `GET /auth/{provider}/authorized` → `GET /auth/callback` (single callback for all)
- `GET /auth/{provider}/verifyemail` → removed (done automatically in callback)
- `GET /auth/getusersession` → `GET /auth/session`
- `GET /login/oauth/authorize` → removed
- `POST /login` → removed
- `POST /login/oauth/access_token` → removed

### New Endpoints
- `GET /auth/login?provider={provider}&redirect_uri={uri}` - Initiate OAuth2 flow
- `GET /auth/callback?state={state}&code={code}` - OAuth2 callback (all providers)
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - Terminate session
- `GET /auth/session` - Get current session info
- `GET /auth/providers` - List available providers

### Changed Response Formats

**Before** (`GET /auth/getusersession`):
```json
{
  "scopes": ["write:email", "write:image"],
  "username": "#1234",
  "email": "user@example.com",
  "fulHaxPass": "..."
}
```

**After** (`GET /auth/session`):
```json
{
  "session_id": "sess_...",
  "member": {
    "id": 1234,
    "number": 1234,
    "email": "user@example.com",
    "name": "John Doe"
  },
  "scopes": ["write:email", "write:image", "write:member"],
  "provider": "google",
  "expires_at": "2026-01-10T20:00:00Z"
}
```

---

## Rollout Plan

### Development
1. Feature branch: `feature/auth-rewrite`
2. Keep old auth code alongside new (use feature flag)
3. Test thoroughly with both systems running

### Staging
1. Deploy with feature flag OFF
2. Run migration to create new tables
3. Enable feature flag for 10% of traffic
4. Monitor logs and errors
5. Gradually increase to 100%

### Production
1. **Pre-deployment**:
   - Announce maintenance window to users
   - Notify that re-login will be required
   - Backup database

2. **Deployment**:
   - Deploy new code with feature flag OFF
   - Run database migrations
   - Verify migrations successful
   - Enable feature flag
   - Monitor for 1 hour

3. **Post-deployment**:
   - Watch error rates and auth success rates
   - Keep old code for 2 weeks for emergency rollback
   - After stable, remove old auth code in separate PR

### Rollback Plan
If critical issues:
1. Disable feature flag (reverts to old auth)
2. Users keep existing sessions
3. New logins use old system
4. Fix issues, redeploy, re-enable flag

---

## Success Metrics

- Auth success rate > 99%
- P95 latency < 500ms for auth flow
- Zero token leaks
- Zero CSRF attacks
- Session validity across restarts
- Token refresh success rate > 95%
- Zero unauthorized access events

---

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| 1. Database Schema | 2 days | Tables, models, crypto |
| 2. Provider Layer | 3 days | Interface, Google, GitHub |
| 3. OAuth2 Handlers | 4 days | Service, handlers, PKCE |
| 4. Middleware | 2 days | Auth middleware, sessions |
| 5. Migration | 2 days | Remove old code, migrate |
| 6. Testing | 5 days | Unit + integration tests |
| 7. Optional | Ongoing | Enhancements |

**Total**: 3-4 weeks for phases 1-6

---

## Open Questions

1. **Token storage**: Encrypt individual fields or entire JSON? 
   - **Decision**: Encrypt only access_token and refresh_token fields

2. **Member auto-creation**: Create member on first OAuth2 login if email not found?
   - **Decision**: No, return error. Require explicit member registration first

3. **Multiple emails**: How to handle GitHub users with multiple verified emails?
   - **Decision**: Link all verified emails, use primary for display

4. **Scope assignment**: Continue hardcoding by member type or make configurable?
   - **Decision**: Keep hardcoded for now, make configurable in Phase 7

5. **Session storage**: Pure JWT or DB-backed?
   - **Decision**: DB-backed for revocation capability, with JWT claims for performance
