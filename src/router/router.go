package router

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	ru "github.com/sebastiw/sidan-backend/src/router_util"
	c "github.com/sebastiw/sidan-backend/src/config"
	a "github.com/sebastiw/sidan-backend/src/auth"
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

type LogEntry struct {
	RequestId  string
	Host       string
	RemoteAddr string
	Method     string
	RequestURI string
	Proto      string
	Status     int
	ContentLen int
	UserAgent  string
	Duration   time.Duration
}

func LogHTTP(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := statusWriter{ResponseWriter: w}
		handler.ServeHTTP(&sw, r)
		duration := time.Now().Sub(start)
		log.Println(LogEntry{
			RequestId:  ru.GetRequestId(r),
			Host:       r.Host,
			RemoteAddr: r.RemoteAddr,
			Method:     r.Method,
			RequestURI: r.RequestURI,
			Proto:      r.Proto,
			Status:     sw.status,
			ContentLen: sw.length,
			UserAgent:  r.Header.Get("User-Agent"),
			Duration:   duration,
		})
	}
}

func nextRequestId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func corsHeaders(router http.Handler) http.Handler {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})
	return corsHandler.Handler(router)
}

func Mux(db *sql.DB, staticPath string, mailConfig c.MailConfiguration, oauth2Configs map[string]c.OAuth2Configuration) http.Handler {
	r := mux.NewRouter()

	auth := a.New()
	// r.HandleFunc("/auth", defaultHandler)
	for provider, oauth2Config := range oauth2Configs {
		oh := a.OAuth2Handler{
			Provider: provider,
			ClientID: oauth2Config.ClientID,
			ClientSecret: oauth2Config.ClientSecret,
			RedirectURL: oauth2Config.RedirectURL,
			Scopes: oauth2Config.Scopes}
		r.HandleFunc("/auth/"+provider, oh.Oauth2RedirectHandler(auth)).Methods("GET", "OPTIONS")
		r.HandleFunc("/auth/"+provider+"/authorized", oh.Oauth2CallbackHandler(auth)).Methods("GET", "OPTIONS")
		r.HandleFunc("/auth/"+provider+"/verifyemail", oh.VerifyEmail(auth, db)).Methods("GET", "OPTIONS")
	}

	// r.HandleFunc("/notify", defaultHandler)

	fh := FileHandler{}
	fileServer := http.FileServer(http.Dir(staticPath))
	r.HandleFunc("/file/image", auth.CheckScope(fh.createImageHandler, a.WriteImageScope)).Methods("POST", "OPTIONS")
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer)).Methods("GET", "OPTIONS")

	mh := MailHandler{Host: mailConfig.Host, Port: mailConfig.Port, Username: mailConfig.User, Password: mailConfig.Password}
	r.HandleFunc("/mail", auth.CheckScope(mh.createMailHandler, a.WriteEmailScope)).Methods("POST", "OPTIONS")

	dbEh := NewEntryHandler(db)
	r.HandleFunc("/db/entries", dbEh.createEntryHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/db/entries/{id:[0-9]+}", dbEh.readEntryHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/db/entries/{id:[0-9]+}", auth.CheckScope(dbEh.updateEntryHandler, a.ModifyEntryScope)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/db/entries/{id:[0-9]+}", auth.CheckScope(dbEh.deleteEntryHandler, a.ModifyEntryScope)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/db/entries", dbEh.readAllEntryHandler).Methods("GET", "OPTIONS")

	dbMh := NewMemberHandler(db)
	r.HandleFunc("/db/members", auth.CheckScope(dbMh.createMemberHandler, a.WriteMemberScope)).Methods("POST", "OPTIONS")
	r.HandleFunc("/db/members/{id:[0-9]+}", dbMh.readMemberHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/db/members/{id:[0-9]+}", auth.CheckScope(dbMh.updateMemberHandler, a.WriteMemberScope)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/db/members/{id:[0-9]+}", auth.CheckScope(dbMh.deleteMemberHandler, a.WriteMemberScope)).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/db/members", dbMh.readAllMemberHandler).Methods("GET", "OPTIONS")

	// r.HandleFunc("/db", defaultHandler)

	return corsHeaders(ru.Tracing(nextRequestId)(LogHTTP(r)))
}
