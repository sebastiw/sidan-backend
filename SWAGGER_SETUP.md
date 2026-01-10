# Swagger Auto-Generation Setup

**Date**: 2026-01-10  
**Status**: Complete

## Summary

Implemented automatic Swagger/OpenAPI documentation generation from code annotations using `swaggo/swag`. Documentation covers all endpoints **except** `/auth/*` which implement the OAuth2 flow. Includes a user-friendly authentication page at `/swagger-auth` for easy JWT token acquisition.

## What Was Done

### 1. Dependencies Added
```bash
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/http-swagger
```

### 2. Main File Annotations (`src/sidan-backend.go`)
- Added API metadata (title, version, description)
- Configured JWT Bearer authentication scheme
- Import generated docs package
- Added link to `/swagger-auth` in description

### 3. Handler Annotations
Added minimal Swagger comments to:
- **Entry endpoints** (`src/router/entry.go`):
  - `POST /db/entries` - Create entry
  - `GET /db/entries` - List entries (with pagination)
  - `GET /db/entries/{id}` - Get entry by ID
  - `PUT /db/entries/{id}` - Update entry (requires auth)
  - `DELETE /db/entries/{id}` - Delete entry (requires auth)

- **Member endpoints** (`src/router/member.go`):
  - `POST /db/members` - Create member (requires auth)
  - `GET /db/members` - List members (optional auth for full data)
  - `GET /db/members/{id}` - Get member (optional auth for full data)
  - `PUT /db/members/{id}` - Update member (requires auth)
  - `DELETE /db/members/{id}` - Delete member (requires auth)

- **File endpoints** (`src/router/file.go`):
  - `POST /file/image` - Upload image (requires auth)

- **Mail endpoint** (`src/router/mail.go`):
  - `POST /mail` - Send email (requires auth)

### 4. Swagger UI Integration
- Added `/swagger/` route in `src/router/router.go`
- Accessible at: `http://localhost:8080/swagger/index.html`

### 5. Authentication Helper Page
- Created `/swagger-auth` route
- Serves `static/swagger-auth.html` 
- Features:
  - One-click Google OAuth2 login
  - Automatic token extraction from callback
  - Copy-to-clipboard functionality
  - Step-by-step instructions
  - Clean, modern UI

### 6. Documentation
- Added `make swagger` command to Makefile
- Updated README.md with complete Swagger usage instructions
- Includes authentication workflow

## Usage

### For API Users (Getting Started)

1. **Start the server**
   ```bash
   go run src/sidan-backend.go
   ```

2. **Get your JWT token**
   - Visit: `http://localhost:8080/swagger-auth`
   - Click "Login with Google"
   - Authenticate with your Google account
   - Copy the displayed JWT token

3. **Use Swagger UI**
   - Visit: `http://localhost:8080/swagger/index.html`
   - Click the "Authorize" button (🔓 icon, top right)
   - Paste your JWT token
   - Click "Authorize"
   - Now you can test all API endpoints!

### For Developers (Maintaining Docs)

**Generate/Regenerate Docs**
```bash
make swagger
```

Or manually:
```bash
~/go/bin/swag init -g src/sidan-backend.go --output docs --parseDependency --parseInternal
```

## What's Excluded

**Auth endpoints** (`/auth/*`) are intentionally excluded from Swagger:
- `/auth/login` - OAuth2 initiation (used by `/swagger-auth`)
- `/auth/callback` - OAuth2 callback (handles token generation)
- `/auth/session` - Get current session
- `/auth/refresh` - Refresh JWT token
- `/auth/logout` - Logout

These implement the OAuth2/JWT flow used by the `/swagger-auth` helper page.

## Security Annotations

Protected endpoints include `@Security BearerAuth`:
- All `PUT/DELETE` operations
- `POST /file/image`
- `POST /mail`
- `POST /db/members`

Public endpoints (no auth required):
- `GET /db/entries` (read-only)
- `POST /db/entries` (guestbook style)
- `GET /db/members` (returns limited data when unauthenticated)

## Generated Files

```
docs/
├── docs.go          # Go package with embedded Swagger spec
├── swagger.json     # OpenAPI JSON format
└── swagger.yaml     # OpenAPI YAML format

static/
└── swagger-auth.html # OAuth2 authentication helper page
```

## Architecture

### Authentication Flow for Swagger Users

```
User visits /swagger-auth
  ↓
Clicks "Login with Google"
  ↓
Popup opens → /auth/login?provider=google
  ↓
Redirects to Google OAuth consent
  ↓
User authorizes
  ↓
Google redirects → /auth/callback
  ↓
Server generates JWT token
  ↓
Page extracts access_token from JSON response
  ↓
User clicks "Copy Token"
  ↓
User pastes into Swagger UI "Authorize" dialog
  ↓
User can now test authenticated endpoints!
```

## Philosophy

Following the **lean approach**:
- Minimal annotations (only `@Summary`, `@Tags`, `@Param`, `@Success`, `@Security`)
- No inline comments or verbose descriptions
- Auto-generates from existing code structure
- Zero impact on runtime logic
- Simple, self-contained auth helper page
- ~50 lines of annotations total across all handlers
- Single HTML file for authentication (~200 lines)

## Verification

```bash
# Check generated endpoints
grep "^  /" docs/swagger.yaml

# Output:
# /db/entries:
# /db/entries/{id}:
# /db/members:
# /db/members/{id}:
# /file/image:
# /mail:

# Verify auth endpoints excluded
grep "/auth" docs/swagger.yaml
# (no output - confirmed excluded)

# Verify auth helper page exists
ls static/swagger-auth.html
# static/swagger-auth.html

# Test auth flow
curl http://localhost:8080/swagger-auth
# (returns HTML page)
```

## Next Steps

When adding new endpoints:
1. Add minimal Swagger annotations above handler function
2. Run `make swagger` to regenerate docs
3. Verify in Swagger UI

Example annotation:
```go
// @Summary Your endpoint summary
// @Tags your-tag
// @Security BearerAuth
// @Param id path int true "ID parameter"
// @Success 200 {object} YourModel
// @Router /your/endpoint/{id} [get]
func YourHandler(w http.ResponseWriter, r *http.Request) {
    // handler code
}
```

## Troubleshooting

**"Authorize button doesn't work"**
- Use `/swagger-auth` instead - it handles the OAuth2 flow for you

**"Token expired"**
- Tokens last 8 hours
- Visit `/swagger-auth` again to get a fresh token

**"Can't copy token"**
- Try right-click → Copy on the token text
- Or manually select and copy (Ctrl+C / Cmd+C)

**"Swagger UI shows 'Failed to fetch'"**
- Make sure server is running on port 8080
- Check browser console for CORS errors

