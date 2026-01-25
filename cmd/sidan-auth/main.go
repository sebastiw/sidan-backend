package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultAPI = "https://api.chalmerslosers.com"

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
}

func main() {
	apiURL := flag.String("api", defaultAPI, "API URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 || args[0] != "token" || args[1] != "add" {
		fmt.Println("Usage: sidan-auth [options] token add")
		os.Exit(1)
	}

	fmt.Printf("[add token using %s/auth]\n", *apiURL)

	// 1. Request Device Code
	resp, err := http.Post(*apiURL+"/auth/device/code", "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error contacting API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "API Error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Activation URL:\n%s?code=%s\n", deviceResp.VerificationURI, deviceResp.UserCode)
	// Or use VerificationURIComplete if we want to be nicer, but prompts user code separate
	// The prompt example showed: https://.../auth/device?code=...
	
	interval := time.Duration(deviceResp.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	// 2. Poll for Token
	fmt.Println("\nWaiting for approval...")
	
	client := &http.Client{}
	
	for {
		time.Sleep(interval)

	
tokenReq := map[string]string{
			"device_code": deviceResp.DeviceCode,
		}
		jsonBody, _ := json.Marshal(tokenReq)

		req, _ := http.NewRequest("POST", *apiURL+"/auth/device/token", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

	
tokenResp, err := client.Do(req)
		if err != nil {
			fmt.Printf(".")
			continue
		}
		defer tokenResp.Body.Close()

		if tokenResp.StatusCode == 200 {
			var tokenData TokenResponse
			if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
				fmt.Fprintf(os.Stderr, "\nError decoding token: %v\n", err)
				os.Exit(1)
			}
			
			saveToken(tokenData)
			fmt.Println("\nSuccess! Token saved.")
			return
		}

		var errData TokenResponse
		json.NewDecoder(tokenResp.Body).Decode(&errData)
		
		if errData.Error == "authorization_pending" {
			// Continue polling
		} else if errData.Error == "slow_down" {
			interval += 5 * time.Second
		} else if errData.Error == "access_denied" {
			fmt.Println("\nAccess denied by user.")
			os.Exit(1)
		} else if errData.Error == "expired_token" {
			fmt.Println("\nSession expired. Please try again.")
			os.Exit(1)
		} else {
			fmt.Printf("\nError: %s\n", errData.Error)
			os.Exit(1)
		}
	}
}

func saveToken(token TokenResponse) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find home directory: %v\n", err)
		return
	}

	configDir := filepath.Join(home, ".sidan")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create config directory: %v\n", err)
		return
	}

	configFile := filepath.Join(configDir, "config.yaml")
	
	// Simple YAML/JSON write
	// We preserve existing config if possible? 
	// For MVP, just append or overwrite specific key is hard without a parser.
	// We'll write a simple file. 
	
	file, err := os.Create(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create config file: %v\n", err)
		return
	}
	defer file.Close()
	
	// Writing simple YAML
	fmt.Fprintf(file, "access_token: %s\n", token.AccessToken)
	fmt.Fprintf(file, "token_type: %s\n", token.TokenType)
	fmt.Fprintf(file, "expires_in: %d\n", token.ExpiresIn)
	
	fmt.Printf("Token written to %s\n", configFile)
}
