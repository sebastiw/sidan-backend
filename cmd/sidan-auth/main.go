package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/atotto/clipboard"
)

const defaultAPI = "https://api.chalmerslosers.com"

type Config struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Provider     string `json:"provider,omitempty"`
	MemberNum    int64  `json:"member_number,omitempty"`
	Email        string `json:"email,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
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
			if len(os.Args) < 4 {
				printUsage()
				os.Exit(1)
			}
			tokenAdd(os.Args[3])
		case "refresh":
			tokenRefresh()
		case "show":
			tokenShow()
		case "raw":
			tokenRaw()
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
	fmt.Println("  token add <provider>   Authenticate and get API token (provider: google, github)")
	fmt.Println("  token refresh          Silently refresh expired token using stored refresh token")
	fmt.Println("  token show             Show current token")
	fmt.Println("  token raw              Print raw token value only")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SIDAN_API_URL   API URL (default: https://api.chalmerslosers.com)")
}

func apiURL() string {
	if u := os.Getenv("SIDAN_API_URL"); u != "" {
		return u
	}
	return defaultAPI
}

func tokenAdd(provider string) {
	// Step 1: Start device flow on server
	startResp, err := startDeviceFlow(provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting device flow: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Show code and open browser
	fmt.Println()
	fmt.Printf("Code: %s\n", startResp.UserCode)
	browserURL := startResp.BrowserURL
	if browserURL == "" {
		browserURL = startResp.VerificationURL
	}
	if openBrowser(browserURL) {
		fmt.Println("Browser opened — approve the request to continue.")
	} else {
		fmt.Printf("Visit: %s\n", browserURL)
	}
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Step 3: Poll server until approved or timed out
	interval := time.Duration(startResp.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(startResp.ExpiresIn) * time.Second)

	var pollResult *pollFlowResponse
	for time.Now().Before(deadline) {
		time.Sleep(interval)

		result, done, err := pollDeviceFlow(startResp.SessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if done {
			pollResult = result
			break
		}
		if result.Status == "slow_down" {
			interval += 5 * time.Second
		}
	}

	if pollResult == nil {
		fmt.Fprintln(os.Stderr, "Error: authorization timed out")
		os.Exit(1)
	}

	fmt.Printf("%s authorization successful!\n", provider)

	cfg := Config{
		AccessToken:  pollResult.AccessToken,
		RefreshToken: pollResult.RefreshToken,
		Provider:     provider,
		MemberNum:    pollResult.Member.Number,
		Email:        pollResult.Member.Email,
		ExpiresAt:    time.Now().Add(8 * time.Hour).Format(time.RFC3339),
	}
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Authenticated as member #%d (%s)\n", pollResult.Member.Number, pollResult.Member.Email)
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

func tokenRaw() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No token found. Run 'sidan-auth token add' first.")
		os.Exit(1)
	}
	fmt.Print(cfg.AccessToken)
}

func tokenRefresh() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No token found. Run 'sidan-auth token add' first.")
		os.Exit(1)
	}
	if cfg.RefreshToken == "" {
		fmt.Fprintln(os.Stderr, "No refresh token stored. Run 'sidan-auth token add' to re-authenticate.")
		os.Exit(1)
	}
	if cfg.Provider == "" {
		fmt.Fprintln(os.Stderr, "No provider stored. Run 'sidan-auth token add' to re-authenticate.")
		os.Exit(1)
	}

	result, err := refreshJWT(cfg.Provider, cfg.RefreshToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error refreshing token: %v\n", err)
		os.Exit(1)
	}

	cfg.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		cfg.RefreshToken = result.RefreshToken
	}
	cfg.MemberNum = result.Member.Number
	cfg.Email = result.Member.Email
	cfg.ExpiresAt = time.Now().Add(8 * time.Hour).Format(time.RFC3339)
	if err := saveConfig(*cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Token refreshed for member #%d (%s)\n", result.Member.Number, result.Member.Email)
	fmt.Printf("Expires: %s\n", cfg.ExpiresAt)
}

// API calls

type startFlowResponse struct {
	SessionID       string `json:"session_id"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	BrowserURL      string `json:"browser_url"`
	Interval        int    `json:"interval"`
	ExpiresIn       int    `json:"expires_in"`
}

type pollFlowResponse struct {
	Status       string     `json:"status"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	Member       MemberInfo `json:"member"`
}

type MemberInfo struct {
	Number int64  `json:"number"`
	Email  string `json:"email"`
}

func startDeviceFlow(provider string) (*startFlowResponse, error) {
	body, _ := json.Marshal(map[string]string{"provider": provider})
	resp, err := http.Post(apiURL()+"/auth/device/start", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp struct{ Error string `json:"error"` }
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	var result startFlowResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func pollDeviceFlow(sessionID string) (*pollFlowResponse, bool, error) {
	body, _ := json.Marshal(map[string]string{"session_id": sessionID})
	resp, err := http.Post(apiURL()+"/auth/device/poll", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusAccepted {
		var result pollFlowResponse
		json.Unmarshal(respBody, &result)
		return &result, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct{ Error string `json:"error"` }
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return nil, false, fmt.Errorf("%s", errResp.Error)
		}
		return nil, false, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	var result pollFlowResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, err
	}
	return &result, true, nil
}

type refreshJWTResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	Member       MemberInfo `json:"member"`
}

func refreshJWT(provider, refreshToken string) (*refreshJWTResponse, error) {
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken, "provider": provider})
	resp, err := http.Post(apiURL()+"/auth/device/refresh", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp struct{ Error string `json:"error"` }
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	var result refreshJWTResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Utilities

func openBrowser(url string) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return false
	}
	return cmd.Start() == nil
}

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
