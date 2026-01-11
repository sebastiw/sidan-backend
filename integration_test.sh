#!/bin/bash

# Integration Test Suite for JWT Authentication and Secret Permissions
# Tests various scenarios with different users and permission levels

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
JWT_SECRET="my-test-secret"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test result tracking
declare -a FAILED_TEST_NAMES

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Integration Test Suite${NC}"
echo -e "${BLUE}JWT Authentication & Secret Permissions${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if backend is running
if ! curl -s http://localhost:8080/db/entries?take=1 > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Backend is not running on http://localhost:8080${NC}"
    echo -e "${YELLOW}Please start the backend first:${NC}"
    echo "  JWT_SECRET=\"$JWT_SECRET\" go run src/sidan-backend.go"
    exit 1
fi

echo -e "${GREEN}✓ Backend is running${NC}\n"

# Helper function to generate JWT token
generate_token() {
    local member_id=$1
    local email=$2
    
    TOKEN=$(JWT_SECRET="$JWT_SECRET" go run generate_test_jwt.go "$member_id" "$email" 2>&1 | \
        grep -A 1 "JWT Token:" | tail -1 | xargs)
    
    if [ -z "$TOKEN" ]; then
        echo -e "${RED}Failed to generate token for member $member_id${NC}"
        return 1
    fi
    
    echo "$TOKEN"
}

# Helper function to make authenticated request
api_request() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=$4
    
    if [ -z "$token" ]; then
        # Unauthenticated request
        if [ -z "$data" ]; then
            curl -s -w "\nHTTP_CODE:%{http_code}" -X "$method" \
                -H "Content-Type: application/json" \
                "$BASE_URL$endpoint"
        else
            curl -s -w "\nHTTP_CODE:%{http_code}" -X "$method" \
                -H "Content-Type: application/json" \
                -d "$data" \
                "$BASE_URL$endpoint"
        fi
    else
        # Authenticated request
        if [ -z "$data" ]; then
            curl -s -w "\nHTTP_CODE:%{http_code}" -X "$method" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                "$BASE_URL$endpoint"
        else
            curl -s -w "\nHTTP_CODE:%{http_code}" -X "$method" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                -d "$data" \
                "$BASE_URL$endpoint"
        fi
    fi
}

# Test assertion helper
assert_test() {
    local test_name=$1
    local expected_code=$2
    local actual_code=$3
    local response=$4
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$actual_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected: HTTP $expected_code, Got: HTTP $actual_code"
        if [ ! -z "$response" ]; then
            echo -e "  Response: ${response:0:100}"
        fi
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$test_name")
        return 1
    fi
}

# Test assertion for field value
assert_field() {
    local test_name=$1
    local field_name=$2
    local expected_value=$3
    local actual_value=$4
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$actual_value" = "$expected_value" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected $field_name: $expected_value, Got: $actual_value"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$test_name")
        return 1
    fi
}

echo -e "${CYAN}=== Generating Test Tokens ===${NC}\n"

# Generate tokens for different users
echo "Generating token for Member #8 (user 295)..."
TOKEN_MEMBER_8=$(generate_token 295 "max.gabrielsson@gmail.com")
echo -e "${GREEN}✓ Member #8 token generated${NC}\n"

echo "Generating token for Member #7 (user 294)..."
TOKEN_MEMBER_7=$(generate_token 294 "MarcBjork@rhyta.com")
echo -e "${GREEN}✓ Member #7 token generated${NC}\n"

echo "Generating token for Member #2 (user 290)..."
TOKEN_MEMBER_2=$(generate_token 290 "MorganBlom@dayrep.com")
echo -e "${GREEN}✓ Member #2 token generated${NC}\n"

# ============================================================================
echo -e "${CYAN}=== Authentication Tests ===${NC}\n"
# ============================================================================

# Test 1: Unauthenticated request to protected endpoint should fail
response=$(api_request "GET" "/auth/session" "")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
assert_test "Unauthenticated user should NOT access /auth/session" "401" "$http_code"

# Test 2: Authenticated user should access session endpoint
response=$(api_request "GET" "/auth/session" "$TOKEN_MEMBER_8")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')
assert_test "Authenticated user (Member #8) SHOULD access /auth/session" "200" "$http_code"

# Test 3: Member should have correct scopes
member_id=$(echo "$body" | python3 -c "import sys, json; print(json.load(sys.stdin)['member']['id'])" 2>/dev/null)
assert_field "Member #8 should have correct ID in session" "id" "295" "$member_id"

echo ""

# ============================================================================
echo -e "${CYAN}=== Entry Creation Tests ===${NC}\n"
# ============================================================================

# Test 4: Create a public entry as Member #8
public_entry_data='{
  "msg": "Public test message from Member #8 - Everyone should see this!",
  "sig": "#8",
  "email": "max.gabrielsson@gmail.com",
  "place": "Integration Test"
}'

response=$(api_request "POST" "/db/entries" "$TOKEN_MEMBER_8" "$public_entry_data")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')
public_entry_id=$(echo "$body" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

assert_test "Member #8 SHOULD create a public entry" "200" "$http_code"

if [ ! -z "$public_entry_id" ]; then
    echo -e "  ${GREEN}Created entry ID: $public_entry_id${NC}"
    
    # Verify the entry was created and appears first (take=1)
    response=$(api_request "GET" "/db/entries?take=1&skip=0" "")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    first_id=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data[0]['id'] if isinstance(data, list) and len(data) > 0 else '')" 2>/dev/null)
    
    assert_test "New entry SHOULD appear first in list (take=1)" "200" "$http_code"
    assert_field "First entry in list should be the new entry" "id" "$public_entry_id" "$first_id"
    
    # Verify secret=false
    secret=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(str(data[0]['secret']).lower() if isinstance(data, list) and len(data) > 0 else '')" 2>/dev/null)
    assert_field "Created public entry should have secret=false" "secret" "false" "$secret"
fi

echo ""

# Test 5: Create entry and make it secret to everyone (user_id=0)
secret_all_data='{
  "msg": "Secret message from Member #7 - Nobody should see this content!",
  "sig": "#7",
  "email": "MarcBjork@rhyta.com",
  "place": "Secret Location"
}'

response=$(api_request "POST" "/db/entries" "$TOKEN_MEMBER_7" "$secret_all_data")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')
secret_entry_id=$(echo "$body" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

assert_test "Member #7 SHOULD create an entry" "200" "$http_code"

if [ ! -z "$secret_entry_id" ]; then
    echo -e "  ${GREEN}Created entry ID: $secret_entry_id${NC}"
    
    # Add permission record to make it secret (user_id=0)
    echo -e "  ${YELLOW}Adding permission to make entry secret to everyone (user_id=0)...${NC}"
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e \
        "INSERT INTO cl2003_permissions (id, user_id) VALUES ($secret_entry_id, 0)" 2>/dev/null
    
    # Verify the entry is now marked as secret
    response=$(api_request "GET" "/db/entries/$secret_entry_id" "$TOKEN_MEMBER_7")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    secret=$(echo "$body" | python3 -c "import sys, json; print(str(json.load(sys.stdin)['secret']).lower())" 2>/dev/null)
    personal_secret=$(echo "$body" | python3 -c "import sys, json; print(str(json.load(sys.stdin)['personal_secret']).lower())" 2>/dev/null)
    
    assert_test "Entry with permission SHOULD be marked as secret" "200" "$http_code"
    assert_field "Secret entry (user_id=0) should have secret=true" "secret" "true" "$secret"
    assert_field "Secret entry (user_id=0) should have personal_secret=false" "personal_secret" "false" "$personal_secret"
fi

echo ""

# Test 6: Create entry with personal secret for specific members
personal_secret_data='{
  "msg": "Personal secret from Member #2 - Only for Members #7 and #8!",
  "sig": "#2",
  "email": "MorganBlom@dayrep.com",
  "place": "Private Chat"
}'

response=$(api_request "POST" "/db/entries" "$TOKEN_MEMBER_2" "$personal_secret_data")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')
personal_entry_id=$(echo "$body" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

assert_test "Member #1 SHOULD create an entry" "200" "$http_code"

if [ ! -z "$personal_entry_id" ]; then
    echo -e "  ${GREEN}Created entry ID: $personal_entry_id${NC}"
    
    # Add permission records for Members #7 and #8
    echo -e "  ${YELLOW}Adding permissions for Members #7 and #8...${NC}"
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e \
        "INSERT INTO cl2003_permissions (id, user_id) VALUES ($personal_entry_id, 7), ($personal_entry_id, 8)" 2>/dev/null
    
    # Verify the entry is marked as personal secret
    response=$(api_request "GET" "/db/entries/$personal_entry_id" "$TOKEN_MEMBER_2")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    secret=$(echo "$body" | python3 -c "import sys, json; print(str(json.load(sys.stdin)['secret']).lower())" 2>/dev/null)
    personal_secret=$(echo "$body" | python3 -c "import sys, json; print(str(json.load(sys.stdin)['personal_secret']).lower())" 2>/dev/null)
    
    assert_test "Entry with user permissions SHOULD be marked as secret" "200" "$http_code"
    assert_field "Personal secret should have secret=true" "secret" "true" "$secret"
    assert_field "Personal secret should have personal_secret=true" "personal_secret" "true" "$personal_secret"
    
    # Verify Member #7 can access
    response=$(api_request "GET" "/db/entries/$personal_entry_id" "$TOKEN_MEMBER_7")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    assert_test "Member #7 SHOULD access personal secret entry (has permission)" "200" "$http_code"
    
    # Verify Member #8 can access
    response=$(api_request "GET" "/db/entries/$personal_entry_id" "$TOKEN_MEMBER_8")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    assert_test "Member #8 SHOULD access personal secret entry (has permission)" "200" "$http_code"
fi

echo ""

# ============================================================================
echo -e "${CYAN}=== Created Entries in List Verification ===${NC}\n"
# ============================================================================

# Test: Verify all created entries appear in the most recent list
response=$(api_request "GET" "/db/entries?take=5&skip=0" "$TOKEN_MEMBER_8")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

assert_test "List recent entries SHOULD return successfully" "200" "$http_code"

# Check if our entries are in the list with proper flags
if [ ! -z "$public_entry_id" ]; then
    contains_public=$(echo "$body" | grep -o "\"id\":$public_entry_id" | wc -l)
    if [ "$contains_public" -ge 1 ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Public entry (ID: $public_entry_id) appears in recent entries"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Public entry appears in recent entries")
        echo -e "${RED}✗ FAIL${NC}: Public entry should appear in list"
    fi
fi

if [ ! -z "$secret_entry_id" ]; then
    contains_secret=$(echo "$body" | grep -o "\"id\":$secret_entry_id" | wc -l)
    if [ "$contains_secret" -ge 1 ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Secret entry (ID: $secret_entry_id) appears in recent entries"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Secret entry appears in recent entries")
        echo -e "${RED}✗ FAIL${NC}: Secret entry should appear in list"
    fi
fi

if [ ! -z "$personal_entry_id" ]; then
    contains_personal=$(echo "$body" | grep -o "\"id\":$personal_entry_id" | wc -l)
    if [ "$contains_personal" -ge 1 ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Personal secret entry (ID: $personal_entry_id) appears in recent entries"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Personal secret entry appears in recent entries")
        echo -e "${RED}✗ FAIL${NC}: Personal secret should appear in list"
    fi
fi

echo ""

# ============================================================================
echo -e "${CYAN}=== Message Content Filtering Tests ===${NC}\n"
# ============================================================================

# Test: Unauthenticated user should see "hemlis" for personal secret message
if [ ! -z "$personal_entry_id" ]; then
    response=$(api_request "GET" "/db/entries/$personal_entry_id" "")
    msg=$(echo "$response" | jq -r '.msg')
    
    if [ "$msg" = "hemlis" ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Unauthenticated user sees 'hemlis' for personal secret"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Unauthenticated access to personal secret shows hemlis")
        echo -e "${RED}✗ FAIL${NC}: Unauthenticated user should see 'hemlis', got: $msg"
    fi
fi

# Test: Unauthorized user (Member #8 not in permission list) should see "hemlis"
# Create entry visible only to Member #7
echo -e "  ${YELLOW}Creating entry visible only to Member #7...${NC}"
restricted_entry=$(api_request "POST" "/db/entries" "$TOKEN_MEMBER_2" '{
    "msg": "This is a restricted message for Member #7 only",
    "sig": "#2"
}')
restricted_entry_id=$(echo "$restricted_entry" | jq -r '.id')

if [ ! -z "$restricted_entry_id" ] && [ "$restricted_entry_id" != "null" ]; then
    echo -e "  ${GREEN}Created restricted entry ID: $restricted_entry_id${NC}"
    
    # Add permission for only Member #7 (user_id 7)
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e \
        "INSERT INTO cl2003_permissions (id, user_id) VALUES ($restricted_entry_id, 7)" 2>/dev/null
    
    # Test: Member #8 (unauthorized) should see "hemlis"
    response=$(api_request "GET" "/db/entries/$restricted_entry_id" "$TOKEN_MEMBER_8")
    msg=$(echo "$response" | jq -r '.msg')
    
    if [ "$msg" = "hemlis" ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Unauthorized user (Member #8) sees 'hemlis'"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Unauthorized user sees hemlis")
        echo -e "${RED}✗ FAIL${NC}: Member #8 should see 'hemlis', got: $msg"
    fi
    
    # Test: Member #7 (authorized) should see full message with prefix
    response=$(api_request "GET" "/db/entries/$restricted_entry_id" "$TOKEN_MEMBER_7")
    msg=$(echo "$response" | jq -r '.msg')
    
    if echo "$msg" | grep -q "hemlis Till #7"; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Authorized user (Member #7) sees message with prefix"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Authorized user sees message with prefix")
        echo -e "${RED}✗ FAIL${NC}: Member #7 should see message with 'hemlis Till #7' prefix"
        echo -e "  Got: $msg"
    fi
    
    # Test: Author (Member #1) should see full message with prefix
    response=$(api_request "GET" "/db/entries/$restricted_entry_id" "$TOKEN_MEMBER_2")
    msg=$(echo "$response" | jq -r '.msg')
    
    if echo "$msg" | grep -q "This is a restricted message"; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Author (Member #1) sees full message"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Author sees full message")
        echo -e "${RED}✗ FAIL${NC}: Author should see full message"
        echo -e "  Got: $msg"
    fi
fi

# Test: Secret to everyone (user_id=0) should be visible to all
if [ ! -z "$secret_entry_id" ]; then
    response=$(api_request "GET" "/db/entries/$secret_entry_id" "")
    msg=$(echo "$response" | jq -r '.msg')
    
    if [ "$msg" != "hemlis" ] && [ ! -z "$msg" ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ PASS${NC}: Secret to everyone (user_id=0) visible to unauthenticated"
    else
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Secret to everyone visible to all")
        echo -e "${RED}✗ FAIL${NC}: Secret to everyone should be visible, got: $msg"
    fi
fi

echo ""

# ============================================================================
echo -e "${CYAN}=== Cleanup Created Entries ===${NC}\n"
# ============================================================================

# Clean up test entries
if [ ! -z "$public_entry_id" ]; then
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_msgs WHERE id=$public_entry_id" 2>/dev/null
    echo -e "${GREEN}✓${NC} Cleaned up public entry (ID: $public_entry_id)"
fi

if [ ! -z "$secret_entry_id" ]; then
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_permissions WHERE id=$secret_entry_id" 2>/dev/null
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_msgs WHERE id=$secret_entry_id" 2>/dev/null
    echo -e "${GREEN}✓${NC} Cleaned up secret entry (ID: $secret_entry_id)"
fi

if [ ! -z "$personal_entry_id" ]; then
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_permissions WHERE id=$personal_entry_id" 2>/dev/null
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_msgs WHERE id=$personal_entry_id" 2>/dev/null
    echo -e "${GREEN}✓${NC} Cleaned up personal secret entry (ID: $personal_entry_id)"
fi

if [ ! -z "$restricted_entry_id" ]; then
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_permissions WHERE id=$restricted_entry_id" 2>/dev/null
    docker exec sidan_sql mysql -uroot -pdbpassword dbschema -e "DELETE FROM cl2003_msgs WHERE id=$restricted_entry_id" 2>/dev/null
    echo -e "${GREEN}✓${NC} Cleaned up restricted entry (ID: $restricted_entry_id)"
fi

echo ""

# ============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Results Summary${NC}"
echo -e "${BLUE}========================================${NC}\n"
# ============================================================================

echo -e "Total Tests:  ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}ALL TESTS PASSED! ✓${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}FAILED TESTS:${NC}"
    echo -e "${RED}========================================${NC}"
    for test_name in "${FAILED_TEST_NAMES[@]}"; do
        echo -e "${RED}  ✗ $test_name${NC}"
    done
    echo ""
    exit 1
fi
