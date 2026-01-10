# Authentication System

**Version**: 3.0 (JWT Bearer Tokens)  
**Date**: 2026-01-10

## Overview

JWT Bearer token authentication with OAuth2 providers. Designed for mobile apps and web clients.

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
# Generate JWT secret (REQUIRED for production)
export JWT_SECRET="$(openssl rand -hex 64)"

# Generate encryption key for OAuth2 tokens
export AUTH_ENCRYPTION_KEY="$(openssl rand -hex 32)"

# Start server
go run src/sidan-backend.go
```

### 4. Start Server

```bash
export JWT_SECRET="$(openssl rand -hex 64)"
export AUTH_ENCRYPTION_KEY="$(openssl rand -hex 32)"
go run src/sidan-backend.go
```

## Authentication Flow

```
User → /auth/login?provider=google
  ↓
Google OAuth2 consent screen
  ↓
/auth/callback (validates, generates JWT)
  ↓
Returns JWT token in response body
  ↓
Client stores JWT
  ↓
Client sends: Authorization: Bearer <token>
  ↓
User authenticated ✓
```

### Login
```
GET /auth/login?provider=google&redirect_uri=/dashboard
```

### Callback (Returns JWT)
```
GET /auth/callback?state=...&code=...

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 28800,
  "member": {
    "id": 123,
    "email": "user@example.com",
    "name": "John Doe"
  },
  "scopes": ["write:email", "write:image", ...]
}
```

### Check Session
```
GET /auth/session
Authorization: Bearer <token>

Response:
{
  "member": {
    "id": 123,
    "number": 42,
    "name": "John Doe",
    "email": "user@example.com"
  },
  "scopes": ["write:email", ...],
  "provider": "google",
  "expires_at": "2026-01-10T22:00:00Z"
}
```

### Logout
```
POST /auth/logout
Authorization: Bearer <token>

Response: {"success": true}
```

### Refresh Token
```
POST /auth/refresh
Authorization: Bearer <token>

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 28800
}
```

## Protected Endpoints

Send JWT Bearer token in Authorization header:

```bash
# Example: Upload image (requires write:image scope)
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@image.jpg" \
  "http://localhost:8080/file/image"

# Example: Create entry
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"msg":"Hello","sig":"John"}' \
  "http://localhost:8080/db/entries"
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
- `auth_tokens` - Encrypted OAuth2 access/refresh tokens
- `auth_provider_links` - Links OAuth2 accounts to members

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
- ✅ **Token encryption** - OAuth2 tokens encrypted with AES-256-GCM at rest
- ✅ **JWT signatures** - HS256 with 512-bit secret
- ✅ **Token expiry** - 8-hour JWT lifetime
- ✅ **Automatic cleanup** - Expired OAuth2 states deleted every 15min

## Client Implementation

### Web (JavaScript)

```javascript
// 1. Login - Redirect to OAuth2
window.location.href = '/auth/login?provider=google'

// 2. Handle callback - Extract JWT from response
const response = await fetch('/auth/callback?...')
const data = await response.json()
localStorage.setItem('access_token', data.access_token)

// 3. API requests - Send Bearer token
const token = localStorage.getItem('access_token')
const res = await fetch('/db/entries', {
  headers: { 'Authorization': `Bearer ${token}` }
})

// 4. Logout
await fetch('/auth/logout', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
})
localStorage.removeItem('access_token')
```

### Mobile (Swift/Kotlin)

```swift
// iOS - Store in Keychain
import Security

// Save token
KeychainHelper.save(token: jwtToken, service: "SidanAPI")

// Use token
let token = KeychainHelper.load(service: "SidanAPI")
request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
```

```kotlin
// Android - Store in EncryptedSharedPreferences
val sharedPrefs = EncryptedSharedPreferences.create(...)
sharedPrefs.edit().putString("jwt_token", token).apply()

// Use token
val token = sharedPrefs.getString("jwt_token", null)
request.addHeader("Authorization", "Bearer $token")
```

## Architecture

### Components

**JWT System**: Token generation and validation  
**OAuth2 Flow**: Provider abstraction (Google, GitHub)  
**Middleware**: RequireAuth, RequireScope, OptionalAuth  
**Blacklist**: Token revocation support

### Files

```
src/auth/
  ├── jwt.go              # JWT sign/verify
  ├── crypto.go           # OAuth2 token encryption (AES-256-GCM)
  ├── pkce.go             # PKCE verifier/challenge generation
  ├── provider.go         # OAuth2 provider abstraction
  └── middleware.go       # Auth middleware & cleanup

src/router/
  └── auth.go             # HTTP handlers (login, callback, session, logout, refresh)

src/models/
  └── auth.go             # Database models (AuthState, AuthToken, JWTBlacklist)

src/data/commondb/
  └── auth.go             # Database operations
```

## Troubleshooting

**"missing authorization header"**:
- Add `Authorization: Bearer <token>` header to request

**"invalid authorization format"**:
- Ensure format is exactly: `Authorization: Bearer <token>`

**"invalid token"**:
- Token is malformed or has invalid signature
- User needs to re-login

**"token expired"**:
- JWT lifetime exceeded (8 hours)
- Use `/auth/refresh` or re-login
- Note: Logout does not revoke JWTs (they remain valid until expiry)

**"redirect_uri_mismatch"**:
- Ensure redirect URI in Google/GitHub console matches config exactly
- Must be: `http://localhost:8080/auth/callback` (or your domain)

**"email not registered"**:
- Member email must exist in `cl2007_members` table
- Email must match OAuth2 provider's verified email

**"Using default JWT secret" warning**:
- Set `JWT_SECRET` environment variable for production
- Generate: `openssl rand -hex 64`

## Production Deployment

1. Set `JWT_SECRET` environment variable (required)
2. Set `AUTH_ENCRYPTION_KEY` environment variable (for OAuth2 tokens)
3. Update OAuth2 redirect URIs to production domain
4. Use HTTPS (required for secure tokens)
5. Run database migration: `db/2026-01-10-jwt-blacklist.sql`
6. Update `config/production.yaml` with production settings
7. Notify users: re-login required after deployment (breaking change)

## Migration from Cookie-Based Auth

**Key Changes**:
- Cookies → JWT Bearer tokens
- Server sessions → Client-side tokens
- `Cookie: session_id` → `Authorization: Bearer <token>`
- ~80% faster auth validation (no DB lookup)
- JWTs valid until expiry (logout is client-side only)

## Code Philosophy

This implementation follows a **lean and pragmatic** approach:

- JWT validation in middleware (no database lookup)
- Direct function calls over service layers
- Standard library JWT patterns
- ~1,700 lines total for auth system
- Zero enterprise bloat

See `AGENT.md` for code philosophy details.
