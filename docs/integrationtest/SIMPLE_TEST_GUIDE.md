# Simple JWT Test Guide - Start Here!

## The Problem You Just Hit

Your test failed with **"invalid token"** because:

❌ The backend generates a **random JWT secret** on every restart  
❌ Your token was signed with a **different secret**  
❌ The backend can't validate it  

## The Fix (2 Steps)

### Step 1: Start Backend with Fixed Secret

```bash
# Terminal 1
JWT_SECRET="my-test-secret" go run src/sidan-backend.go
```

### Step 2: Generate Token with SAME Secret & Test

```bash
# Terminal 2
JWT_SECRET="my-test-secret" go run generate_test_jwt.go 1 test@example.com

# Copy the export command from the output above, then:
export SIDAN_JWT='eyJhbGci...'

# Run tests
./test_jwt_auth.sh
```

**✅ Now it will work!** Both backend and token use `"my-test-secret"`

---

## Even Simpler: One-Line Per Terminal

```bash
# Terminal 1: Start backend
JWT_SECRET="dev-secret" go run src/sidan-backend.go

# Terminal 2: Generate & test in one go
JWT_SECRET="dev-secret" go run generate_test_jwt.go 1 test@example.com | tail -3 | head -1 | bash && ./test_jwt_auth.sh
```

---

## What You're Testing

When the test works, you'll verify:

✅ Backend reads JWT from `Authorization: Bearer` header  
✅ Backend validates signature and expiry  
✅ Backend loads member from database  
✅ Backend injects user into request context  
✅ Handlers can access user via `auth.GetMember(r)`  
✅ Scope checks work (403 if permission missing)  

---

## For Production / Browser Testing

When you get tokens from your browser (after real login), the JWT_SECRET must match what the backend uses.

**Set it in config/local.yaml:**
```yaml
jwt:
  secret: "your-production-secret-here"
  expiryHours: 8
```

**Or use environment variable:**
```bash
JWT_SECRET="your-production-secret" go run src/sidan-backend.go
```

**Browser tokens will then work:**
```bash
# Get token from browser DevTools (F12 → Application → localStorage)
export SIDAN_JWT='<browser_token>'
./test_jwt_auth.sh
```

---

## Quick Reference

| What | Command |
|------|---------|
| Start backend | `JWT_SECRET="test" go run src/sidan-backend.go` |
| Generate token | `JWT_SECRET="test" go run generate_test_jwt.go 1 test@example.com` |
| Export token | `export SIDAN_JWT='<token>'` |
| Run tests | `./test_jwt_auth.sh` |

**Remember:** Same `JWT_SECRET` for both backend and token generator!

