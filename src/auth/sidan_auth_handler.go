package auth

import (
	// "crypto/sha256"
	// "encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"log/slog"
	"net/http"

	"github.com/sebastiw/sidan-backend/src/data"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

// store will hold all temp data
var providerStore *sessions.CookieStore

type SidanAuthProvider struct {
	Store *sessions.CookieStore
}

func NewSidanAuthProvider() SidanAuthProvider {
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	providerStore = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	providerStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60,
		HttpOnly: true,
	}

	return SidanAuthProvider{Store: providerStore}
}

func (s SidanAuthProvider) BasicLoginWindow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := s.Store.Get(r, "provider-session")
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error getting session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectUrl := r.URL.Query().Get("redirect_uri")
		if redirectUrl != "" {
			session.AddFlash(redirectUrl, "redirectUrl")
		}

		redirectState := r.URL.Query().Get("state")
		session.Values["redirect_state"] = redirectState
		slog.Debug(ru.GetRequestId(r), "redirect state=", redirectState)

		random := hex.EncodeToString(securecookie.GenerateRandomKey(64))
		session.Values["state"] = random
		slog.Debug(ru.GetRequestId(r), "sidan state=", random)

		err = s.Store.Save(r, w, session)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "./src/auth/login.html")
	}
}

func (s SidanAuthProvider) LoginCheck(db data.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")
		_, err := db.GetUserFromLogin(username, password)

		if err != nil {
			// Invalid credentials, show the login page with an error message.
			slog.Warn(ru.GetRequestId(r), err)
			fmt.Fprintf(w, "Invalid credentials. Please try again.")
			return
		}

		session, err := s.Store.Get(r, "provider-session")
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error getting session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state := session.Values["state"]
		if state == nil {
			slog.Warn(ru.GetRequestId(r), err)
			http.Error(w, "Incorrect code", http.StatusInternalServerError)
			return
		}

		redirectState := session.Values["redirect_state"]

		code := hex.EncodeToString(securecookie.GenerateRandomKey(64))
		session.Values["code"] = code

		err = s.Store.Save(r, w, session)
		if err != nil {
			slog.Warn(ru.GetRequestId(r), err)
			err := errors.New("Error saving session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectUrl := session.Flashes("redirectUrl")
		if redirectUrl == nil || redirectUrl[0] == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.Redirect(w, r, "/welcome", http.StatusSeeOther)
		} else {
			url := fmt.Sprintf("%s?code=%s&state=%s", redirectUrl[0], code, redirectState)
			http.Redirect(w, r, url, http.StatusSeeOther)
		}
	}
}

func (s SidanAuthProvider) ExchangeAccess() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		if code == "" {
			slog.Warn(ru.GetRequestId(r), "Empty code")
			http.Error(w, "Incorrect code", http.StatusInternalServerError)
			return
		}

		slog.Info(ru.GetRequestId(r), code)
		http.Error(w, "CODODCODOCODOCDOCODOCODODOCODOCODOCODOODC code", http.StatusInternalServerError)
		return
	}
}
