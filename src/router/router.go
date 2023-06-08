package router

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

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
			RequestId:  get_request_id(r),
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

func next_request_id() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
func get_request_id(r *http.Request) string {
	requestID, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		requestID = "unknown"
	}
	return requestID
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Placeholder")
}

func Mux(db *sql.DB, staticPath string, mailConfig c.MailConfiguration, oauth2Configs map[string]c.OAuth2Configuration) http.Handler {
	r := mux.NewRouter()

	// r.HandleFunc("/auth", defaultHandler)
	// r.HandleFunc("/notify", defaultHandler)
	for provider, oauth2Config := range oauth2Configs {
		oh := OAuth2Handler{Provider: provider, ClientID: oauth2Config.ClientID, ClientSecret: oauth2Config.ClientSecret, RedirectURL: oauth2Config.RedirectURL, Scopes: oauth2Config.Scopes}
		r.HandleFunc("/auth/" + provider, oh.oauth2RedirectHandler).Methods("GET")
	}

	fh := FileHandler{}
	fileServer := http.FileServer(http.Dir(staticPath))
	r.HandleFunc("/file/image", fh.createImageHandler).Methods("PUT")
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer)).Methods("GET")

	mh := MailHandler{Host: mailConfig.Host, Port: mailConfig.Port, Username: mailConfig.User, Password: mailConfig.Password}
	r.HandleFunc("/mail", mh.createMailHandler).Methods("PUT")

	db_eh := NewEntryHandler(db)
	//swagger:route POST /db/entry entry createEntry
	r.HandleFunc("/db/entry", db_eh.createEntryHandler).Methods("POST")
	//swagger:route GET /db/entry/{id} entry readEntry
	//	Parameters:
	//    + name: id
	//      in: path
	//  	format: int32
	//	Responses:
	//  	200: Entry
	r.HandleFunc("/db/entry/{id:[0-9]+}", db_eh.readEntryHandler).Methods("GET")
	//swagger:route PUT /db/entry/{id} entry updateEntry
	r.HandleFunc("/db/entry/{id:[0-9]+}", db_eh.updateEntryHandler).Methods("PUT")
	//swagger:route DELETE /db/entry/{id} entry deleteEntry
	r.HandleFunc("/db/entry/{id:[0-9]+}", db_eh.deleteEntryHandler).Methods("DELETE")

	r.HandleFunc("/db/entries", db_eh.readAllEntryHandler).Methods("GET")

	db_mh := NewMemberHandler(db)
	//swagger:route POST /db/member member createMember
	r.HandleFunc("/db/member", db_mh.createMemberHandler).Methods("POST")
	//swagger:route GET /db/member/{id} member readMember
	r.HandleFunc("/db/member/{id:[0-9]+}", db_mh.readMemberHandler).Methods("GET")
	//swagger:route PUT /db/member/{id} member updateMember
	r.HandleFunc("/db/member/{id:[0-9]+}", db_mh.updateMemberHandler).Methods("PUT")
	//swagger:route DELETE /db/member/{id} member deleteMember
	r.HandleFunc("/db/member/{id:[0-9]+}", db_mh.deleteMemberHandler).Methods("DELETE")
	//swagger:route GET /db/members member readAllMember
	r.HandleFunc("/db/members", db_mh.readAllMemberHandler).Methods("GET")

	// r.HandleFunc("/db", defaultHandler)

	// Sidan APIs for deprecation
	restv1h := RestHandler{version: 1, db: db, user_id: "8"} // authed user goes here
	v1Rest := r.PathPrefix("/rest/v1").Subrouter()
	v1Rest.HandleFunc("/ReadEntries", restv1h.getEntries).Methods("GET")

	return tracing(next_request_id)(LogHTTP(r))
}
