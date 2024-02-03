package auth

import (
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"strings"
	"time"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

type OAuth2Handler struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
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

func init() {
	gob.Register(&oauth2.Token{})
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

func (oh OAuth2Handler) Oauth2RedirectHandler(auth AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := oh.oauth2Config()

		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			// ignore errors due to reboots of server
			log.Println(ru.GetRequestId(r), err)
		}

		val := session.Values[oh.Provider]
		// var token = &oauth2.Token{} // unsure why this isn't needed here
		if token, ok := val.(*oauth2.Token); ok {
			// if earlier then expiry is probably indefinite
			if token.Expiry.Before(time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)) || token.Expiry.After(time.Now()) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(token)
			}
		}

		state := hex.EncodeToString(securecookie.GenerateRandomKey(64))
		session.AddFlash(state, "state")

		// Generate the URL to redirect the user to for authentication
		url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

		err = session.Save(r, w)
		if err != nil {
			log.Println(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect the user to the generated URL
		w.Header().Set("Content-Type", "application/json")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (oh OAuth2Handler) Oauth2AuthCallbackHandler(auth AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := oh.oauth2Config()

		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val := session.Values[oh.Provider]
		var token = &oauth2.Token{}
		if token, ok := val.(*oauth2.Token); ok {
			if token.Expiry.Before(time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)) || token.Expiry.After(time.Now()) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(token)
			}
		}

		queryParams := r.URL.Query()

		code := queryParams.Get("code")
		state := queryParams.Get("state")

		sessionState := session.Flashes("state")
		if sessionState == nil || state != sessionState[0] {
			err := errors.New("Invalid state")
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Exchange the Authorization code for an Access Token
		token, err = conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values[oh.Provider] = token
		sidanScopes := []string{"write:email", "write:image", "write:member", "read:member"}
		session.Values["scopes"] = sidanScopes
		err = session.Save(r, w)
		if err != nil {
			log.Println(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(token)
	}
}

func (oh OAuth2Handler) RetrieveEmail(auth AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			err := errors.New("Error getting session")
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		if session.Values[oh.Provider] == nil {
			err := fmt.Errorf("Not authorized with %s", oh.Provider)
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		token := session.Values[oh.Provider].(*oauth2.Token)
		bearer := fmt.Sprintf("%s %s", token.TokenType, token.AccessToken)

		if strings.TrimSpace(bearer) == "" {
			err := errors.New("Empty Authorization Header")
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		emails, err := GetEmails(w, r, oh, bearer)
		if err != nil {
			log.Println(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}



		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emails)
	}
}

func GetUserInfoURL(oh OAuth2Handler) (string, error) {
	var err error
	url := ""
	switch oh.Provider {
	case "google":
		url = "https://www.googleapis.com/userinfo/v2/me"
	case "github":
		url = "https://api.github.com/user/emails"
	default:
		err = fmt.Errorf("Provider not supported %s", oh.Provider)
	}
	return url, err
}

func GetEmails(w http.ResponseWriter, r *http.Request, oh OAuth2Handler, bearer string) ([]string, error) {
	url, err := GetUserInfoURL(oh)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		err := fmt.Errorf("Error getting emails: %s", resp.Status)
		return nil, err
	}

	var emails []string
	switch oh.Provider {
	case "google":
		var userInfo GoogleUserInfo
		err := json.NewDecoder(resp.Body).Decode(&userInfo)
		if err != nil {
			return nil, err
		}
		if userInfo.VerifiedEmail {
			emails = append(emails, userInfo.Email)
		}
	case "github":
		var userInfo []GithubUserInfo
		err := json.NewDecoder(resp.Body).Decode(&userInfo)
		if err != nil {
			return nil, err
		}
		for _, email := range userInfo {
			if email.Verified {
				emails = append(emails, email.Email)
			}
		}
	default:
		err = fmt.Errorf("Provider not supported %s", oh.Provider)
		return nil, err
	}

	if len(emails) == 0 {
		return nil, errors.New("No verified emails found")
	}
	return emails, nil
}