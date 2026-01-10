# Phase 4 Testing Guide

## Prerequisites

1. Database running with Phase 1 tables
2. Valid OAuth2 credentials configured
3. At least one member with registered email in database

## Test 1: Environment Variable

### Test AUTH_ENCRYPTION_KEY is read

```bash
# Without env var (should see warning)
go run src/sidan-backend.go 2>&1 | grep "encryption"
# Expected: "Using default encryption key - set AUTH_ENCRYPTION_KEY in production"

# With env var (no warning)
export AUTH_ENCRYPTION_KEY="$(openssl rand -hex 32)"
go run src/sidan-backend.go 2>&1 | grep "encryption"
# Expected: No warning
```

## Test 2: Cleanup Job Starts

```bash
# Start server and check logs
go run src/sidan-backend.go 2>&1 | grep "cleanup"
# Expected: "cleanup job started" interval=15m0s
```

## Test 3: Refresh Endpoint

### Step 1: Login via Google
```bash
# Open in browser:
http://localhost:8080/auth/login?provider=google

# Complete OAuth2 flow
# You should be redirected back
```

### Step 2: Get Session Cookie
```bash
# Check browser cookies or use:
curl -v "http://localhost:8080/auth/session" \
  --cookie "session_id=YOUR_SESSION_ID"

# Should return member info and scopes
```

### Step 3: Call Refresh Endpoint
```bash
curl -X POST "http://localhost:8080/auth/refresh" \
  --cookie "session_id=YOUR_SESSION_ID"

# Expected response:
# {"success":true}
```

### Step 4: Verify Token Updated
```sql
SELECT member_id, provider, expires_at, updated_at 
FROM auth_tokens 
ORDER BY updated_at DESC 
LIMIT 5;

-- updated_at should be very recent (within last minute)
```

## Test 4: Automatic Cleanup

### Step 1: Create Expired Session
```sql
INSERT INTO auth_sessions (id, member_id, data, created_at, expires_at, last_activity)
VALUES ('test-expired', 1, '{"scopes":["test"],"provider":"test"}', 
        NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR);
```

### Step 2: Check It Exists
```sql
SELECT COUNT(*) FROM auth_sessions WHERE id = 'test-expired';
-- Should return: 1
```

### Step 3: Wait for Cleanup
```bash
# Start server
go run src/sidan-backend.go

# Wait 15+ minutes (or modify interval to 1 minute for testing)

# Check logs for cleanup
# Expected: "cleaned up expired auth data"
```

### Step 4: Verify Deletion
```sql
SELECT COUNT(*) FROM auth_sessions WHERE id = 'test-expired';
-- Should return: 0
```

## Test 5: Middleware Functions (Unit Test Style)

Since middleware will be wired up in Phase 5, we can test the logic directly:

```bash
# Create a simple test file
cat > src/auth/middleware_test.go << 'EOF'
package auth

import (
	"testing"
	"time"
)

func TestCleanupExpired(t *testing.T) {
	// This would need a test database
	// Just verify it compiles and can be called
	// Actual testing in Phase 5
}

func TestGetSession(t *testing.T) {
	// Test context extraction
	// Actual testing in Phase 5
}
EOF

# Run tests
go test ./src/auth/
```

## Test 6: Token Refresh Logic

### Manual Test with SQL

```sql
-- 1. Create a token that expires in 3 minutes
INSERT INTO auth_tokens (member_id, provider, access_token, refresh_token, token_type, expires_at, scopes)
VALUES (1, 'google', 'encrypted_old_token', 'encrypted_refresh_token', 'Bearer', 
        NOW() + INTERVAL 3 MINUTE, '["write:email"]');

-- 2. Call refresh endpoint (see Test 3 above)

-- 3. Check token was updated
SELECT member_id, provider, expires_at, updated_at 
FROM auth_tokens 
WHERE member_id = 1 AND provider = 'google';

-- expires_at should be updated to future time
-- updated_at should be recent
```

## Test 7: Error Scenarios

### Test 1: Refresh Without Session
```bash
curl -X POST "http://localhost:8080/auth/refresh"
# Expected: 401 "no session"
```

### Test 2: Refresh With Invalid Session
```bash
curl -X POST "http://localhost:8080/auth/refresh" \
  --cookie "session_id=invalid-session-id"
# Expected: 401 "invalid session"
```

### Test 3: Session Expired
```sql
-- Create expired session
INSERT INTO auth_sessions (id, member_id, data, created_at, expires_at, last_activity)
VALUES ('test-expired-2', 1, '{"scopes":["test"],"provider":"test"}', 
        NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR);
```

```bash
curl "http://localhost:8080/auth/session" \
  --cookie "session_id=test-expired-2"
# Expected: 401 "session not found or expired"
```

## Test 8: Environment in Production

### Generate Production Key
```bash
# Generate secure 64-character hex key
openssl rand -hex 32

# Output example:
# a3f8d9e2c1b4567890abcdef1234567890abcdef1234567890abcdef12345678

# Set in production environment
export AUTH_ENCRYPTION_KEY="a3f8d9e2c1b4567890abcdef1234567890abcdef1234567890abcdef12345678"

# Start server
go run src/sidan-backend.go

# Verify no warning in logs
```

### Verify Key Used
```bash
# The key should be 64 characters (32 bytes in hex)
echo $AUTH_ENCRYPTION_KEY | wc -c
# Expected: 65 (64 chars + newline)
```

## Test 9: Compilation

```bash
# Clean build
go clean
go build -o sidan-backend ./src/sidan-backend.go

# Check binary created
ls -lh sidan-backend

# Run binary
./sidan-backend
# Should start without errors
```

## Test 10: Code Quality Check

```bash
# Run go fmt
go fmt ./src/auth/middleware.go
go fmt ./src/router/auth.go
go fmt ./src/router/router.go

# Run go vet
go vet ./src/auth/middleware.go
go vet ./src/router/auth.go
go vet ./src/router/router.go

# No errors expected
```

## Success Criteria

- [ ] Server starts without errors
- [ ] Cleanup job logs "cleanup job started"
- [ ] Environment variable read correctly (or warning shown)
- [ ] Refresh endpoint returns success
- [ ] Expired sessions cleaned up after 15+ minutes
- [ ] Token updated_at timestamp changes after refresh
- [ ] Error responses correct for invalid inputs
- [ ] Code compiles without warnings
- [ ] go fmt shows no changes
- [ ] go vet shows no issues

## Next Steps

After all tests pass:
1. Commit Phase 4 changes
2. Review phase4-complete.md
3. Get product owner approval
4. Move to Phase 5 (migration and cleanup)

## Notes

- Middleware won't be fully tested until Phase 5 (when endpoints migrated)
- Token refresh logic tested via refresh endpoint
- Cleanup job requires 15+ minute wait (or modify interval for testing)
- Some tests require real OAuth2 tokens (Google/GitHub)
- Database should have at least one valid member for testing

## Common Issues

**Issue**: "Using default encryption key" warning  
**Fix**: Set AUTH_ENCRYPTION_KEY environment variable

**Issue**: Cleanup job not running  
**Fix**: Wait 15+ minutes or modify interval in code

**Issue**: Refresh fails with "no refresh token"  
**Fix**: Provider doesn't support refresh tokens, or token not requested with offline_access

**Issue**: "session not found"  
**Fix**: Session expired or invalid, re-login needed

**Issue**: Database connection errors  
**Fix**: Ensure MySQL running (`make` to start Docker container)
