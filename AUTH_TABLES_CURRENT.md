# Active Auth Tables (Post-Cleanup)

## JWT Authentication System

The system now uses **3 tables** for OAuth2-based JWT authentication:

### 1. `auth_states` (Temporary CSRF Protection)
**Purpose:** Track OAuth2 state during login flow (expires in 10 minutes)

**Usage:**
- Created when user initiates OAuth2 login
- Validated during OAuth2 callback
- Auto-deleted after use or expiry

**Fields:**
- `id` - Random state ID (sent to OAuth2 provider)
- `provider` - Provider name (google, github, etc.)
- `nonce` - Additional CSRF protection
- `pkce_verifier` - PKCE code verifier
- `redirect_uri` - Where to send user after auth
- `expires_at` - Expires after 10 minutes

### 2. `auth_tokens` (OAuth2 Provider Tokens)
**Purpose:** Store encrypted OAuth2 access/refresh tokens from providers

**Usage:**
- Stores encrypted tokens received from Google/GitHub
- Used for token refresh when expired
- One row per member per provider

**Fields:**
- `member_id` - Which member owns this token
- `provider` - OAuth2 provider (google, github)
- `access_token` - Encrypted provider access token
- `refresh_token` - Encrypted refresh token (nullable)
- `expires_at` - When provider token expires
- `scopes` - JSON array of granted scopes

**Note:** These are **provider tokens** (Google/GitHub), NOT our JWT tokens.
Our JWT tokens are stateless and never stored in database.

### 3. `auth_provider_links` (Member-Provider Linking)
**Purpose:** Link member accounts to OAuth2 provider identities

**Usage:**
- Links member email to provider account
- Prevents duplicate account creation
- Tracks which emails are verified

**Fields:**
- `member_id` - Member ID in cl2007_members
- `provider` - OAuth2 provider
- `provider_user_id` - User ID from provider
- `provider_email` - Email from provider
- `email_verified` - Provider verified the email
- `linked_at` - When link was created

## What's NOT Stored

❌ **JWT tokens** - Stateless, validated via signature only
❌ **Sessions** - No server-side session storage
❌ **Cookies** - Authentication is via Bearer token header

## Data Flow

```
1. User clicks "Login with Google"
   → Create auth_states entry (CSRF protection)
   
2. User authorizes with Google
   → Receive OAuth2 tokens
   → Store in auth_tokens (encrypted)
   → Create/update auth_provider_links
   → Delete auth_states (cleanup)
   → Generate JWT token (not stored!)
   
3. User makes API requests
   → Send JWT in Authorization header
   → Validate JWT signature (no DB lookup!)
   → Check scopes from JWT claims
   
4. Token refresh (if needed)
   → Use refresh_token from auth_tokens
   → Get new access_token from provider
   → Update auth_tokens with new tokens
   → Generate new JWT (not stored!)
```

## Why This Design?

**Stateless JWT:**
- No database hit for every auth check
- ~80% faster authentication
- Scales horizontally without shared session store

**Store Provider Tokens:**
- Need them for token refresh
- Need them to call provider APIs (if needed)
- Encrypted at rest for security

**Track Provider Links:**
- Prevent duplicate accounts
- Support multiple OAuth providers per member
- Audit trail of linked accounts
