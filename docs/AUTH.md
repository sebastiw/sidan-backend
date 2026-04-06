# Authentication System

**Version**: 3.0 (JWT Bearer Tokens)  
**Date**: 2026-01-10

## Overview

JWT Bearer token authentication with OAuth2 providers. Designed for mobile apps and web clients.

## Quick Start

### 1. Setup OAuth2 Providers

**Google Cloud Console** (https://console.cloud.google.com/):
1. Create OAuth 2.0 Client ID (Web application)
2. Add redirect URI: `http://localhost:8080/auth/web/callback` (or your domain)
3. Copy Client ID and Client Secret

**GitHub** (https://github.com/settings/developers):
1. Create OAuth App
2. Authorization callback URL: `http://localhost:8080/auth/web/callback`
3. Copy Client ID and Client Secret

### 2. Configure Application

**config/local.yaml**:
```yaml
oauth2:
  google:
    clientId: "YOUR_GOOGLE_CLIENT_ID"
    clientSecret: "YOUR_GOOGLE_CLIENT_SECRET"
    redirectURL: "http://localhost:8080/auth/web/callback"
    scopes: ["openid", "email", "profile"]
  github:
    clientId: "YOUR_GITHUB_CLIENT_ID"
    clientSecret: "YOUR_GITHUB_CLIENT_SECRET"
    redirectURL: "http://localhost:8080/auth/web/callback"
    scopes: ["user:email"]
```

### 3. Set Environment Variables

```bash
# Generate JWT secret (REQUIRED for production)
export JWT_SECRET="$(openssl rand -hex 64)"

# Start server
go run src/sidan-backend.go
```

### 4. Start Server

```bash
export JWT_SECRET="$(openssl rand -hex 64)"
go run src/sidan-backend.go
```

## Authentication Flow

```
User → /auth/web/login?provider=google
  ↓
Google OAuth2 consent screen
  ↓
/auth/web/callback (validates, generates JWT)
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
GET /auth/web/login?provider=google&redirect_uri=/dashboard
```

### Callback (Returns JWT)
```
GET /auth/web/callback?state=...&code=...

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "...",
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
POST /auth/web/logout
Authorization: Bearer <token>

Response: {"success": true}
```

### Refresh Token
```
POST /auth/web/refresh
Content-Type: application/json

{"refresh_token": "..."}

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "...",
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
- `oauth2_sessions` - Long-lived refresh tokens (30d TTL)

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
- ✅ **JWT signatures** - HS256 with 512-bit secret
- ✅ **Token expiry** - 8-hour JWT lifetime
- ✅ **Refresh token rotation** - Single-use refresh tokens (30d)
- ✅ **Automatic cleanup** - Expired states and sessions deleted every 15min

## Client Implementation

### Web (JavaScript)

```javascript
// 1. Login - Redirect to OAuth2
window.location.href = '/auth/web/login?provider=google'

// 2. Handle callback - Extract JWT from response
const response = await fetch('/auth/web/callback?...')
const data = await response.json()
localStorage.setItem('access_token', data.access_token)
localStorage.setItem('refresh_token', data.refresh_token)

// 3. API requests - Send Bearer token
const token = localStorage.getItem('access_token')
const res = await fetch('/db/entries', {
  headers: { 'Authorization': `Bearer ${token}` }
})

// 4. Refresh token when JWT expires
const refresh = await fetch('/auth/web/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ refresh_token: localStorage.getItem('refresh_token') })
})
const refreshData = await refresh.json()
localStorage.setItem('access_token', refreshData.access_token)
localStorage.setItem('refresh_token', refreshData.refresh_token)

// 5. Logout
await fetch('/auth/web/logout', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
})
localStorage.removeItem('access_token')
localStorage.removeItem('refresh_token')
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
**Session Store**: Refresh token storage in `oauth2_sessions` table

### Files

```
src/auth/
  ├── jwt.go              # JWT sign/verify
  ├── pkce.go             # PKCE verifier/challenge generation
  ├── provider.go         # OAuth2 provider abstraction
  └── middleware.go       # Auth middleware & cleanup

src/router/
  └── auth.go             # HTTP handlers (login, callback, session, logout, refresh)

src/models/
  └── session.go          # Session model (oauth2_sessions table)

src/data/commondb/
  └── session.go          # Session CRUD operations
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
- Use `POST /auth/web/refresh` with a valid refresh token, or re-login
- Note: Logout does not revoke JWTs (they remain valid until expiry)

**"redirect_uri_mismatch"**:
- Ensure redirect URI in Google/GitHub console matches config exactly
- Must be: `http://localhost:8080/auth/web/callback` (or your domain)

**"email not registered"**:
- Member email must exist in `cl2007_members` table
- Email must match OAuth2 provider's verified email

**"Using default JWT secret" warning**:
- Set `JWT_SECRET` environment variable for production
- Generate: `openssl rand -hex 64`

## Production Deployment

1. Set `JWT_SECRET` environment variable (required)
2. Update OAuth2 redirect URIs to production domain (`/auth/web/callback`)
3. Use HTTPS (required for secure tokens)
4. Update `config/production.yaml` with production settings
5. Notify users: re-login required after deployment (breaking change)

## Code Philosophy

This implementation follows a **lean and pragmatic** approach:

- JWT validation in middleware (no database lookup)
- Direct function calls over service layers
- Standard library JWT patterns
- Zero enterprise bloat

See `AGENT.md` for code philosophy details.
