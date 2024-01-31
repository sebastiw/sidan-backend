package router

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	c "github.com/sebastiw/sidan-backend/src/config"
)

type key int

const (
	requestIDKey key = 0
)

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
			RequestId:  getRequestId(r),
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
func getRequestId(r *http.Request) string {
	requestID, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		requestID = "unknown"
	}
	return requestID
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

	// r.HandleFunc("/auth", defaultHandler)
	for provider, oauth2Config := range oauth2Configs {
		oh := OAuth2Handler{
			Provider: provider,
			ClientID: oauth2Config.ClientID,
			ClientSecret: oauth2Config.ClientSecret,
			RedirectURL: oauth2Config.RedirectURL,
			Scopes: oauth2Config.Scopes}
		r.HandleFunc("/auth/"+provider, oh.oauth2RedirectHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/auth/"+provider+"/authorized", oh.oauth2AuthCallbackHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/auth/"+provider+"/getemail", oh.retrieveEmail).Methods("GET", "OPTIONS")
	}

	// r.HandleFunc("/notify", defaultHandler)

	fh := FileHandler{}
	fileServer := http.FileServer(http.Dir(staticPath))
	r.HandleFunc("/file/image", fh.createImageHandler).Methods("POST", "OPTIONS")
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer)).Methods("GET", "OPTIONS")

	mh := MailHandler{Host: mailConfig.Host, Port: mailConfig.Port, Username: mailConfig.User, Password: mailConfig.Password}
	r.HandleFunc("/mail", mh.createMailHandler).Methods("POST", "OPTIONS")

	dbEh := NewEntryHandler(db)
	//swagger:route POST /db/entries entry createEntry
	r.HandleFunc("/db/entries", dbEh.createEntryHandler).Methods("POST", "OPTIONS")
	//swagger:route GET /db/entries/{id} entry readEntry
	//	Parameters:
	//    + name: id
	//      in: path
	//  	format: int32
	//	Responses:
	//  	200: Entry
	r.HandleFunc("/db/entries/{id:[0-9]+}", dbEh.readEntryHandler).Methods("GET", "OPTIONS")
	//swagger:route PUT /db/entries/{id} entry updateEntry
	r.HandleFunc("/db/entries/{id:[0-9]+}", dbEh.updateEntryHandler).Methods("PUT", "OPTIONS")
	//swagger:route DELETE /db/entries/{id} entry deleteEntry
	r.HandleFunc("/db/entries/{id:[0-9]+}", dbEh.deleteEntryHandler).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/db/entries", dbEh.readAllEntryHandler).Methods("GET", "OPTIONS")

	dbMh := NewMemberHandler(db)
	//swagger:route POST /db/members member createMember
	r.HandleFunc("/db/members", dbMh.createMemberHandler).Methods("POST", "OPTIONS")
	//swagger:route GET /db/members/{id} member readMember
	r.HandleFunc("/db/members/{id:[0-9]+}", dbMh.readMemberHandler).Methods("GET", "OPTIONS")
	//swagger:route PUT /db/members/{id} member updateMember
	r.HandleFunc("/db/members/{id:[0-9]+}", dbMh.updateMemberHandler).Methods("PUT", "OPTIONS")
	//swagger:route DELETE /db/members/{id} member deleteMember
	r.HandleFunc("/db/members/{id:[0-9]+}", dbMh.deleteMemberHandler).Methods("DELETE", "OPTIONS")
	//swagger:route GET /db/members member readAllMember
	r.HandleFunc("/db/members", dbMh.readAllMemberHandler).Methods("GET", "OPTIONS")

	// r.HandleFunc("/db", defaultHandler)

	return corsHeaders(tracing(nextRequestId)(LogHTTP(r)))
}
