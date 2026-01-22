# sidan-auth CLI

Command-line tool for authenticating with the Sidan API using OAuth2 device flow.

## Installation

```bash
go build -o sidan-auth ./cmd/sidan-auth
```

Or install to your Go bin:
```bash
go install ./cmd/sidan-auth
```

## Quick Start

```bash
# Add a new token (default: google)
./sidan-auth token add

# Add a token for specific provider
./sidan-auth token add google

# List saved tokens
./sidan-auth token list

# Remove a token
./sidan-auth token remove google

# Show configuration
./sidan-auth config show

# Set API endpoint
./sidan-auth config set-api https://api.chalmerslosers.com
```

## Usage

### Adding a Token

```bash
$ ./sidan-auth token add google
Requesting device authorization for google...

==============================================
To authorize this device, visit:

    http://localhost:8080/auth/device/verify?code=ABCD-1234&provider=google

And enter this code:

    ABCD-1234

==============================================

Waiting for authorization...
```

Open the URL in a browser, sign in with Google, and authorize. The CLI will automatically receive and save your token.

### Using the Token

The token is saved to `~/.sidan/config.json`. Use it in API requests:

```bash
# Get your token
TOKEN=$(jq -r '.tokens.google' ~/.sidan/config.json)

# Make authenticated request
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/auth/session
```

Or export it as an environment variable:

```bash
export SIDAN_TOKEN=$(jq -r '.tokens.google' ~/.sidan/config.json)

curl -H "Authorization: Bearer $SIDAN_TOKEN" \
  http://localhost:8080/db/members
```

## Commands

### token add [provider]
Request a new authentication token using device flow.
- **provider**: OAuth2 provider (default: google)
- Displays verification URL and code
- Polls for authorization
- Saves token to config file

### token list
List all saved tokens by provider.

### token remove [provider]
Remove a saved token.
- **provider**: Provider to remove (required)

### config show
Show current configuration including:
- Config file path
- API endpoint
- Number of saved tokens

### config set-api [url]
Set the API endpoint URL.
- **url**: API base URL (required)
- Example: `https://api.chalmerslosers.com`

## Configuration

Configuration is stored in `~/.sidan/config.json`:

```json
{
  "api_endpoint": "http://localhost:8080",
  "tokens": {
    "google": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

File permissions are set to 0600 (owner read/write only) for security.

## Providers

Currently supported OAuth2 providers:
- **google**: Google OAuth2
- **github**: GitHub OAuth2

## Token Expiry

JWT tokens expire after 8 hours. When a token expires, simply run `sidan-auth token add` again to get a fresh token.

## Examples

### Local Development
```bash
# Default configuration uses localhost
./sidan-auth token add google
```

### Production Use
```bash
# Configure for production API
./sidan-auth config set-api https://api.chalmerslosers.com

# Add token
./sidan-auth token add google

# Verify it works
TOKEN=$(jq -r '.tokens.google' ~/.sidan/config.json)
curl -H "Authorization: Bearer $TOKEN" \
  https://api.chalmerslosers.com/auth/session
```

### Multiple Providers
```bash
# Add Google token
./sidan-auth token add google

# Add GitHub token
./sidan-auth token add github

# List all tokens
./sidan-auth token list
```

## Troubleshooting

### "Failed to request device code"
- Check that the backend server is running
- Verify API endpoint: `./sidan-auth config show`
- Update if needed: `./sidan-auth config set-api http://localhost:8080`

### "Authorization timeout"
- Device codes expire after 10 minutes
- Complete authorization within this time
- If expired, run `sidan-auth token add` again

### "Email not registered"
- Your email must exist in the members database
- Contact an admin to add your email to the system

### "Invalid token" when making API requests
- Token may have expired (8 hour lifetime)
- Run `./sidan-auth token add` to refresh
- Check token with: `./sidan-auth token list`

## Security

- Config file stored at `~/.sidan/config.json` with 0600 permissions
- Only the file owner can read/write tokens
- Tokens are JWT-based and expire after 8 hours
- Always use HTTPS in production (`https://api.chalmerslosers.com`)

## See Also

- [Device Flow Setup Guide](../docs/DEVICE_FLOW_SETUP.md) - Complete setup instructions for Google Console
- [API Documentation](../swagger.yaml) - Full API reference
