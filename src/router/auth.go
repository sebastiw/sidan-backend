package router

import (
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
)

type OAuth2Handler struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Expiry      string `json:"expiry"`
}

type GoogleUserInfo struct {
	Id string `json:"id"`
	Email string `json:"email"`
	VerifiedEmail bool `json:"verified_email"`
	Picture string `json:"picture"`
}

type GithubUserInfo struct {
	Email string `json:"email"`
	PrimaryEmail bool `json:"primary"`
	Verified bool `json:"verified"`
	Visibility string `json:"visibility"`
}

func (oh OAuth2Handler) oauth2Config() *oauth2.Config {
	switch oh.Provider {
	case "google":
		return &oauth2.Config{
			ClientID:     oh.ClientID,
			ClientSecret: oh.ClientSecret,
			RedirectURL:  oh.RedirectURL,
			Scopes:       []string{"openid", "email"},
			Endpoint:     google.Endpoint,
		}
	case "github":
		return &oauth2.Config{
			ClientID:     oh.ClientID,
			ClientSecret: oh.ClientSecret,
			RedirectURL:  oh.RedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	default:
		panic(fmt.Errorf("provider not supported %s", oh.Provider))
	}
}

func (oh OAuth2Handler) oauth2RedirectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(getRequestId(r), oh)

	conf := oh.oauth2Config()

	session, err := store.Get(r, "auth-session")
	CheckError(w, r, err)

	state := hex.EncodeToString(securecookie.GenerateRandomKey(64))
	session.AddFlash(state, "state")

	// Generate the URL to redirect the user to for authentication
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	err = session.Save(r, w)
	CheckError(w, r, err)

	// Redirect the user to the generated URL
	w.Header().Set("Content-Type", "application/json")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (oh OAuth2Handler) oauth2AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	conf := oh.oauth2Config()

	queryParams := r.URL.Query()

	code := queryParams.Get("code")
	state := queryParams.Get("state")

	session, err := store.Get(r, "auth-session")
	CheckError(w, r, err)

	sessionState := session.Flashes("state")
	if sessionState == nil || state != sessionState[0] {
		http.Error(w, "Not correct state", http.StatusInternalServerError)
		return
	}

	// Exchange the Authorization code for an Access Token
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		w.WriteHeader(401)
	}
	CheckError(w, r, err)

	session.Values["provider"] = oh.Provider
	session.Values["access_token"] = token.AccessToken
	session.Values["token_type"] = token.TokenType
	session.Values["refresh_token"] = token.RefreshToken
	session.Values["expiry"] = token.Expiry
	err = session.Save(r, w)
	CheckError(w, r, err)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (oh OAuth2Handler) retrieveEmail(w http.ResponseWriter, r *http.Request) {
	//bearer := r.Header.Get("Authorization")
	session, err := store.Get(r, "auth-session")
	CheckError(w, r, err)
	bearer := fmt.Sprintf("%s %s", session.Values["token_type"], session.Values["access_token"])

	if bearer == "" {
		err := errors.New("Empty Authorization Header")
		log.Println(getRequestId(r), err)
		return
	}

	emails := GetEmails(w, r, oh, bearer)
	if len(emails) == 0 {
		err := errors.New("No emails found")
		log.Println(getRequestId(r), err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(emails)
}

func GetUserInfoURL(oh OAuth2Handler) (string, error) {
	var err error
	url := ""
	if oh.Provider == "google" {
		url = "https://www.googleapis.com/userinfo/v2/me"
	} else if oh.Provider == "github" {
		url = "https://api.github.com/user/emails"
	} else {
		err = fmt.Errorf("Provider not supported %s", oh.Provider)
	}
	return url, err
}

func GetEmails(w http.ResponseWriter, r *http.Request, oh OAuth2Handler, bearer string) []string {
	url, err := GetUserInfoURL(oh)
	CheckError(w, r, err)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	CheckError(w, r, err)

	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		w.WriteHeader(resp.StatusCode)
	}
	CheckError(w, r, err)

	var emails []string
	if oh.Provider == "google" {
		var userInfo GoogleUserInfo
		err := json.NewDecoder(resp.Body).Decode(&userInfo)
		CheckError(w, r, err)
		if userInfo.VerifiedEmail {
			emails = append(emails, userInfo.Email)
		}
	} else if oh.Provider == "github" {
		var userInfo []GithubUserInfo
		err := json.NewDecoder(resp.Body).Decode(&userInfo)
		CheckError(w, r, err)
		for _, email := range userInfo {
			if email.Verified {
				emails = append(emails, email.Email)
			}
		}
	} else {
		err = fmt.Errorf("Provider not supported %s", oh.Provider)
		CheckError(w, r, err)
	}
	return emails
}
