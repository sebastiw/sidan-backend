# Authentication System

**Version**: 2.0 (OAuth2 with PKCE)  
**Date**: 2026-01-10

## Overview

OAuth2-based authentication with database-backed sessions. Supports Google and GitHub providers.

## Quick Start

### 1. Setup OAuth2 Providers

**Google Cloud Console** (https://console.cloud.google.com/):
1. Create OAuth 2.0 Client ID (Web application)
2. Add redirect URI: `http://localhost:8080/auth/callback` (or your domain)
3. Copy Client ID and Client Secret

**GitHub** (https://github.com/settings/developers):
1. Create OAuth App
2. Authorization callback URL: `http://localhost:8080/auth/callback`
3. Copy Client ID and Client Secret

### 2. Configure Application

**config/local.yaml**:
```yaml
oauth2:
  google:
    clientId: "YOUR_GOOGLE_CLIENT_ID"
    clientSecret: "YOUR_GOOGLE_CLIENT_SECRET"
    redirectURL: "http://localhost:8080/auth/callback"
    scopes: ["openid", "email", "profile"]
  github:
    clientId: "YOUR_GITHUB_CLIENT_ID"
    clientSecret: "YOUR_GITHUB_CLIENT_SECRET"
    redirectURL: "http://localhost:8080/auth/callback"
    scopes: ["user:email"]
```

### 3. Set Environment Variables

```bash
# Generate encryption key for production
export AUTH_ENCRYPTION_KEY="$(openssl rand -hex 32)"

# Start server
go run src/sidan-backend.go
```

## Authentication Flow

```
User → /auth/login?provider=google
  ↓
Google OAuth2 consent screen
  ↓
/auth/callback (validates, creates session)
  ↓
Session cookie set
  ↓
User authenticated ✓
```

### Login
```
GET /auth/login?provider=google&redirect_uri=/dashboard
```

### Check Session
```
GET /auth/session
Cookie: session_id=...

Response:
{
  "session_id": "abc...",
  "member": {
    "id": 295,
    "number": 8,
    "name": "Max Gabrielsson",
    "email": "max@example.com"
  },
  "scopes": ["write:email", "write:image", "write:member", "read:member", "modify:entry"],
  "provider": "google",
  "expires_at": "2026-01-10T22:00:00Z"
}
```

### Logout
```
POST /auth/logout
Cookie: session_id=...

Response: {"success": true}
```

### Refresh Token
```
POST /auth/refresh
Cookie: session_id=...

Response: {"success": true}
```

## Protected Endpoints

Use session cookie to access protected resources:

```bash
# Example: Upload image (requires write:image scope)
curl -X POST --cookie "session_id=YOUR_SESSION_ID" \
  -F "file=@image.jpg" \
  "http://localhost:8080/file/image"
```

## Scopes

- `write:email` - Send emails
- `write:image` - Upload images
- `write:member` - Create/update/delete members
- `read:member` - Read full member details
- `modify:entry` - Update/delete entries

Scopes automatically assigned based on member's `isvalid` status.

## Database Tables

- `auth_states` - OAuth2 state (CSRF protection, 10min TTL)
- `auth_tokens` - Encrypted access/refresh tokens
- `auth_provider_links` - Links OAuth2 accounts to members
- `auth_sessions` - Active sessions (8hr expiry)

## Member Registration

Members must exist in `cl2007_members` table with verified email before they can login via OAuth2.

```sql
-- Add member email
UPDATE cl2007_members SET email = 'user@example.com' WHERE id = 1;

-- Or create new member
INSERT INTO cl2007_members (number, name, email, isvalid)
VALUES (9999, 'John Doe', 'john@example.com', true);
```

## Security Features

- ✅ **PKCE** - Protects against authorization code interception
- ✅ **State validation** - CSRF protection
- ✅ **Token encryption** - AES-256-GCM at rest
- ✅ **HttpOnly cookies** - XSS protection
- ✅ **Session expiry** - 8-hour TTL
- ✅ **Activity tracking** - Last activity timestamp
- ✅ **Automatic cleanup** - Expired sessions deleted every 15min

## Architecture

### Components

**Phase 1**: Database schema & token encryption  
**Phase 2**: Provider abstraction (Google, GitHub)  
**Phase 3**: OAuth2 handlers (login, callback, session, logout)  
**Phase 4**: Middleware (RequireAuth, RequireScope, OptionalAuth)  
**Phase 5**: Legacy code removal

### Files

```
src/auth/
  ├── crypto.go           # Token encryption (AES-256-GCM)
  ├── pkce.go             # PKCE verifier/challenge generation
  ├── provider.go         # OAuth2 provider abstraction
  └── middleware.go       # Auth middleware & cleanup

src/router/
  └── auth.go             # HTTP handlers (login, callback, session, logout, refresh)

src/models/
  └── auth.go             # Database models (AuthState, AuthToken, etc.)

src/data/commondb/
  └── auth.go             # Database operations
```

## Troubleshooting

**"redirect_uri_mismatch"**:
- Ensure redirect URI in Google/GitHub console matches config exactly
- Must be: `http://localhost:8080/auth/callback` (or your domain)

**"email not registered"**:
- Member email must exist in `cl2007_members` table
- Email must match OAuth2 provider's verified email

**"no session"**:
- Session expired (8hr) or invalid
- User needs to login again

**"Using default encryption key" warning**:
- Set `AUTH_ENCRYPTION_KEY` environment variable for production
- Generate: `openssl rand -hex 32`

## Production Deployment

1. Set `AUTH_ENCRYPTION_KEY` environment variable
2. Update OAuth2 redirect URIs to production domain
3. Use HTTPS (required for secure cookies)
4. Update `config/production.yaml` with production settings
5. Notify users: re-login required after deployment

## Code Philosophy

This implementation follows a **lean and pragmatic** approach:

- Direct function calls over service layers
- Inline logic over abstractions
- Standard library patterns over frameworks
- ~1,500 lines total vs typical 3,000+ lines
- Zero enterprise bloat

See `AGENT.md` for code philosophy details.
