package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// m "github.com/sebastiw/sidan-backend/src/database/models"
	d "github.com/sebastiw/sidan-backend/src/database/operations"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

type SidanAuthProvider struct {
	ClientID     string
	ClientSecret string
}

func NewSidanAuthProvider() SidanAuthProvider {
	return SidanAuthProvider{}
}

func (s SidanAuthProvider) BasicLoginWindow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "./src/auth/login.html")
	}
}

func (s SidanAuthProvider) LoginCheck(db *sql.DB) http.HandlerFunc {
	usr := d.NewUserOperation(db)
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")
		user, err := usr.GetUserFromLogin(username, password)

		if err == nil {
			log.Println(ru.GetRequestId(r), "works " + user.Type)
			// Successful login, redirect to a welcome page.
			http.Redirect(w, r, "/welcome", http.StatusSeeOther)
			return
		}

		// Invalid credentials, show the login page with an error message.
		log.Println(ru.GetRequestId(r), err)
		fmt.Fprintf(w, "Invalid credentials. Please try again.")
		return
	}
}

