package router

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/sebastiw/sidan-backend/src/data"
	a "github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

func LogHTTP(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := statusWriter{ResponseWriter: w}
		handler.ServeHTTP(&sw, r)
		duration := time.Now().Sub(start)
		slog.Debug("http-request",
			slog.String("RequestId",  ru.GetRequestId(r)),
			slog.String("Host",       r.Host),
			slog.Duration("Duration", duration),
			slog.String("RemoteAddr", r.RemoteAddr),
			slog.String("Method",     r.Method),
			slog.String("RequestURI", r.RequestURI),
			slog.String("Proto",      r.Proto),
			slog.Int("Status",        sw.status),
			slog.Int("ContentLen",    sw.length),
			slog.String("UserAgent",  r.Header.Get("User-Agent")),
		)
	}
}

func nextRequestId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func corsHeaders(router http.Handler) http.Handler {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://api.chalmerslosers.com",
			"https://api.chalmerslosers.com:*",
			"https://chalmerslosers.com",
			"https://chalmerslosers.com:*",
			"https://sidan.cl",
			"https://sidan.cl:*",
			"http://localhost",
			"http://localhost:*",
		},
		AllowCredentials: true,
	})
	return corsHandler.Handler(router)
}

func Mux(db data.Database) http.Handler {
	r := mux.NewRouter()

	auth := a.New()
	sidanProvider := a.NewSidanAuthProvider()
	r.HandleFunc("/login/oauth/authorize", sidanProvider.BasicLoginWindow()).Methods("GET", "OPTIONS")
	r.HandleFunc("/login", sidanProvider.LoginCheck(db)).Methods("POST", "OPTIONS")
	r.HandleFunc("/login/oauth/access_token", sidanProvider.ExchangeAccess()).Methods("POST", "OPTIONS")

	// r.HandleFunc("/auth", defaultHandler)
	for provider, oauth2Config := range config.Get().OAuth2 {
		oh := a.OAuth2Handler{
			Provider:     provider,
			ClientID:     oauth2Config.ClientID,
			ClientSecret: oauth2Config.ClientSecret,
			RedirectURL:  oauth2Config.RedirectURL,
			Scopes:       oauth2Config.Scopes}
		r.HandleFunc("/auth/"+provider, oh.Oauth2RedirectHandler(auth)).Methods("GET", "OPTIONS")
		r.HandleFunc("/auth/"+provider+"/authorized", oh.Oauth2CallbackHandler(auth)).Methods("GET", "OPTIONS")
		r.HandleFunc("/auth/"+provider+"/verifyemail", oh.VerifyEmail(auth, db)).Methods("GET", "OPTIONS")
	}
	r.HandleFunc("/auth/getusersession", a.GetUserSession(auth)).Methods("GET", "OPTIONS")

	// r.HandleFunc("/notify", defaultHandler)

	fh := FileHandler{}
	fileServer := http.FileServer(http.Dir(config.GetServer().StaticPath))
	r.HandleFunc("/file/image", auth.CheckScope(fh.createImageHandler, a.WriteImageScope)).Methods("POST", "OPTIONS")
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer)).Methods("GET", "OPTIONS")

	mh := MailHandler{Host: config.GetMail().Host, Port: config.GetMail().Port, Username: config.GetMail().User, Password: config.GetMail().Password}
	r.HandleFunc("/mail", auth.CheckScope(mh.createMailHandler, a.WriteEmailScope)).Methods("POST", "OPTIONS")

	dbEh := NewEntryHandler(db)
	r.HandleFunc("/db/entries", dbEh.createEntryHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/db/entries/{id:[0-9]+}", dbEh.readEntryHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/db/entries/{id:[0-9]+}", auth.CheckScope(dbEh.updateEntryHandler, a.ModifyEntryScope)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/db/entries/{id:[0-9]+}", auth.CheckScope(dbEh.deleteEntryHandler, a.ModifyEntryScope)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/db/entries", dbEh.readAllEntryHandler).Methods("GET", "OPTIONS")

	dbMh := NewMemberHandler(db)
	r.HandleFunc("/db/members", auth.CheckScope(dbMh.createMemberHandler, a.WriteMemberScope)).Methods("POST", "OPTIONS")
	r.HandleFunc("/db/members/{id:[0-9]+}", routeAuthAndUnauthed(auth, dbMh.readMemberHandler, dbMh.readMemberUnauthedHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/db/members/{id:[0-9]+}", auth.CheckScope(dbMh.updateMemberHandler, a.WriteMemberScope)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/db/members/{id:[0-9]+}", auth.CheckScope(dbMh.deleteMemberHandler, a.WriteMemberScope)).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/db/members", routeAuthAndUnauthed(auth, dbMh.readAllMemberHandler, dbMh.readAllMemberUnauthedHandler)).Methods("GET", "OPTIONS")

	// r.HandleFunc("/db", defaultHandler)

	return corsHeaders(ru.Tracing(nextRequestId)(LogHTTP(r)))
}

func routeAuthAndUnauthed(auth a.AuthHandler, authedRoute, unauthedRoute http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.ScopeOk(w, r, a.ReadMemberScope) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			unauthedRoute(w, r)
		} else {
			authedRoute(w, r)
		}
	}
}
