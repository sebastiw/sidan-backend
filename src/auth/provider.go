package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	ErrAuthPending = errors.New("authorization_pending")
	ErrSlowDown    = errors.New("slow_down")
)

// ProviderConfig holds OAuth2 provider configuration
type ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Endpoint     oauth2.Endpoint
	UserInfoURL  string
}


// UserInfo represents user information from OAuth2 provider
type UserInfo struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Name           string
	Picture        string
}

// GetProviderConfig returns OAuth2 configuration for a provider
func GetProviderConfig(provider string, clientID, clientSecret, redirectURL string, scopes []string) (*ProviderConfig, error) {
	switch provider {
	case "google":
		return &ProviderConfig{
			Name:         "google",
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		}, nil
	case "github":
		return &ProviderConfig{
			Name:         "github",
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     github.Endpoint,
			UserInfoURL:  "https://api.github.com/user",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// GetAuthURL builds the OAuth2 authorization URL
func (p *ProviderConfig) GetAuthURL(state, codeChallenge string) string {
	params := url.Values{
		"client_id":             {p.ClientID},
		"redirect_uri":          {p.RedirectURL},
		"response_type":         {"code"},
		"scope":                 {strings.Join(p.Scopes, " ")},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	// Google-specific: add access_type for refresh token
	if p.Name == "google" {
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	}

	return p.Endpoint.AuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges authorization code for access token
func (p *ProviderConfig) ExchangeCode(code, codeVerifier string) (*oauth2.Token, error) {
	config := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		RedirectURL:  p.RedirectURL,
		Scopes:       p.Scopes,
		Endpoint:     p.Endpoint,
	}

	ctx := context.Background()
	token, err := config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	return token, nil
}

// StartDeviceFlow initiates the device authorization flow with the provider
func (p *ProviderConfig) StartDeviceFlow(ctx context.Context) (*oauth2.DeviceAuthResponse, error) {
	cfg := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       p.Scopes,
		Endpoint:     p.Endpoint,
	}
	return cfg.DeviceAuth(ctx)
}

// PollDeviceToken polls the provider token endpoint to check if the user has authorized the device
func (p *ProviderConfig) PollDeviceToken(deviceCode string) (accessToken, refreshToken string, err error) {
	data := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"device_code":   {deviceCode},
		"grant_type":    {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	resp, err := http.PostForm(p.Endpoint.TokenURL, data)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	switch result.Error {
	case "":
		return result.AccessToken, result.RefreshToken, nil
	case "authorization_pending":
		return "", "", ErrAuthPending
	case "slow_down":
		return "", "", ErrSlowDown
	default:
		return "", "", fmt.Errorf("provider error: %s", result.Error)
	}
}

// RefreshAccessToken exchanges a refresh token for a new provider access token
func (p *ProviderConfig) RefreshAccessToken(refreshToken string) (string, error) {
	data := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	resp, err := http.PostForm(p.Endpoint.TokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("refresh failed: %s", result.ErrorDesc)
	}
	return result.AccessToken, nil
}

// GetUserInfo fetches user information from provider using access token
func (p *ProviderConfig) GetUserInfo(accessToken string) (*UserInfo, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	client := oauth2.NewClient(ctx, ts)

	req, err := http.NewRequest("GET", p.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	// GitHub requires Accept header
	if p.Name == "github" {
		req.Header.Set("Accept", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed: %d %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch p.Name {
	case "google":
		return parseGoogleUserInfo(body)
	case "github":
		return parseGitHubUserInfo(body, client)
	default:
		return nil, errors.New("unknown provider")
	}
}

func parseGoogleUserInfo(body []byte) (*UserInfo, error) {
	var data struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &UserInfo{
		ProviderUserID: data.ID,
		Email:          data.Email,
		EmailVerified:  data.VerifiedEmail,
		Name:           data.Name,
		Picture:        data.Picture,
	}, nil
}

func parseGitHubUserInfo(body []byte, client *http.Client) (*UserInfo, error) {
	var data struct {
		ID     int64  `json:"id"`
		Login  string `json:"login"`
		Name   string `json:"name"`
		Avatar string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	email, verified, err := getGitHubEmail(client)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ProviderUserID: fmt.Sprintf("%d", data.ID),
		Email:          email,
		EmailVerified:  verified,
		Name:           data.Name,
		Picture:        data.Avatar,
	}, nil
}

func getGitHubEmail(client *http.Client) (string, bool, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", false, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified"`
		Primary  bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, true, nil
		}
	}

	for _, e := range emails {
		if e.Verified {
			return e.Email, true, nil
		}
	}

	return "", false, errors.New("no verified email found")
}
