package router

import(
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	d "github.com/sebastiw/sidan-backend/src/database/operations"
	model "github.com/sebastiw/sidan-backend/src/database/models"
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

func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			log.Println(get_request_id(r), r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		}()
		h.ServeHTTP(w, r)
	})
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

func Mux(db *sql.DB) http.Handler {
	r := mux.NewRouter()

	// r.HandleFunc("/auth", defaultHandler)
	// r.HandleFunc("/file", defaultHandler)
	// r.HandleFunc("/mail", defaultHandler)
	// r.HandleFunc("/notify", defaultHandler)
	r.HandleFunc("/db/member", func (w http.ResponseWriter, r *http.Request) {
		var m model.Member
		_ = json.NewDecoder(r.Body).Decode(&m)

		log.Println(get_request_id(r), m.Fmt())
		member := d.Create(db, m)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(member)
	}).Methods("POST")

	r.HandleFunc("/db/member/{id:[0-9]+}", func (w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.Atoi(vars["id"])

		member := d.Read(db, id)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(member)
	}).Methods("GET")

	r.HandleFunc("/db/members", func (w http.ResponseWriter, r *http.Request) {
		members := d.ReadAll(db)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(members)
	})

	// r.HandleFunc("/db", defaultHandler)

	return tracing(next_request_id)(logging(r))
}
