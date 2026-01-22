package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	googleDeviceURL = "https://oauth2.googleapis.com/device/code"
	googleTokenURL  = "https://oauth2.googleapis.com/token"
	defaultAPI      = "https://api.chalmerslosers.com"
)

type Config struct {
	AccessToken string `json:"access_token"`
	MemberNum   int64  `json:"member_number,omitempty"`
	Email       string `json:"email,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "token":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "add":
			tokenAdd()
		case "show":
			tokenShow()
		default:
			printUsage()
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: sidan-auth <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  token add    Authenticate with Google and get API token")
	fmt.Println("  token show   Show current token")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SIDAN_API_URL       API URL (default: https://api.chalmerslosers.com)")
	fmt.Println("  GOOGLE_CLIENT_ID    Google OAuth2 client ID (required for token add)")
}

func tokenAdd() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		fmt.Fprintln(os.Stderr, "Error: GOOGLE_CLIENT_ID environment variable required")
		fmt.Fprintln(os.Stderr, "Get it from Google Cloud Console > APIs & Services > Credentials")
		os.Exit(1)
	}

	apiURL := os.Getenv("SIDAN_API_URL")
	if apiURL == "" {
		apiURL = defaultAPI
	}

	// Step 1: Request device code from Google
	deviceResp, err := requestDeviceCode(clientID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error requesting device code: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Display verification URL and code
	fmt.Println()
	fmt.Println("To authenticate, visit:")
	fmt.Printf("  %s\n", deviceResp.VerificationURL)
	fmt.Println()
	fmt.Printf("And enter code: %s\n", deviceResp.UserCode)
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Step 3: Poll for Google token
	googleToken, err := pollForToken(clientID, deviceResp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Google authorization successful!")

	// Step 4: Exchange Google token for our JWT
	jwt, member, err := exchangeForJWT(apiURL, googleToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exchanging token: %v\n", err)
		os.Exit(1)
	}

	// Step 5: Save to config
	cfg := Config{
		AccessToken: jwt,
		MemberNum:   member.Number,
		Email:       member.Email,
		ExpiresAt:   time.Now().Add(8 * time.Hour).Format(time.RFC3339),
	}
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Authenticated as member #%d (%s)\n", member.Number, member.Email)
	fmt.Printf("Token saved to %s\n", configPath())
}

func tokenShow() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No token found. Run 'sidan-auth token add' first.")
		os.Exit(1)
	}

	fmt.Printf("Member: #%d\n", cfg.MemberNum)
	fmt.Printf("Email: %s\n", cfg.Email)
	fmt.Printf("Expires: %s\n", cfg.ExpiresAt)
	fmt.Println()
	fmt.Println("Token:")
	fmt.Println(cfg.AccessToken)
}

// Google device flow types and functions

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func requestDeviceCode(clientID string) (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {clientID},
		"scope":     {"email"},
	}

	resp, err := http.PostForm(googleDeviceURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google returned %d: %s", resp.StatusCode, body)
	}

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func pollForToken(clientID string, device *DeviceCodeResponse) (string, error) {
	interval := time.Duration(device.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(interval)

		data := url.Values{
			"client_id":   {clientID},
			"device_code": {device.DeviceCode},
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		}

		resp, err := http.PostForm(googleTokenURL, data)
		if err != nil {
			return "", err
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		json.Unmarshal(body, &result)

		switch result.Error {
		case "":
			return result.AccessToken, nil
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "access_denied":
			return "", fmt.Errorf("access denied by user")
		default:
			return "", fmt.Errorf("google error: %s", result.Error)
		}
	}

	return "", fmt.Errorf("authorization timed out")
}

// Backend exchange

type MemberInfo struct {
	Number int64  `json:"number"`
	Email  string `json:"email"`
}

func exchangeForJWT(apiURL, googleToken string) (string, *MemberInfo, error) {
	body, _ := json.Marshal(map[string]string{"access_token": googleToken})

	resp, err := http.Post(apiURL+"/auth/device/verify", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		var errResp struct{ Error string `json:"error"` }
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return "", nil, fmt.Errorf(errResp.Error)
		}
		return "", nil, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string     `json:"access_token"`
		Member      MemberInfo `json:"member"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, err
	}

	return result.AccessToken, &result.Member, nil
}

// Config file management

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sidan", "config.json")
}

func saveConfig(cfg Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
