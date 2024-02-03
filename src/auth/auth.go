package auth

import (
	"net/http"
	"errors"
	"time"
	"golang.org/x/oauth2"
	"github.com/gorilla/sessions"
	"github.com/gorilla/securecookie"
)

const WriteEmailScope = "write:email"
const WriteImageScope = "write:image"
const WriteMemberScope = "write:member"
const ReadMemberScope = "read:member"

var providers = []string{"google", "github"}

// store will hold all session data
var store *sessions.CookieStore

type AuthHandler struct {
	Store *sessions.CookieStore
}

func New() AuthHandler {
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	store = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
	}
	return AuthHandler{Store: store}
}

func (a AuthHandler) CheckScope(router http.HandlerFunc, scope string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.ScopeOk(w, r, ReadMemberScope) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		router(w, r)
	}
}

func (auth AuthHandler) getToken(w http.ResponseWriter, r *http.Request) (*oauth2.Token, error) {
	session, err := auth.Store.Get(r, "auth-session")
	if err != nil {
		return nil, err
	}

	for _, provider := range providers {
		val := session.Values[provider]
		// var token = &oauth2.Token{}
		if token, ok := val.(*oauth2.Token); ok {
			if token.Expiry.Before(time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)) || token.Expiry.After(time.Now()) {
				return token, nil
			}
		}
	}

	return nil, errors.New("Token not found")
}

func (auth AuthHandler) ScopeOk(w http.ResponseWriter, r *http.Request, scope string) bool {
	session, err := auth.Store.Get(r, "auth-session")
	if err != nil {
		return false
	}

	_, err = auth.getToken(w, r)
	if err != nil {
		return false
	}

	val := session.Values["scopes"]
	if scopes, ok := val.([]string); ok {
		for _, s := range scopes {
			if s == scope {
				return true
			}
		}
	}
	return false
}

