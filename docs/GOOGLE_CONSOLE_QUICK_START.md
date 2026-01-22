# Google Console Setup for Device Flow - Quick Guide

## What You Need to Do in Google Console

### Step 1: Create OAuth2 Credentials
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **APIs & Services > Credentials**
3. Click **+ CREATE CREDENTIALS**
4. Select **OAuth client ID**
5. **IMPORTANT**: For Application type, select **"TVs and Limited Input devices"**
   - This is the key! Device flow requires this specific type
   - Don't select "Web application" or "Desktop app"
6. Enter name: "Sidan CLI Device Flow"
7. Click **Create**
8. Copy your Client ID and Client Secret

### Step 2: Configure OAuth Consent Screen (if not done)
1. Go to **APIs & Services > OAuth consent screen**
2. Choose **External** (or Internal for Google Workspace)
3. Fill in app name and contact emails
4. Add scopes: `openid`, `email`
5. Add test users (if in development mode)

### Step 3: Update Your Config
Edit `config/local.yaml`:
```yaml
oauth2:
  google:
    clientId: "YOUR_CLIENT_ID_HERE.apps.googleusercontent.com"
    clientSecret: "YOUR_CLIENT_SECRET_HERE"
    redirectURL: "http://localhost:8080/auth/callback"
    scopes: ["openid", "email"]
```

## Why "TVs and Limited Input devices"?

Google's OAuth2 device flow is specifically designed for:
- Smart TVs
- Game consoles  
- IoT devices
- **CLI tools** (our use case)
- Any device with limited input capabilities

This credential type enables the device flow protocol where:
1. Device shows a code (like ABCD-1234)
2. User enters code on another device (phone/computer)
3. Device receives token automatically

## Testing Your Setup

1. **Start the backend**:
   ```bash
   make              # Start database
   go run src/sidan-backend.go
   ```

2. **Build and run CLI**:
   ```bash
   go build -o sidan-auth ./cmd/sidan-auth
   ./sidan-auth token add google
   ```

3. **Verify you see**:
   - Verification URL displayed
   - 8-character code (e.g., ABCD-1234)
   - CLI waiting for authorization

4. **Complete flow**:
   - Open URL in browser
   - Sign in with Google
   - CLI receives token automatically

## Common Issues

### "Invalid client" error
- Check Client ID and Secret are correct
- Verify you used "TVs and Limited Input devices" type

### "Email not registered" error  
- User's email must exist in `cl2007_members` table
- Add member to database first

### Redirect URL doesn't matter
- Device flow doesn't use traditional redirects
- The redirect URL in config is required but not used by device flow
- Actual verification happens via `/auth/device/verify` endpoint

## Production Deployment

For production (`api.chalmerslosers.com`):
1. Create new OAuth2 credentials (same type: "TVs and Limited Input devices")
2. Update production config with production credentials
3. Users configure CLI: `sidan-auth config set-api https://api.chalmerslosers.com`

## Security Notes

- Keep Client Secret secure (don't commit to git)
- Device codes expire in 10 minutes
- JWT tokens expire in 8 hours
- Tokens stored with 0600 permissions (~/.sidan/config.json)

## More Information

See full documentation:
- Complete guide: `docs/DEVICE_FLOW_SETUP.md`
- CLI usage: `cmd/sidan-auth/README.md`
- API docs: `swagger.yaml`
