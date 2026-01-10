# AGENT.md - Project Reference for LLMs

## Project Overview
**sidan-backend** is a forum-like REST API backend service written in Go (Golang 1.15+). It provides endpoints for user authentication via OAuth2, database operations on forum entries and member profiles, file uploads, and email functionality. The service is designed for the Chalmers Losers community forum application.

## Technology Stack
- **Language**: Go 1.15
- **Web Framework**: Gorilla Mux (routing), Gorilla Sessions (session management)
- **Database**: MySQL with GORM ORM
- **Authentication**: OAuth2 (Google, GitHub) + custom "Sidan" provider
- **CORS**: rs/cors library
- **Config Management**: Viper (YAML-based configuration)
- **API Specification**: OpenAPI 3.0.1 (swagger.yaml)

## Project Structure

```
sidan-backend/
├── src/                          # Main source code
│   ├── sidan-backend.go         # Entry point, starts HTTP server
│   ├── auth/                     # Authentication & authorization
│   │   ├── auth.go              # Session management, scope checking
│   │   ├── auth_handlers.go    # OAuth2 handlers for providers
│   │   └── sidan_auth_handler.go # Custom login system
│   ├── config/                   # Configuration parsing
│   │   └── config.go            # Reads YAML configs via Viper
│   ├── data/                     # Database abstraction layer
│   │   ├── database.go          # Database interface definition
│   │   ├── commondb/            # Shared database operations
│   │   └── mysqldb/             # MySQL-specific implementation
│   ├── models/                   # Data models (GORM entities)
│   │   ├── entry.go             # Forum entry/post model
│   │   ├── members.go           # Member profile model
│   │   ├── user.go              # User authentication model
│   │   ├── prospect.go          # Prospect user model
│   │   └── settings.go          # System settings model
│   ├── router/                   # HTTP handlers & routing
│   │   ├── router.go            # Main router setup with all endpoints
│   │   ├── entry.go             # CRUD handlers for entries
│   │   ├── member.go            # CRUD handlers for members
│   │   ├── file.go              # Image upload handler
│   │   ├── mail.go              # Email sending handler
│   │   └── common.go            # Shared utilities
│   ├── router_util/              # Request utilities (tracing, IDs)
│   ├── logger/                   # Logging setup
│   └── enums/                    # Enumeration types
├── config/                       # Configuration files
│   └── local.yaml               # Local dev config (DB, OAuth2, etc.)
├── db/                           # Database migrations & SQL scripts
│   ├── 2021-02-17-tabeller.sql  # Table definitions
│   ├── 2021-02-24-test-data.sql # Test data
│   └── 2021-04-05-procedures.sql # Stored procedures
├── static/                       # Static file serving directory
├── swagger.yaml                  # OpenAPI 3.0.1 specification
├── Makefile                      # Docker commands for MySQL
├── Dockerfile.sql                # MySQL Docker image
└── go.mod                        # Go module dependencies
```

## Key Components

### 1. Main Entry Point (`src/sidan-backend.go`)
- Initializes logger and config
- Creates database connection
- Sets up router with all endpoints
- Starts HTTP server on configured port (default: 8080)

### 2. Router (`src/router/router.go`)
**Core responsibilities**:
- Defines all HTTP endpoints
- Applies CORS headers (allows specific origins)
- HTTP request logging with duration tracking
- Request ID generation for tracing

**Endpoint groups**:
- `/auth/{provider}` - OAuth2 authentication flows
- `/login/*` - Custom Sidan authentication system
- `/file/*` - Static file serving & image uploads
- `/mail` - Email sending
- `/db/entries` - Forum entry CRUD operations
- `/db/members` - Member profile CRUD operations

### 3. Authentication (`src/auth/`)

**OAuth2 Providers** (`auth_handlers.go`):
- Supports Google, GitHub, and custom "Sidan" provider
- Three-step flow: redirect → callback → verify email
- Session-based token storage with 8-hour expiry
- Scope-based authorization (write:email, write:image, write:member, modify:entry, read:member)

**Custom "Sidan" Provider** (`sidan_auth_handler.go`):
- Username/password authentication via HTML form
- Username format: `#<number>`, `P<number>`, or `S<number>` (Member, Prospect, Suspect)
- Password verification against `password_classic` field in database
- Returns authorization code for OAuth2-like flow

**Session Management** (`auth.go`):
- Cookie-based sessions with secure random keys
- Token validation with expiry checking
- Scope enforcement via `CheckScope()` middleware

### 4. Database Layer

**Interface** (`src/data/database.go`):
Defines contract for all database operations:
- User operations: `GetUserFromEmails()`, `GetUserFromLogin()`
- Entry CRUD: `CreateEntry()`, `ReadEntry()`, `ReadEntries()`, `UpdateEntry()`, `DeleteEntry()`
- Member CRUD: `CreateMember()`, `ReadMember()`, `ReadMembers()`, `UpdateMember()`, `DeleteMember()`
- Settings: `GetSettingsById()`

**Implementation** (`src/data/mysqldb/db.go`):
- Uses GORM v1.26+ with MySQL driver
- Connection string format: `user:pass@tcp(host:port)/schema?charset=utf8mb4&parseTime=True`
- Session SQL mode configuration for strict error handling
- Delegates to `commondb` package for shared logic

**Common DB Operations** (`src/data/commondb/`):
- Implements database interface methods
- User lookup by email list (for OAuth2 email verification)
- User login validation with username/password

### 5. Data Models (`src/models/`)

**Entry** (`entry.go`):
```go
type Entry struct {
    Id             int64      // Primary key
    Date           string     // Entry date (legacy format)
    Time           string     // Entry time (legacy format)
    DateTime       time.Time  // ISO datetime
    Msg            string     // Forum post content (required)
    Status         int64      // Entry status code
    Cl             int64      // CL identifier
    Sig            string     // Signature (required)
    Email          string     // Author email
    Place          string     // Location string
    Ip, Host       *string    // Network info (nullable)
    Olsug          int64      // Beer suggestions
    Enheter        int64      // Units
    Lat, Lon       *float64   // GPS coordinates (nullable)
    Report         bool       // Reported flag
    Likes          int64      // Like count
    Secret         bool       // Secret flag
    PersonalSecret bool       // Personal secret flag
    SideKicks      []SideKick // Related side-kick entries
}
// Table: cl2003_msgs
```

**Member** (`members.go`):
```go
type Member struct {
    Id                           int64   // Primary key
    Number                       int64   // Member number (required)
    Name, Email, Phone           *string // Contact info (nullable)
    Im                           string  // Instant messenger
    Adress, Adressurl            *string // Address details
    Title                        *string // Member title
    History                      *string // Member history text
    Picture                      *string // Profile picture URL
    Password                     *string // OAuth2 password
    Password_classic             *string // Legacy password
    Password_resetstring         *string // Reset tokens
    Password_classic_resetstring *string
    Isvalid                      *bool   // Active member flag
}
// Table: cl2007_members
```

**User** (`user.go`):
```go
type User struct {
    Type       UserType // "#" (Member), "P" (Prospect), "S" (Suspect)
    Number     int64    // User number
    Email      string   // Email address
    FulHaxPass string   // Password hash
}
```

**MemberLite** (`members.go`):
- Subset of Member fields for unauthenticated requests
- Only exposes: Id, Number, Title

### 6. HTTP Handlers (`src/router/`)

**Entry Handlers** (`entry.go`):
- `POST /db/entries` - Create new entry (public)
- `GET /db/entries/{id}` - Read single entry (public)
- `GET /db/entries?skip=0&take=20` - List entries with pagination (public)
- `PUT /db/entries/{id}` - Update entry (requires `modify:entry` scope)
- `DELETE /db/entries/{id}` - Delete entry (requires `modify:entry` scope)

**Member Handlers** (`member.go`):
- `POST /db/members` - Create member (requires `write:member` scope)
- `GET /db/members/{id}` - Read member (returns full data if authenticated with `read:member`, otherwise MemberLite)
- `GET /db/members?onlyValid=false` - List members (conditional auth, returns MemberLite if not authenticated)
- `PUT /db/members/{id}` - Update member (requires `write:member` scope)
- `DELETE /db/members/{id}` - Delete member (requires `write:member` scope)

**File Handlers** (`file.go`):
- `POST /file/image` - Upload image (requires `write:image` scope)
  - Max size: 10 MB
  - Supported types: image/gif, image/png, image/jpeg
  - Saves to `static/` directory with random filename
  - Returns: `{"filename": "upload-*.ext"}`
- `GET /file/{filename}` - Serve static files from `static/` directory (public)

**Mail Handlers** (`mail.go`):
- `POST /mail` - Send email (requires `write:email` scope)
  - JSON body: `{"from_email": "", "to_emails": [], "message": "", "title": ""}`
  - Uses SMTP with configured host/port/credentials
  - Returns: `{"Result": "ok"}`

### 7. Configuration (`config/local.yaml`)
```yaml
server:
  port: 8080
  staticPath: "./static"

database:
  schema: "dbschema"      # MySQL database name
  user: "dbuser"
  password: "dbpassword"

mail:
  host: "localhost"
  port: 25

oauth2:
  sidan:                  # Custom provider
    clientId: "666"
    clientSecret: "s1d4n"
    redirectURL: "auth/sidan/authorized"
    scopes: ["user:email"]
  google:
    clientId: "1234"
    clientSecret: "secret"
    redirectURL: "google/dummy"
    scopes: ["openid", "email"]
  github:
    clientId: "54321"
    clientSecret: "really_secret"
    redirectURL: "github/dummy"
    scopes: ["user:email"]
```

**Environment Variables**:
- `CONFIG_FILE` - Override config file name (default: "local")
- `SESSION_KEY` - Auto-generated secure random key for sessions

## Database Schema

**Key Tables**:
- `cl2003_msgs` - Forum entries/posts
- `cl2003_msgs_kumpaner` - Side-kick relationships (foreign key to entries)
- `cl2007_members` - Member profiles and authentication
- `2003_ditch` - Deleted/ditched entries archive
- `2003_likes` - Entry likes tracking

**Important Fields**:
- Members: `isvalid` (bool) filters active vs inactive members
- Members: `password_classic` used for custom authentication
- Entries: `secret` and `personal_secret` flags for visibility control

## Running the Application

**Local Development**:
1. Start MySQL: `make` (runs Docker container with migrations)
2. Start service: `go run src/sidan-backend.go`
3. Service listens on: `http://localhost:8080`

**Docker**:
- Database runs in Docker container named `sidan_sql`
- Network: `backend-network`
- Port: 3306 (MySQL)
- Migrations auto-run on container start from `db/` directory

## API Endpoints Summary

### Authentication
- `GET /auth/{provider}` - Initiate OAuth2 flow (provider: google, github, sidan)
- `GET /auth/{provider}/authorized` - OAuth2 callback handler
- `GET /auth/{provider}/verifyemail` - Verify email against database
- `GET /auth/getusersession` - Get current session info with scopes
- `GET /login/oauth/authorize` - Show custom login form
- `POST /login` - Submit username/password for custom auth
- `POST /login/oauth/access_token` - Exchange code for access token

### Database Operations
- `GET /db/entries?skip=0&take=20` - List entries (pagination)
- `POST /db/entries` - Create entry
- `GET /db/entries/{id}` - Read entry
- `PUT /db/entries/{id}` - Update entry (auth required)
- `DELETE /db/entries/{id}` - Delete entry (auth required)
- `GET /db/members?onlyValid=false` - List members
- `POST /db/members` - Create member (auth required)
- `GET /db/members/{id}` - Read member
- `PUT /db/members/{id}` - Update member (auth required)
- `DELETE /db/members/{id}` - Delete member (auth required)

### Files & Mail
- `POST /file/image` - Upload image (auth required, max 10MB)
- `GET /file/{filename}` - Download static file
- `POST /mail` - Send email (auth required)

## Authentication Flow Examples

**OAuth2 (Google/GitHub)**:
1. Frontend redirects to `GET /auth/google`
2. User logs in with Google, redirected to `GET /auth/google/authorized?code=...&state=...`
3. Backend exchanges code for access token, stores in session
4. Frontend calls `GET /auth/google/verifyemail` to link email to member account
5. Session now contains scopes based on member type

**Custom Sidan Auth**:
1. Frontend redirects to `GET /login/oauth/authorize?redirect_uri=...&state=...`
2. User sees login form (login.html)
3. User submits `POST /login` with username (e.g., "#1234") and password
4. Backend validates against `cl2007_members.password_classic`
5. Redirects back with authorization code
6. Frontend exchanges code via `POST /login/oauth/access_token`

## Scope-Based Authorization

**Available Scopes**:
- `write:email` - Send emails via `/mail`
- `write:image` - Upload images via `/file/image`
- `write:member` - Create/update/delete members
- `read:member` - Read full member details (otherwise only MemberLite)
- `modify:entry` - Update/delete entries

**Scope Assignment**:
- Scopes stored in session after successful authentication
- Assigned based on user type and OAuth2 provider configuration
- Enforced via `auth.CheckScope()` middleware wrapper

## Common Development Tasks

**Add New Endpoint**:
1. Define model in `src/models/` if needed
2. Add database method to `src/data/database.go` interface
3. Implement in `src/data/mysqldb/` and/or `src/data/commondb/`
4. Create handler in `src/router/`
5. Register route in `src/router/router.go`
6. Update `swagger.yaml`

**Add New Database Operation**:
1. Add method signature to `Database` interface in `src/data/database.go`
2. Implement in `src/data/mysqldb/<model>.go` or `src/data/commondb/<model>.go`
3. Use GORM methods: `db.Create()`, `db.First()`, `db.Find()`, `db.Save()`, `db.Delete()`

**Add New OAuth2 Provider**:
1. Add provider config to `config/local.yaml` under `oauth2:`
2. Add provider name to `providers` array in `src/auth/auth.go`
3. Routes auto-registered in `src/router/router.go` via loop
4. Implement provider-specific email verification if needed

**Change Database Schema**:
1. Create new migration SQL file in `db/` with date prefix (e.g., `2024-01-10-new-feature.sql`)
2. Add to Dockerfile.sql or mount volume if using existing container
3. Restart database container: `make db-stop && make`

## Dependencies (go.mod)
- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/gorilla/mux` - HTTP router
- `github.com/gorilla/sessions` - Session management
- `github.com/rs/cors` - CORS middleware
- `github.com/spf13/viper` - Configuration management
- `golang.org/x/oauth2` - OAuth2 client
- `gorm.io/gorm` - ORM framework
- `gorm.io/driver/mysql` - GORM MySQL driver

## Important Notes

1. **Session Security**: Sessions use secure random keys generated at startup. Keys change on restart, invalidating existing sessions.

2. **CORS**: Only allows specific origins (chalmerslosers.com, sidan.cl, localhost). Add new origins in `src/router/router.go` `corsHeaders()`.

3. **Error Handling**: Most handlers return HTTP 500 for any error. Consider adding more granular error codes in production.

4. **Pagination**: Entry listing supports `skip` and `take` query params (default: skip=0, take=20).

5. **Member Visibility**: Unauthenticated requests only see MemberLite data (Id, Number, Title). Full details require `read:member` scope.

6. **Database Tables**: Use legacy naming convention (e.g., `cl2003_msgs`, `cl2007_members`). Don't change without migration plan.

7. **Image Storage**: Uploaded images stored in `static/` directory with random filenames. No database tracking of uploads.

8. **Mail Configuration**: SMTP credentials must be configured in `config/<env>.yaml`. No mail sent in local dev by default.

9. **Logging**: Uses Go's `log/slog` package. Debug logs include request IDs for tracing. Set log level via environment.

10. **HTTP Methods**: 
    - POST = Create
    - GET = Read
    - PUT = Update
    - DELETE = Delete
    - (Not RESTful: should use POST for create, not PUT as documented in README)

## Testing
- No automated tests currently exist (marked as TODO in README)
- Manual testing via curl or API client recommended
- Test data available in `db/2021-02-24-test-data.sql`

## Important Notes for LLMs

**NEVER use git commands**. The developer will review and commit all changes manually. This includes:
- `git add`
- `git commit`
- `git push`
- `git checkout`
- `git branch`
- Any other git operations

Focus on code implementation, testing, and documentation only.

## Code Philosophy: Lean and Pragmatic

**Under NO circumstances implement enterprise patterns or boilerplate code.** This project values:

### What to DO ✅
- **Simple, direct solutions** - Solve the problem with minimal code
- **Clear code flow** - Top-to-bottom readability
- **Plain functions** - Not everything needs to be a class or interface
- **Switch statements** - Perfectly fine for 2-5 cases
- **Direct calls** - No unnecessary wrappers or indirection
- **Explicit code** - Better than clever abstractions
- **Inline logic** - If it's only used once, don't extract it

### What to AVOID ❌
- **Abstract factories** - Use simple factory functions only when genuinely needed
- **Excessive interfaces** - Only create interfaces when you have 3+ implementations
- **Plugin systems** - YAGNI (You Aren't Gonna Need It)
- **Middleware layers** - Only add when actually reused 3+ times
- **Dependency injection frameworks** - Pass dependencies as function parameters
- **Complex hierarchies** - Flat is better than nested
- **Registry patterns** - A map and a switch statement work fine
- **Builder patterns** - For simple objects, just use struct literals
- **Strategy patterns** - A function parameter works fine
- **Decorator patterns** - Wrap directly when needed

### Decision Framework
Before adding abstraction, ask:
1. **Do I have 3+ implementations?** If no, use a simple function
2. **Is this code reused 3+ times?** If no, keep it inline
3. **Will this change frequently?** If no, hardcode is fine
4. **Does this add real value?** If no, delete it

### Example: The Right Way
```go
// ✅ GOOD: Direct and clear
func GetProviderConfig(provider string) (*Config, error) {
    switch provider {
    case "google": return googleConfig()
    case "github": return githubConfig()
    default: return nil, errors.New("unknown provider")
    }
}

// ❌ BAD: Enterprise bloat
type ProviderFactory interface {
    CreateProvider(name string) (Provider, error)
}
type Provider interface {
    GetConfig() Config
    Initialize() error
    // ... 10 more methods
}
```

### Measuring Success
- **Line count**: Less is more (within reason)
- **Time to understand**: Can someone grok it in 5 minutes?
- **Dependencies**: Fewer is better
- **Indirection levels**: Max 2 (caller → function → implementation)
- **Test complexity**: If tests are complex, code is too clever

**Remember**: Professional code is not about showing off patterns. It's about solving problems clearly and maintainably.

## Security Considerations
- Passwords stored in plaintext in `password_classic` field (legacy system)
- OAuth2 tokens stored in encrypted cookies (gorilla/securecookie)
- Session cookies are HttpOnly but not Secure (consider enabling for HTTPS)
- No rate limiting implemented
- File uploads limited to 10MB, only images allowed
- SQL injection protected by GORM parameterization

## Future Improvements (from README)
- [ ] Test locally
- [ ] Tests for APIs  
- [ ] Content-Type: application/json enforcement
- [ ] /notify/ endpoints for notifications
- [ ] Optional: New tables and data migration
- [ ] Optional: HEAD on API for descriptions
- [ ] Optional: GET on base URIs for API documentation
