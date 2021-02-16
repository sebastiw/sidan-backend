package router

import(
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
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
			requestID, ok := r.Context().Value(requestIDKey).(string)
			if !ok {
				requestID = "unknown"
			}
			log.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		}()
		h.ServeHTTP(w, r)
	})
}

func Mux() http.Handler {
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Authentication")
	})

	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "File")
	})

	mux.HandleFunc("/mail", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Mail")
	})

	mux.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Notification")
	})

	mux.HandleFunc("/db", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Database")
	})

	return tracing(nextRequestID)(logging(mux))
}
