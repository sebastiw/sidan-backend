package router

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/models"
)

// GenerateUserCode creates a short, readable code (e.g., ABCD-1234)
func generateUserCode() (string, error) {
	b := make([]byte, 5) // 5 bytes = 40 bits. Base32 is 5 bits/char -> 8 chars
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	str := base32.StdEncoding.EncodeToString(b)
	return fmt.Sprintf("%s-%s", str[:4], str[4:8]), nil
}

// DeviceInit initiates the device authorization flow
// POST /auth/device/code
func (h *AuthHandler) DeviceInit(w http.ResponseWriter, r *http.Request) {
	// Generate codes
	deviceCode := uuid.New().String()
	userCode, err := generateUserCode()
	if err != nil {
		slog.Error("failed to generate user code", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Store in DB
	state := &models.DeviceAuthState{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		Status:     "pending",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
		LastChecked: time.Now(),
	}

	if err := h.db.CreateDeviceAuthState(state); err != nil {
		slog.Error("failed to create device auth state", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Return response (RFC 8628 style)
	host := config.GetServer().Host // e.g., "api.chalmerslosers.com" - might need to check how to get public URL
	// If config Host doesn't include protocol/port, we might need to construct it.
	// Assuming config has enough or we infer from request?
	// The example output uses "https://api.chalmerslosers.com/auth/device"
	// I'll assume I can construct it relative or use config.
	
	// For now, construct verification URI
	verificationURI := fmt.Sprintf("https://%s/auth/device", r.Host) // Using request host is often safer if behind proxy

	resp := map[string]interface{}{
		"device_code":      deviceCode,
		"user_code":        userCode,
		"verification_uri": verificationURI,
		"verification_uri_complete": fmt.Sprintf("%s?code=%s", verificationURI, userCode),
		"expires_in":       900, // 15 minutes
		"interval":         5,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeviceToken polls for the token
// POST /auth/device/token
func (h *AuthHandler) DeviceToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceCode string `json:"device_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	state, err := h.db.GetDeviceAuthStateByDeviceCode(req.DeviceCode)
	if err != nil {
		http.Error(w, "invalid_grant", http.StatusBadRequest) // or expired
		return
	}

	switch state.Status {
	case "pending":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // RFC says 400 for authorization_pending
		json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
		return
	case "denied":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "access_denied"})
		return
	case "approved":
		// Issue Token
		if state.MemberNumber == nil {
			slog.Error("approved state missing member number")
			http.Error(w, "server_error", http.StatusInternalServerError)
			return
		}

		member, err := h.db.ReadMemberByNumber(*state.MemberNumber)
		if err != nil {
			http.Error(w, "server_error", http.StatusInternalServerError)
			return
		}

		// Use stored scopes or default
		scopes := getScopesForMemberType(member)
		
		// Generate JWT
		// We use "device" as provider or maybe "sidan-auth"
		jwtToken, err := auth.GenerateJWT(member.Number, *member.Email, scopes, "sidan-cli", config.GetJWTSecret())
		if err != nil {
			http.Error(w, "server_error", http.StatusInternalServerError)
			return
		}

		// Delete state (single use)
		// Or maybe keep it to prevent replay? But device code flow usually allows refresh tokens.
		// For MVP, we just return the Access Token.
		// Ideally we should delete or mark as used.
		// Let's mark as used/completed to avoid reuse if we want strictness.
		// h.db.DeleteDeviceAuthState... (Not implemented yet, but Update works)
		state.Status = "completed"
		h.db.UpdateDeviceAuthState(state)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": jwtToken,
			"token_type":   "Bearer",
			"expires_in":   28800,
		})
		return
	default:
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}
}

// DeviceVerifyPage renders the verification UI
// GET /auth/device
func (h *AuthHandler) DeviceVerifyPage(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	
	// Simple HTML UI
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Device Activation</title>
    <style>
        body { font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #f0f0f0; }
        .card { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); max-width: 400px; width: 100%%; text-align: center; }
        input { font-size: 1.5rem; letter-spacing: 0.2rem; text-align: center; width: 100%%; padding: 0.5rem; margin: 1rem 0; text-transform: uppercase; }
        button { background: #007bff; color: white; border: none; padding: 0.75rem 1.5rem; font-size: 1rem; border-radius: 4px; cursor: pointer; width: 100%%; }
        button:disabled { background: #ccc; }
        .hidden { display: none; }
    </style>
</head>
<body>
    <div class="card" id="app">
        <h2>Connect Device</h2>
        <p>Enter the code displayed on your device</p>
        <input type="text" id="code" value="%s" placeholder="ABCD-1234" maxlength="9">
        <div id="status"></div>
        <button id="btn" onclick="verify()">Connect</button>
    </div>
    <script>
        const codeInput = document.getElementById('code');
        const btn = document.getElementById('btn');
        const status = document.getElementById('status');

        // Check if logged in (check for token in URL hash if just redirected, or localStorage)
        // Since we don't have cookies, we need to handle auth flow here.
        
        let token = localStorage.getItem('sidan_token');
        
        // Handle redirect from login
        const params = new URLSearchParams(window.location.search);
        if (params.get('token')) {
            const tokenData = JSON.parse(params.get('token'));
            token = tokenData.access_token;
            localStorage.setItem('sidan_token', token);
            // Clean URL
            window.history.replaceState({}, document.title, window.location.pathname + (codeInput.value ? '?code='+codeInput.value : ''));
        }

        async function checkLogin() {
            if (!token) {
                btn.innerText = "Log in to Continue";
                btn.onclick = () => {
                     // Redirect to login, with callback to here
                     const currentUrl = window.location.href;
                     // We need to pass redirect_uri to /auth/login
                     // The login handler expects redirect_uri
                     window.location.href = '/auth/login?provider=google&redirect_uri=' + encodeURIComponent(window.location.protocol + '//' + window.location.host + window.location.pathname + '?code=' + codeInput.value);
                };
                return false;
            }
            return true;
        }

        async function verify() {
            if (!await checkLogin()) return;

            const code = codeInput.value;
            if (!code) return;

            btn.disabled = true;
            btn.innerText = "Verifying...";

            try {
                const res = await fetch('/auth/device/verify', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + token
                    },
                    body: JSON.stringify({ user_code: code })
                });

                if (res.ok) {
                    document.getElementById('app').innerHTML = '<h2>Success!</h2><p>You can now return to your device.</p>';
                } else {
                    const data = await res.json();
                    status.innerText = data.error || 'Error verifying code';
                    status.style.color = 'red';
                    btn.disabled = false;
                    btn.innerText = "Connect";
                }
            } catch (e) {
                status.innerText = 'Network error';
                btn.disabled = false;
                btn.innerText = "Connect";
            }
        }

        checkLogin();
    </script>
</body>
</html>
	`, code)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// DeviceVerifyAction handles the approval
// POST /auth/device/verify
// Requires Auth
func (h *AuthHandler) DeviceVerifyAction(w http.ResponseWriter, r *http.Request) {
    // Get Member from context (set by middleware)
    member := auth.GetMember(r)
    if member == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    var req struct {
        UserCode string `json:"user_code"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    // Normalize code (uppercase)
    req.UserCode = strings.ToUpper(req.UserCode)

    state, err := h.db.GetDeviceAuthStateByUserCode(req.UserCode)
    if err != nil {
        http.Error(w, "invalid code", http.StatusBadRequest)
        return
    }
    
    if state.Status != "pending" {
        http.Error(w, "code already used or expired", http.StatusBadRequest)
        return
    }

    // Approve
    state.Status = "approved"
    state.MemberNumber = &member.Number
    // state.Scopes = ... (could store specific scopes requested, for now defaults)
    
    if err := h.db.UpdateDeviceAuthState(state); err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
