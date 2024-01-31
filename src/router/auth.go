package router

import (
	"fmt"
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

type OAuth2AuthToken struct {
	Code        string `json:"code"`
	State       string `json:"state"`
}

type OAuth2AccessToken struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
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

	// Generate the URL to redirect the user to for authentication
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	// Redirect the user to the generated URL
	w.Header().Set("Content-Type", "application/json")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (oh OAuth2Handler) oauth2AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	conf := oh.oauth2Config()

	queryParams := r.URL.Query()

	code := queryParams.Get("code")
	state := queryParams.Get("state")

	// Exchange the Authorization code for an Access Token
	e := OAuth2AuthToken{Code: code, State: state}
	token, err := conf.Exchange(oauth2.NoContext, e.Code)
	if err != nil {
		log.Printf("ERROR: %s, %s", err.Error(), code)
		return
	}

	log.Println(getRequestId(r), token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (oh Oauth2Handler) retrieveEmail(w http.ResponseWriter, r *http.Request) {
	conf := oh.oauth2Config()

	bearer := r.Header.Get("Authorization")
	if !bearer {
		log.Printf("ERROR: Empty Authorization Header")
		return
	}

	url := ""
	if conf.provider == "google" {
		url = "https://www.googleapis.com//userinfo/v2/me"
	} else if conf.provider == "github" {
		url = "https://api.github.com/user/emails"
	} else {
		log.Printf("ERROR: Provider not supported %s", conf.provider)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return
	}
	req.Header.Set("Authorization", bearer)
	resp, err := http.Get(url)

	json.NewEncoder(w).Encode(resp)
}
