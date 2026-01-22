package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configDir  = ".sidan"
	configFile = "config.json"
)

// Config stores CLI configuration including tokens
type Config struct {
	APIEndpoint string            `json:"api_endpoint"`
	Tokens      map[string]string `json:"tokens"` // provider -> token
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// loadConfig loads configuration from disk
func loadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Create default config if doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := &Config{
			APIEndpoint: "http://localhost:8080",
			Tokens:      make(map[string]string),
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// saveConfig saves configuration to disk
func saveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// printUsage prints CLI usage
func printUsage() {
	fmt.Println("sidan-auth - Sidan API authentication CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  sidan-auth token add [provider]    Add new authentication token")
	fmt.Println("  sidan-auth token list               List saved tokens")
	fmt.Println("  sidan-auth token remove [provider]  Remove a token")
	fmt.Println("  sidan-auth config show              Show configuration")
	fmt.Println("  sidan-auth config set-api [url]     Set API endpoint")
	fmt.Println()
	fmt.Println("Providers: google, github")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "token":
		if len(os.Args) < 3 {
			fmt.Println("Usage: sidan-auth token [add|list|remove]")
			os.Exit(1)
		}
		handleTokenCommand(os.Args[2], os.Args[3:])
	case "config":
		if len(os.Args) < 3 {
			fmt.Println("Usage: sidan-auth config [show|set-api]")
			os.Exit(1)
		}
		handleConfigCommand(os.Args[2], os.Args[3:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleTokenCommand(subcommand string, args []string) {
	switch subcommand {
	case "add":
		provider := "google"
		if len(args) > 0 {
			provider = args[0]
		}
		if err := addToken(provider); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := listTokens(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if len(args) < 1 {
			fmt.Println("Usage: sidan-auth token remove [provider]")
			os.Exit(1)
		}
		if err := removeToken(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown token command: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleConfigCommand(subcommand string, args []string) {
	switch subcommand {
	case "show":
		if err := showConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "set-api":
		if len(args) < 1 {
			fmt.Println("Usage: sidan-auth config set-api [url]")
			os.Exit(1)
		}
		if err := setAPIEndpoint(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown config command: %s\n", subcommand)
		os.Exit(1)
	}
}

func addToken(provider string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Requesting device authorization for %s...\n", provider)

	// Request device code
	deviceResp, err := requestDeviceCode(cfg.APIEndpoint, provider)
	if err != nil {
		return fmt.Errorf("failed to request device code: %w", err)
	}

	// Display verification URL
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("To authorize this device, visit:")
	fmt.Println()
	fmt.Printf("    %s\n", deviceResp.VerificationURI)
	fmt.Println()
	fmt.Println("And enter this code:")
	fmt.Println()
	fmt.Printf("    %s\n", deviceResp.UserCode)
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Poll for token
	token, err := pollForToken(cfg.APIEndpoint, deviceResp.DeviceCode, deviceResp.Interval, deviceResp.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Save token
	cfg.Tokens[provider] = token
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Authorization successful!")
	fmt.Printf("Token saved for provider: %s\n", provider)
	fmt.Println()
	fmt.Println("You can now use the API with:")
	fmt.Printf("  export SIDAN_TOKEN=\"%s\"\n", token[:20]+"...")
	return nil
}

func listTokens() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Tokens) == 0 {
		fmt.Println("No tokens saved.")
		fmt.Println("Run 'sidan-auth token add' to add a token.")
		return nil
	}

	fmt.Println("Saved tokens:")
	for provider := range cfg.Tokens {
		fmt.Printf("  - %s\n", provider)
	}
	return nil
}

func removeToken(provider string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.Tokens[provider]; !exists {
		return fmt.Errorf("no token found for provider: %s", provider)
	}

	delete(cfg.Tokens, provider)
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Token removed for provider: %s\n", provider)
	return nil
}

func showConfig() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	path, _ := getConfigPath()
	fmt.Printf("Config file: %s\n", path)
	fmt.Printf("API endpoint: %s\n", cfg.APIEndpoint)
	fmt.Printf("Saved tokens: %d\n", len(cfg.Tokens))
	return nil
}

func setAPIEndpoint(url string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.APIEndpoint = url
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("API endpoint set to: %s\n", url)
	return nil
}

// Device flow API types
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func requestDeviceCode(apiEndpoint, provider string) (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/auth/device?provider=%s", apiEndpoint, provider)
	
	resp, err := httpPost(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("server error: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s: %s", errResp.Error, errResp.ErrorDescription)
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, err
	}

	return &deviceResp, nil
}

func pollForToken(apiEndpoint, deviceCode string, interval, expiresIn int) (string, error) {
	url := fmt.Sprintf("%s/auth/device/token", apiEndpoint)
	
	timeout := time.Now().Add(time.Duration(expiresIn) * time.Second)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		if time.Now().After(timeout) {
			return "", fmt.Errorf("authorization timeout")
		}

		body := map[string]string{
			"device_code": deviceCode,
			"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		}

		resp, err := httpPost(url, body)
		if err != nil {
			return "", err
		}

		if resp.StatusCode == 200 {
			var tokenResp TokenResponse
			if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
				resp.Body.Close()
				return "", err
			}
			resp.Body.Close()
			return tokenResp.AccessToken, nil
		}

		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		resp.Body.Close()

		if errResp.Error == "authorization_pending" {
			// User hasn't approved yet, continue polling
			<-ticker.C
			continue
		} else if errResp.Error == "expired_token" {
			return "", fmt.Errorf("device code expired")
		} else {
			return "", fmt.Errorf("error: %s", errResp.Error)
		}
	}
}
