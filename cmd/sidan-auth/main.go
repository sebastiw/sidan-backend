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
	githubDeviceURL = "https://github.com/login/device/code"
	githubTokenURL  = "https://github.com/login/oauth/access_token"
	defaultAPI      = "https://api.chalmerslosers.com"
)

type providerCfg struct {
	clientIDEnv string
	deviceURL   string
	tokenURL    string
	scope       string
}

var providers = map[string]providerCfg{
	"google": {
		clientIDEnv: "GOOGLE_CLIENT_ID",
		deviceURL:   googleDeviceURL,
		tokenURL:    googleTokenURL,
		scope:       "email",
	},
	"github": {
		clientIDEnv: "GITHUB_CLIENT_ID",
		deviceURL:   githubDeviceURL,
		tokenURL:    githubTokenURL,
		scope:       "user:email",
	},
}

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
			provider := "google"
			if len(os.Args) > 3 {
				provider = os.Args[3]
			}
			tokenAdd(provider)
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
	fmt.Println("  token add [provider]   Authenticate and get API token (provider: google, github; default: google)")
	fmt.Println("  token show             Show current token")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SIDAN_API_URL       API URL (default: https://api.chalmerslosers.com)")
	fmt.Println("  GOOGLE_CLIENT_ID    Google OAuth2 client ID (for token add google)")
	fmt.Println("  GITHUB_CLIENT_ID    GitHub OAuth2 client ID (for token add github)")
}

func tokenAdd(provider string) {
	pcfg, ok := providers[provider]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unsupported provider '%s'. Use 'google' or 'github'\n", provider)
		os.Exit(1)
	}

	clientID := os.Getenv(pcfg.clientIDEnv)
	if clientID == "" {
		fmt.Fprintf(os.Stderr, "Error: %s environment variable required\n", pcfg.clientIDEnv)
		os.Exit(1)
	}

	apiURL := os.Getenv("SIDAN_API_URL")
	if apiURL == "" {
		apiURL = defaultAPI
	}

	// Step 1: Request device code
	deviceResp, err := requestDeviceCode(clientID, pcfg.deviceURL, pcfg.scope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error requesting device code: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Display verification URL and code
	verifyURL := deviceResp.VerificationURL
	if verifyURL == "" {
		verifyURL = deviceResp.VerificationURI
	}
	fmt.Println()
	fmt.Println("To authenticate, visit:")
	fmt.Printf("  %s\n", verifyURL)
	fmt.Println()
	fmt.Printf("And enter code: %s\n", deviceResp.UserCode)
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Step 3: Poll for provider token
	providerToken, err := pollForToken(clientID, pcfg.tokenURL, deviceResp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s authorization successful!\n", provider)

	// Step 4: Exchange provider token for our JWT
	jwt, member, err := exchangeForJWT(apiURL, provider, providerToken)
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
	VerificationURL string `json:"verification_url"` // Google
	VerificationURI string `json:"verification_uri"` // GitHub
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func requestDeviceCode(clientID, deviceURL, scope string) (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {clientID},
		"scope":     {scope},
	}

	resp, err := http.PostForm(deviceURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provider returned %d: %s", resp.StatusCode, body)
	}

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func pollForToken(clientID, tokenURL string, device *DeviceCodeResponse) (string, error) {
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

		resp, err := http.PostForm(tokenURL, data)
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

func exchangeForJWT(apiURL, provider, accessToken string) (string, *MemberInfo, error) {
	body, _ := json.Marshal(map[string]string{"access_token": accessToken, "provider": provider})

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
