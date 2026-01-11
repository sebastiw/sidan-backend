# Simple UI for Sidan Forum

A single-page application for reading and posting forum entries.

## Features

### ✅ Implemented
- **API URL Selector** - Switch between production and local development
- **Token Authentication** - Paste JWT Bearer token to authenticate
- **Read Entries** - View all public entries and authorized secret entries
- **Create Entries** - Post new messages (requires authentication)
- **Pagination** - Browse entries 20 at a time
- **Secret Badges** - Visual indicators for secret and personal secret entries
- **Hemlis Protection** - Unauthorized personal secrets show only "hemlis"
- **Local Storage** - Token and API URL persist across page reloads

### 🎨 UI Elements
- **API URL selector** dropdown (production/local)
- Token input field at the top
- Authentication status indicator
- Create entry form (only visible when authenticated)
- Paginated entry list with badges
- Clean, gradient design

### 📋 Entry Display
Each entry shows:
- **Signature** (if authorized)
- **Date & Time**
- **Location** (if provided and authorized)
- **Message** (full text or "hemlis" if unauthorized)
- **Email** (if provided and authorized)
- **Badges**: SECRET, PERSONAL, ❤️ likes count

## Usage

### 1. Open the UI
Open `simple-ui/index.html` in your browser:
```bash
# Option 1: Direct file
open simple-ui/index.html

# Option 2: Simple HTTP server (recommended)
cd simple-ui
python3 -m http.server 8081
# Then visit http://localhost:8081
```

### 2. Select API Server
Use the dropdown at the top:
- **Production**: `https://api.chalmerslosers.com` (default)
- **Local Development**: `http://localhost:8080`

### 3. Start Local Backend (Optional)
Only needed if using "Local Development" mode:
```bash
cd /home/max/code/sidan-backend
JWT_SECRET='my-test-secret' ./sidan-backend
```

### 4. Get a Token (Optional - for posting)
Generate a test token:
```bash
JWT_SECRET='my-test-secret' go run generate_test_jwt.go 295 "max.gabrielsson@gmail.com"
```

Copy the token and paste it into the token field at the top of the page.

### 4. Browse & Post
- **Select API server** from dropdown (Production or Local)
- **Without token**: You can read all public entries
- **With token**: You can also see authorized secrets and post new entries

## API Endpoints Used

### Read Operations (No auth required)
- `GET /db/entries?skip=0&take=20` - List entries with pagination

### Write Operations (Auth required)
- `POST /db/entries` - Create new entry

## Field Redaction

When viewing unauthorized personal secrets:
- Message: "hemlis"
- Signature: Hidden
- Email: Hidden
- Location: Hidden
- Likes: 0
- All other fields: Cleared

## Configuration

The UI includes an **API server selector** with two options:
- **Production**: `https://api.chalmerslosers.com` (default)
- **Local Development**: `http://localhost:8080`

The selected API URL is saved in localStorage and persists across page reloads.

To add more API servers, edit the dropdown in `index.html`:
```html
<select id="apiSelector" onchange="changeApiUrl()">
    <option value="https://api.chalmerslosers.com">Production</option>
    <option value="http://localhost:8080">Local Development</option>
    <option value="https://staging.example.com">Staging</option>  <!-- Add more here -->
</select>
```

## Browser Compatibility
- Modern browsers (Chrome, Firefox, Safari, Edge)
- Requires ES6+ support (async/await, fetch API)
- Uses localStorage for token persistence

## Security Notes
- Token stored in browser localStorage (not secure for production)
- No token encryption in browser
- CORS must be enabled on backend
- For production: Use secure token storage (httpOnly cookies, etc.)

## Skipped Features (Intentionally Minimal)
- Member CRUD operations (admin feature)
- Email sending (backend feature)
- File upload (nice-to-have)
- Entry update/delete (requires modify:entry scope)
- Advanced filtering/search
- Likes management

## Future Enhancements
- Click entry to view full details
- Rich text editor for messages
- Image upload support
- Real-time updates (WebSocket)
- User profile display
- Dark mode toggle
