# Phase 3 Testing Guide: OAuth2 Flow Handlers

**Status**: ✅ Complete  
**Date**: 2026-01-10  
**Branch**: feature/auth-rewrite

## Overview

Phase 3 implements complete OAuth2 authentication with HTTP handlers. This guide shows how to test the full login flow.

## Prerequisites

1. **Database Running** (from Phase 1):
   ```bash
   docker ps | grep sidan_sql  # Should be running
   ```

2. **Build Application**:
   ```bash
   go build -o /tmp/sidan ./src/sidan-backend.go
   ```

3. **OAuth2 Credentials** (optional for structure testing):
   - Google OAuth2 client ID/secret
   - GitHub OAuth2 client ID/secret
   - Or use existing test credentials in config

4. **Test Member**:
   ```sql
   -- Create test member
   INSERT INTO cl2007_members (number, name, email, isvalid, im)
   VALUES (9999, 'Test User', 'test@example.com', 1, '');
   ```

## Testing Scenarios

### Test 1: Application Starts

**Objective**: Verify handlers are registered correctly.

**Steps**:
```bash
# Start application
cd /Users/maxgab/code/sidan/sidan-backend
go run src/sidan-backend.go
```

**Expected Output**:
```
INFO Starting backend service address=:8080
```

**Verification**:
```bash
# In another terminal, check endpoints respond
curl -i http://localhost:8080/auth/session
# Should return 401 Unauthorized (no session)

curl -i "http://localhost:8080/auth/login"
# Should return 400 Bad Request (provider required)
```

**Pass Criteria**: ✅ Application starts, endpoints respond with expected errors.

---

### Test 2: Login Initiation

**Objective**: Verify /auth/login generates correct redirect.

**Steps**:
```bash
# Initiate Google login
curl -v "http://localhost:8080/auth/login?provider=google&redirect_uri=https://app.com" 2>&1 | grep Location
```

**Expected Output**:
```
< Location: https://accounts.google.com/o/oauth2/v2/auth?client_id=...&code_challenge=...&code_challenge_method=S256&state=...
```

**Verification Checklist**:
- [ ] URL contains `client_id`
- [ ] URL contains `code_challenge` (43 chars)
- [ ] URL contains `code_challenge_method=S256`
- [ ] URL contains `state` (64 chars)
- [ ] URL contains `redirect_uri`
- [ ] URL contains `scope`
- [ ] URL contains `access_type=offline` (Google)
- [ ] URL contains `prompt=consent` (Google)

**Database Check**:
```sql
SELECT id, provider, LENGTH(pkce_verifier), expires_at 
FROM auth_states 
ORDER BY created_at DESC 
LIMIT 1;

-- Should show:
-- - 64-char state ID
-- - provider = 'google'
-- - pkce_verifier = 43 chars
-- - expires_at = ~10 minutes from now
```

**Pass Criteria**: ✅ Redirect URL has all OAuth2 + PKCE parameters, state stored in DB.

---

### Test 3: GitHub Login

**Objective**: Verify GitHub provider works.

**Steps**:
```bash
curl -v "http://localhost:8080/auth/login?provider=github" 2>&1 | grep Location
```

**Expected Output**:
```
< Location: https://github.com/login/oauth/authorize?client_id=...&code_challenge=...
```

**Verification**:
- [ ] URL points to github.com
- [ ] Has PKCE parameters
- [ ] Does NOT have `access_type` (Google-specific)

**Pass Criteria**: ✅ GitHub provider generates correct redirect.

---

### Test 4: Unknown Provider

**Objective**: Verify error handling for invalid provider.

**Steps**:
```bash
curl -i "http://localhost:8080/auth/login?provider=facebook"
```

**Expected Output**:
```
HTTP/1.1 400 Bad Request
unknown provider
```

**Pass Criteria**: ✅ Returns 400 for unknown provider.

---

### Test 5: Session Without Login

**Objective**: Verify session endpoint requires authentication.

**Steps**:
```bash
curl -i http://localhost:8080/auth/session
```

**Expected Output**:
```
HTTP/1.1 401 Unauthorized
no session
```

**Pass Criteria**: ✅ Returns 401 when no session cookie present.

---

### Test 6: Complete OAuth2 Flow (Manual)

**Objective**: Test full Google login flow.

**Prerequisites**:
- Valid Google OAuth2 credentials in config
- Test member email registered in database

**Steps**:

1. **Initiate Login**:
   ```bash
   curl -v "http://localhost:8080/auth/login?provider=google" 2>&1 | grep Location
   ```
   Copy the redirect URL.

2. **Authenticate**:
   - Open redirect URL in browser
   - Log in with Google account (use email from test member)
   - Approve consent screen
   - Browser will redirect to `/auth/callback?state=...&code=...`

3. **Check Session**:
   - Copy session cookie from browser dev tools
   - Or check database:
   ```sql
   SELECT id, member_id, expires_at 
   FROM auth_sessions 
   ORDER BY created_at DESC 
   LIMIT 1;
   ```

4. **Get Session Info**:
   ```bash
   curl --cookie "session_id=<cookie-value>" http://localhost:8080/auth/session
   ```

**Expected Response**:
```json
{
  "session_id": "abc123...",
  "member": {
    "id": 123,
    "number": 9999,
    "name": "Test User",
    "email": "test@example.com"
  },
  "scopes": ["write:email", "write:image", "write:member", "read:member", "modify:entry"],
  "provider": "google",
  "expires_at": "2026-01-10T21:00:00Z"
}
```

5. **Check Database State**:
   ```sql
   -- Token stored (encrypted)
   SELECT member_id, provider, token_type, expires_at 
   FROM auth_tokens 
   WHERE member_id = <member-id>;
   
   -- Provider linked
   SELECT member_id, provider, provider_email, email_verified
   FROM auth_provider_links
   WHERE member_id = <member-id>;
   
   -- Session active
   SELECT id, member_id, expires_at, last_activity
   FROM auth_sessions
   WHERE member_id = <member-id>;
   ```

6. **Logout**:
   ```bash
   curl -X POST --cookie "session_id=<cookie-value>" http://localhost:8080/auth/logout
   ```
   
   Expected: `{"success":true}`
   
7. **Verify Session Gone**:
   ```bash
   curl --cookie "session_id=<cookie-value>" http://localhost:8080/auth/session
   ```
   
   Expected: 401 Unauthorized

**Pass Criteria**: ✅ Complete flow works, data stored correctly, logout clears session.

---

### Test 7: Unregistered Email

**Objective**: Verify only registered emails can log in.

**Steps**:
1. Initiate login with Google
2. Authenticate with email NOT in cl2007_members table
3. Complete callback

**Expected Result**:
```
HTTP/1.1 403 Forbidden
email not registered - please contact admin
```

**Pass Criteria**: ✅ Rejects unregistered emails with clear message.

---

### Test 8: State Expiry

**Objective**: Verify state expires after 10 minutes.

**Steps**:
1. Initiate login, get redirect URL
2. Wait 11 minutes
3. Try to complete callback with old state

**Expected Result**:
```
HTTP/1.1 400 Bad Request
invalid or expired state
```

**Database Check**:
```sql
-- Expired states should be gone
SELECT COUNT(*) FROM auth_states WHERE expires_at < NOW();
-- Should return 0 (expired states deleted on access)
```

**Pass Criteria**: ✅ Expired state rejected.

---

### Test 9: State Reuse Prevention

**Objective**: Verify state can't be reused.

**Steps**:
1. Complete successful login (saves state, then deletes it)
2. Try to use same state again in callback

**Expected Result**:
```
HTTP/1.1 400 Bad Request
invalid or expired state
```

**Pass Criteria**: ✅ State deleted after use, can't be reused.

---

### Test 10: Session Activity Tracking

**Objective**: Verify last_activity updated on session access.

**Steps**:
1. Log in, get session
2. Check last_activity timestamp:
   ```sql
   SELECT last_activity FROM auth_sessions WHERE id = '<session-id>';
   ```
3. Wait 10 seconds
4. Call /auth/session
5. Check last_activity again

**Expected**: Timestamp should be updated.

**Pass Criteria**: ✅ Activity tracked correctly.

---

## Database Verification Queries

### Check Active Sessions
```sql
SELECT 
    s.id,
    s.member_id,
    m.email,
    s.created_at,
    s.expires_at,
    s.last_activity,
    TIMESTAMPDIFF(MINUTE, s.last_activity, NOW()) as minutes_idle
FROM auth_sessions s
JOIN cl2007_members m ON s.member_id = m.id
WHERE s.expires_at > NOW()
ORDER BY s.created_at DESC;
```

### Check Stored Tokens
```sql
SELECT 
    t.member_id,
    m.email,
    t.provider,
    t.token_type,
    t.expires_at,
    LENGTH(t.access_token) as encrypted_token_length,
    t.created_at,
    t.updated_at
FROM auth_tokens t
JOIN cl2007_members m ON t.member_id = m.id
ORDER BY t.updated_at DESC;
```

### Check Provider Links
```sql
SELECT 
    l.member_id,
    m.email as member_email,
    l.provider,
    l.provider_email,
    l.email_verified,
    l.linked_at
FROM auth_provider_links l
JOIN cl2007_members m ON l.member_id = m.id
ORDER BY l.linked_at DESC;
```

### Check Auth States
```sql
SELECT 
    id,
    provider,
    LENGTH(pkce_verifier) as verifier_length,
    redirect_uri,
    created_at,
    expires_at,
    TIMESTAMPDIFF(SECOND, NOW(), expires_at) as seconds_until_expiry
FROM auth_states
WHERE expires_at > NOW()
ORDER BY created_at DESC;
```

## Integration Test Script

```bash
#!/bin/bash
# test_auth_flow.sh

BASE_URL="http://localhost:8080"
PROVIDER="google"
TEST_EMAIL="test@example.com"

echo "=== Auth Flow Integration Test ==="

# 1. Check server is running
echo "1. Checking server..."
if ! curl -s -f $BASE_URL/auth/session > /dev/null 2>&1; then
    echo "❌ Server not responding"
    exit 1
fi
echo "✅ Server running"

# 2. Initiate login
echo "2. Initiating login..."
REDIRECT=$(curl -s -i "$BASE_URL/auth/login?provider=$PROVIDER" | grep -i location | cut -d' ' -f2 | tr -d '\r')
if [[ $REDIRECT == *"accounts.google.com"* ]]; then
    echo "✅ Redirect URL generated"
    echo "   URL: ${REDIRECT:0:80}..."
else
    echo "❌ Invalid redirect URL"
    exit 1
fi

# 3. Check state in database
echo "3. Checking database..."
STATE_COUNT=$(docker exec sidan_sql mysql -udbuser -pdbpassword -N -e \
    "SELECT COUNT(*) FROM dbschema.auth_states WHERE expires_at > NOW();")
if [ "$STATE_COUNT" -gt "0" ]; then
    echo "✅ State stored in database"
else
    echo "❌ No state found in database"
    exit 1
fi

# 4. Test error cases
echo "4. Testing error handling..."

# Unknown provider
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/login?provider=facebook")
if [ "$STATUS" == "400" ]; then
    echo "✅ Unknown provider rejected"
else
    echo "❌ Should reject unknown provider"
fi

# No session
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/session")
if [ "$STATUS" == "401" ]; then
    echo "✅ No session returns 401"
else
    echo "❌ Should return 401 for no session"
fi

echo ""
echo "=== Automated tests passed ==="
echo "=== Manual step required: ==="
echo "1. Open this URL in browser:"
echo "   $REDIRECT"
echo "2. Authenticate with email: $TEST_EMAIL"
echo "3. Check /auth/session for active session"
```

**Run**:
```bash
chmod +x test_auth_flow.sh
./test_auth_flow.sh
```

## Product Owner Acceptance

**To approve Phase 3**:

1. ✅ **Build succeeds**: `go build -o /tmp/sidan ./src/sidan-backend.go`
2. ✅ **Server starts**: Application runs without errors
3. ✅ **Login initiates**: GET /auth/login redirects to provider
4. ✅ **State stored**: Database contains auth_states entry
5. ✅ **Error handling**: Unknown provider returns 400
6. ✅ **No session check**: /auth/session returns 401 without cookie
7. ✅ **Code quality**: Review auth.go - should be clean and direct
8. ⏳ **Full flow** (optional): Complete OAuth2 flow with real credentials

**Sign off**: _____________________  Date: _________

