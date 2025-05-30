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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
	// d "github.com/sebastiw/sidan-backend/src/database/operations"
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

type SessionInfo struct {
	Scopes []string `json:"scopes"`
	UserName string `json:"username"`
	Email string `json:"email"`
	FulHaxPass string `json:"fulHaxPass"`
}


func init() {
	gob.Register(&oauth2.Token{})
	gob.Register(&models.User{})
}

func (oh OAuth2Handler) get_endpoint() *oauth2.Endpoint {
	switch oh.Provider {
	case "sidan":
		return &oauth2.Endpoint{
			AuthURL:       "/login/oauth/authorize",
			TokenURL:      "http://localhost:8080/login/oauth/access_token",
		}
	case "google":
		return &google.Endpoint
	case "github":
		return &github.Endpoint
	default:
		panic(fmt.Errorf("provider not supported %s", oh.Provider))
	}
}
func (oh OAuth2Handler) oauth2Config() *oauth2.Config {
	endpoint := oh.get_endpoint()
	return &oauth2.Config{
			ClientID:     oh.ClientID,
			ClientSecret: oh.ClientSecret,
			RedirectURL:  oh.RedirectURL,
			Scopes:       oh.Scopes,
			Endpoint:     *endpoint,
		}
}

func (oh OAuth2Handler) Oauth2RedirectHandler(auth AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := oh.oauth2Config()

		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			// ignore errors due to reboots of server
			slog.Warn(ru.GetRequestId(r), err)
		}

		// Remember callback url
		queryParams := r.URL.Query()
		redirectUrl := queryParams.Get("redirect_uri")
		if redirectUrl != "" {
			session.AddFlash(redirectUrl, "redirectUrl")
		}

		val := session.Values[oh.Provider]
		// var token = &oauth2.Token{} // unsure why this isn't needed here
		if token, ok := val.(*oauth2.Token); ok {
			// if earlier then expiry is probably indefinite
			if token.Expiry.Before(time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)) || token.Expiry.After(time.Now()) {
				http.Redirect(w, r, "/auth/" + oh.Provider + "/authorized", http.StatusTemporaryRedirect)
				return
			}
		}

		state := hex.EncodeToString(securecookie.GenerateRandomKey(64))
		session.AddFlash(state, "state")

		// Generate the URL to redirect the user to for authentication
		url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

		err = session.Save(r, w)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect the user to the generated URL
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (oh OAuth2Handler) Oauth2CallbackHandler(auth AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := oh.oauth2Config()

		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val := session.Values[oh.Provider]
		var token = &oauth2.Token{}
		if token, ok := val.(*oauth2.Token); ok {
			if token.Expiry.Before(time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)) || token.Expiry.After(time.Now()) {
				http.Redirect(w, r, "/auth/" + oh.Provider + "/verifyemail", http.StatusTemporaryRedirect)
				return
			}
		}

		queryParams := r.URL.Query()
		code := queryParams.Get("code")
		state := queryParams.Get("state")
		sessionState := session.Flashes("state")
		if sessionState == nil || state != sessionState[0] {
			err := errors.New("Invalid state")
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Exchange the Authorization code for an Access Token
		token, err = conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values[oh.Provider] = token
		err = session.Save(r, w)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Not sure if redirecting the user to verify email is
		// lazy and we should incorporate VerifyEmail here, or
		// if it's a good idea to keep both APIs separate
		http.Redirect(w, r, "/auth/" + oh.Provider + "/verifyemail", http.StatusTemporaryRedirect)
	}
}

func (oh OAuth2Handler) VerifyEmail(auth AuthHandler, db data.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			err := errors.New("Error getting session")
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		if session.Values[oh.Provider] == nil {
			err := fmt.Errorf("Not authorized with %s", oh.Provider)
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		token := session.Values[oh.Provider].(*oauth2.Token)
		bearer := fmt.Sprintf("%s %s", token.TokenType, token.AccessToken)

		if strings.TrimSpace(bearer) == "" {
			err := errors.New("Empty Authorization Header")
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		emails, err := GetEmailsFromProvider(w, r, oh, bearer)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := db.GetUserFromEmails(emails)
		if err != nil {
			// Probably clean up session here
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Email not registered with user")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		sidanScopes := getScopes(user.Type)
		slog.Info(ru.GetRequestId(r), "User found: ", string(user.Type) + string(user.Number), "<" + user.Email + ">", sidanScopes)
		session.Values["scopes"] = sidanScopes
		session.Values["user"] = user

		err = session.Save(r, w)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectUrl := session.Flashes("redirectUrl")
		if redirectUrl == nil || redirectUrl[0] == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, "./src/auth/close.html")
		} else {
			url := fmt.Sprintf("%s?bearer=%s", redirectUrl[0], bearer)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}
	}
}

func getScopes(userType models.UserType) []string {
	switch userType {
	case models.MemberType:
		return []string{WriteEmailScope, WriteImageScope, WriteMemberScope, ReadMemberScope}
	case models.ProspectType:
		return []string{WriteEmailScope, WriteImageScope, ReadMemberScope}
	default:
		return []string{}
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

func GetEmailsFromProvider(w http.ResponseWriter, r *http.Request, oh OAuth2Handler, bearer string) ([]string, error) {
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

func GetUserSession(auth AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = auth.getToken(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		vals := session.Values["scopes"]
		sidanScopes, oks := vals.([]string)
		if !oks {
			http.Error(w, "Couldn't get user", http.StatusInternalServerError)
			return
		}

		valu := session.Values["user"]
		var user = &models.User{}
		user, oku := valu.(*models.User)
		if !oku {
			http.Error(w, "Couldn't get user", http.StatusInternalServerError)
			return
		}

		sessInfo := SessionInfo{
			Scopes: sidanScopes,
			UserName: string(user.Type) + string(user.Number),
			Email: user.Email,
			FulHaxPass: user.FulHaxPass,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessInfo)
	}
}
