package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
)

type DeviceFlowHandler struct {
	db data.Database
}

func NewDeviceFlowHandler(db data.Database) *DeviceFlowHandler {
	return &DeviceFlowHandler{db: db}
}

// DeviceAuthRequest handles device authorization requests
// POST /auth/device?provider=google
func (h *DeviceFlowHandler) DeviceAuthRequest(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, `{"error":"invalid_request","error_description":"provider required"}`, http.StatusBadRequest)
		return
	}

	// Validate provider exists in config
	if _, exists := config.Get().OAuth2[provider]; !exists {
		http.Error(w, `{"error":"invalid_request","error_description":"unknown provider"}`, http.StatusBadRequest)
		return
	}

	// Generate device code and user code
	deviceCode, err := auth.GenerateDeviceCode()
	if err != nil {
		slog.Error("failed to generate device code", "error", err)
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	userCode, err := auth.GenerateUserCode()
	if err != nil {
		slog.Error("failed to generate user code", "error", err)
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	// Build verification URI
	serverURL := fmt.Sprintf("http://%s", r.Host)
	if r.TLS != nil {
		serverURL = fmt.Sprintf("https://%s", r.Host)
	}
	verificationURI := fmt.Sprintf("%s/auth/device/verify?code=%s&provider=%s", serverURL, userCode, provider)

	// Create device code entry
	deviceCodeEntry := &models.DeviceCode{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: verificationURI,
		ExpiresAt:       time.Now().Add(10 * time.Minute),
		Interval:        5, // 5 seconds polling interval
		Approved:        false,
		Provider:        provider,
		CreatedAt:       time.Now(),
	}

	if err := h.db.CreateDeviceCode(deviceCodeEntry); err != nil {
		slog.Error("failed to store device code", "error", err)
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	slog.Info("device code created", "user_code", userCode, "provider", provider)

	// Return device authorization response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_code":      deviceCode,
		"user_code":        userCode,
		"verification_uri": verificationURI,
		"expires_in":       600, // 10 minutes
		"interval":         5,   // Poll every 5 seconds
	})
}

// DeviceVerifyPage shows the verification page
// GET /auth/device/verify?code=XXXX-XXXX&provider=google
func (h *DeviceFlowHandler) DeviceVerifyPage(w http.ResponseWriter, r *http.Request) {
	userCode := r.URL.Query().Get("code")
	provider := r.URL.Query().Get("provider")

	if userCode == "" || provider == "" {
		http.Error(w, "Invalid request: code and provider required", http.StatusBadRequest)
		return
	}

	// Validate user code format
	if !auth.ValidateUserCode(userCode) {
		http.Error(w, "Invalid code format", http.StatusBadRequest)
		return
	}

	// Check if code exists
	deviceCode, err := h.db.GetDeviceCodeByUserCode(userCode)
	if err != nil {
		http.Error(w, "Invalid or expired code", http.StatusBadRequest)
		return
	}

	if deviceCode.Approved {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Device Authorization</title></head>
<body>
<h1>Already Approved</h1>
<p>This code has already been approved. You can close this window.</p>
</body>
</html>`)
		return
	}

	// Show verification page with OAuth2 login option
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<title>Device Authorization</title>
<style>
body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
.code { font-size: 32px; font-weight: bold; text-align: center; margin: 20px 0; letter-spacing: 5px; }
.btn { display: block; width: 100%%; padding: 15px; margin: 10px 0; font-size: 18px; 
       background: #4285f4; color: white; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; text-align: center; }
.btn:hover { background: #357ae8; }
</style>
</head>
<body>
<h1>Authorize Device</h1>
<p>Please confirm that you see this code on your device:</p>
<div class="code">%s</div>
<p>To authorize this device, sign in with %s:</p>
<a href="/auth/login?provider=%s&redirect_uri=%s" class="btn">Sign in with %s</a>
<p><small>Code expires in 10 minutes</small></p>
</body>
</html>`, userCode, provider, provider, 
		fmt.Sprintf("http://%s/auth/device/callback?code=%s", r.Host, userCode),
		provider)
}

// DeviceCallback handles OAuth2 callback for device verification
// GET /auth/device/callback?code=XXXX-XXXX&token=<url_encoded_json>
func (h *DeviceFlowHandler) DeviceCallback(w http.ResponseWriter, r *http.Request) {
	userCode := r.URL.Query().Get("code")
	tokenParam := r.URL.Query().Get("token")

	if userCode == "" || tokenParam == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// Parse token JSON
	var tokenData struct {
		AccessToken string `json:"access_token"`
		Member      struct {
			Number int64  `json:"number"`
			Email  string `json:"email"`
		} `json:"member"`
		Scopes []string `json:"scopes"`
	}

	if err := json.Unmarshal([]byte(tokenParam), &tokenData); err != nil {
		slog.Error("failed to parse token data", "error", err)
		http.Error(w, "Invalid token data", http.StatusBadRequest)
		return
	}

	// Get device code entry
	deviceCode, err := h.db.GetDeviceCodeByUserCode(userCode)
	if err != nil {
		http.Error(w, "Invalid or expired code", http.StatusBadRequest)
		return
	}

	// Update device code with approval
	scopesJSON, _ := json.Marshal(tokenData.Scopes)
	scopesStr := string(scopesJSON)
	deviceCode.Approved = true
	deviceCode.MemberNumber = &tokenData.Member.Number
	deviceCode.Email = &tokenData.Member.Email
	deviceCode.Scopes = &scopesStr

	if err := h.db.UpdateDeviceCode(deviceCode); err != nil {
		slog.Error("failed to update device code", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	slog.Info("device code approved", "user_code", userCode, "member", tokenData.Member.Number)

	// Show success page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Device Authorized</title></head>
<body>
<h1>Success!</h1>
<p>Your device has been authorized. You can close this window and return to your device.</p>
</body>
</html>`)
}

// DeviceToken handles token polling from device
// POST /auth/device/token
// Body: {"device_code": "...", "grant_type": "urn:ietf:params:oauth:grant-type:device_code"}
func (h *DeviceFlowHandler) DeviceToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceCode string `json:"device_code"`
		GrantType  string `json:"grant_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	if req.GrantType != "urn:ietf:params:oauth:grant-type:device_code" {
		http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
		return
	}

	// Get device code entry
	deviceCode, err := h.db.GetDeviceCodeByDeviceCode(req.DeviceCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "expired_token",
			"error_description": "The device code has expired",
		})
		return
	}

	// Check if approved
	if !deviceCode.Approved {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "authorization_pending",
			"error_description": "User has not yet authorized the device",
		})
		return
	}

	// Parse scopes
	var scopes []string
	if deviceCode.Scopes != nil {
		json.Unmarshal([]byte(*deviceCode.Scopes), &scopes)
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateJWT(*deviceCode.MemberNumber, *deviceCode.Email, scopes, deviceCode.Provider, config.GetJWTSecret())
	if err != nil {
		slog.Error("JWT generation failed", "error", err)
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	// Delete device code (one-time use)
	h.db.DeleteDeviceCode(req.DeviceCode)

	slog.Info("device token issued", "member", *deviceCode.MemberNumber, "provider", deviceCode.Provider)

	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": jwtToken,
		"token_type":   "Bearer",
		"expires_in":   28800, // 8 hours
	})
}
