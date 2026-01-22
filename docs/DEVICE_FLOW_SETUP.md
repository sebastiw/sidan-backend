# Device Flow Authentication Setup

## Overview
This guide explains how to set up Google OAuth2 device flow authentication for the Sidan CLI tool. Device flow enables authentication on devices with limited input capabilities (like CLIs, IoT devices, etc.) or where typing a full URL is inconvenient.

## What is Device Flow?
Device flow is an OAuth2 extension designed for devices that either:
- Lack a web browser (CLI tools, headless servers)
- Have limited input capabilities (smart TVs, IoT devices)
- Make it inconvenient to type URLs (command-line interfaces)

The flow works by:
1. Device requests a verification code
2. User visits a web page on another device (phone, computer)
3. User enters the code and authorizes
4. Device receives the access token

## Google Cloud Console Setup

### 1. Create or Select a Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Note your project name

### 2. Enable Required APIs
1. Navigate to **APIs & Services > Library**
2. Search for and enable:
   - **Google+ API** (for user info)
   - **OAuth2 API**

### 3. Configure OAuth Consent Screen
1. Go to **APIs & Services > OAuth consent screen**
2. Choose **External** user type (or Internal if using Google Workspace)
3. Fill in the required fields:
   - **App name**: "Sidan API CLI"
   - **User support email**: Your email
   - **Developer contact email**: Your email
4. Add scopes:
   - `.../auth/userinfo.email`
   - `.../auth/userinfo.profile`
   - `openid`
5. Click **Save and Continue**
6. Add test users if using External type (in development mode)
7. Click **Save and Continue** until finished

### 4. Create OAuth 2.0 Credentials

**Important**: Google's device flow works with the **TV and Limited Input** device type.

1. Go to **APIs & Services > Credentials**
2. Click **+ CREATE CREDENTIALS** at the top
3. Select **OAuth client ID**
4. For **Application type**, select:
   - **TVs and Limited Input devices** (this enables device flow)
5. Enter a name: "Sidan CLI Device Flow"
6. Click **Create**
7. Save the credentials:
   - **Client ID**: Copy this value
   - **Client Secret**: Copy this value

### 5. Configure Backend

Update your `config/local.yaml` (or production config) with the device flow credentials:

```yaml
oauth2:
  google:
    clientId: "YOUR_CLIENT_ID_HERE"
    clientSecret: "YOUR_CLIENT_SECRET_HERE"
    redirectURL: "http://localhost:8080/auth/callback"  # Not used for device flow but required
    scopes: ["openid", "email"]
```

**Note**: For device flow, the `redirectURL` is still required in the config but not actually used in the device flow process. The actual verification happens via the `/auth/device/verify` endpoint.

## Testing Device Flow

### 1. Start the Backend
```bash
# Start database
make

# Start server
go run src/sidan-backend.go
```

The server should be running on `http://localhost:8080`.

### 2. Build the CLI Tool
```bash
go build -o sidan-auth ./cmd/sidan-auth
```

### 3. Test Device Flow
```bash
# Request device authorization
./sidan-auth token add google
```

You should see output like:
```
Requesting device authorization for google...

==============================================
To authorize this device, visit:

    http://localhost:8080/auth/device/verify?code=ABCD-1234&provider=google

And enter this code:

    ABCD-1234

==============================================

Waiting for authorization...
```

### 4. Complete Authorization
1. Open the verification URL in a browser (on any device)
2. Click "Sign in with google"
3. Complete the Google OAuth flow
4. You'll be redirected back and see "Success! Your device has been authorized"
5. The CLI will automatically receive the token and save it

### 5. Verify Token is Saved
```bash
# List saved tokens
./sidan-auth token list

# Show config
./sidan-auth config show
```

## Production Configuration

### Update API Endpoint
For production, configure the CLI to use the production API:

```bash
./sidan-auth config set-api https://api.chalmerslosers.com
```

### Update Backend Configuration
Update your production `config/production.yaml`:

```yaml
server:
  port: 8080

oauth2:
  google:
    clientId: "YOUR_PRODUCTION_CLIENT_ID"
    clientSecret: "YOUR_PRODUCTION_CLIENT_SECRET"
    redirectURL: "https://api.chalmerslosers.com/auth/callback"
    scopes: ["openid", "email"]
```

### CORS Configuration
Ensure the device verification page can redirect properly. The CORS settings in `src/router/router.go` already allow the required origins.

## Security Considerations

1. **Client Secret**: Keep your OAuth2 client secret secure. Don't commit it to version control.
   - Use environment variables: `GOOGLE_CLIENT_SECRET`
   - Or use a secrets management system

2. **Token Storage**: Tokens are stored in `~/.sidan/config.json` with 0600 permissions (owner read/write only)

3. **Token Expiry**: JWT tokens expire after 8 hours. Use the `sidan-auth token add` command to refresh.

4. **Device Code Expiry**: Device codes expire after 10 minutes. Users must complete authorization within this time.

5. **HTTPS**: In production, always use HTTPS for the API endpoint to protect tokens in transit.

## Troubleshooting

### "Invalid client" Error
- Verify your Client ID and Client Secret are correct in the config
- Make sure you created credentials for "TVs and Limited Input devices"

### "Email not registered" Error
- The user's email must exist in the `cl2007_members` table
- Check that the email in the members table matches the Google account email

### Device Code Expired
- Device codes expire after 10 minutes
- Run `sidan-auth token add` again to get a new code

### CLI Can't Connect to API
- Check that the backend is running
- Verify the API endpoint: `./sidan-auth config show`
- Update if needed: `./sidan-auth config set-api http://localhost:8080`

## Alternative: Using with Production API

Once deployed to production:

```bash
# Configure for production
./sidan-auth config set-api https://api.chalmerslosers.com

# Add token
./sidan-auth token add google
```

The verification URL will be:
```
https://api.chalmerslosers.com/auth/device/verify?code=XXXX-XXXX&provider=google
```

## API Documentation

### Request Device Code
```bash
POST /auth/device?provider=google
```

Response:
```json
{
  "device_code": "AH-1Ng4LRmXu...",
  "user_code": "ABCD-1234",
  "verification_uri": "http://localhost:8080/auth/device/verify?code=ABCD-1234&provider=google",
  "expires_in": 600,
  "interval": 5
}
```

### Poll for Token
```bash
POST /auth/device/token
Content-Type: application/json

{
  "device_code": "AH-1Ng4LRmXu...",
  "grant_type": "urn:ietf:params:oauth:grant-type:device_code"
}
```

Response (when authorized):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 28800
}
```

Response (still pending):
```json
{
  "error": "authorization_pending",
  "error_description": "User has not yet authorized the device"
}
```

## Using the Token

Once you have a token, use it with API requests:

```bash
# Get your token
TOKEN=$(jq -r '.tokens.google' ~/.sidan/config.json)

# Make authenticated request
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/auth/session
```

## References
- [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)
- [Google Identity - OAuth 2.0 for TV and Limited-Input Devices](https://developers.google.com/identity/protocols/oauth2/limited-input-device)
- [Google Cloud Console](https://console.cloud.google.com/)
