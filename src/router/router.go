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

	// Create middleware for protected endpoints
	authMiddleware := a.NewMiddleware(db)
	
	// Auth handlers (public endpoints)
	authHandler := NewAuthHandler(db)
	r.HandleFunc("/auth/login", authHandler.Login).Methods("GET", "OPTIONS")
	r.HandleFunc("/auth/callback", authHandler.Callback).Methods("GET", "OPTIONS")
	
	// Auth handlers - authenticated endpoints
	r.Handle("/auth/session", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.GetSession))).Methods("GET", "OPTIONS")
	r.Handle("/auth/refresh", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Refresh))).Methods("POST", "OPTIONS")
	r.Handle("/auth/logout", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Logout))).Methods("POST", "OPTIONS")

	// Start cleanup job (runs every 15 minutes)
	a.StartCleanupJob(db, 15*time.Minute)


	// File endpoints
	fh := FileHandler{}
	fileServer := http.FileServer(http.Dir(config.GetServer().StaticPath))
	r.Handle("/file/image", 
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.WriteImageScope)(
				http.HandlerFunc(fh.createImageHandler),
			),
		),
	).Methods("POST", "OPTIONS")
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer)).Methods("GET", "OPTIONS")

	// Mail endpoint
	mh := MailHandler{Host: config.GetMail().Host, Port: config.GetMail().Port, Username: config.GetMail().User, Password: config.GetMail().Password}
	r.Handle("/mail",
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.WriteEmailScope)(
				http.HandlerFunc(mh.createMailHandler),
			),
		),
	).Methods("POST", "OPTIONS")

	// Entry endpoints
	dbEh := NewEntryHandler(db)
	r.HandleFunc("/db/entries", dbEh.createEntryHandler).Methods("POST", "OPTIONS")
	r.Handle("/db/entries/{id:[0-9]+}",
		authMiddleware.OptionalAuth(http.HandlerFunc(dbEh.readEntryHandler)),
	).Methods("GET", "OPTIONS")
	r.Handle("/db/entries/{id:[0-9]+}",
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.ModifyEntryScope)(
				http.HandlerFunc(dbEh.updateEntryHandler),
			),
		),
	).Methods("PUT", "OPTIONS")
	r.Handle("/db/entries/{id:[0-9]+}",
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.ModifyEntryScope)(
				http.HandlerFunc(dbEh.deleteEntryHandler),
			),
		),
	).Methods("DELETE", "OPTIONS")
	r.Handle("/db/entries",
		authMiddleware.OptionalAuth(http.HandlerFunc(dbEh.readAllEntryHandler)),
	).Methods("GET", "OPTIONS")

	// Member endpoints (with optional auth for read operations)
	dbMh := NewMemberHandler(db)
	r.Handle("/db/members",
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.WriteMemberScope)(
				http.HandlerFunc(dbMh.createMemberHandler),
			),
		),
	).Methods("POST", "OPTIONS")
	r.Handle("/db/members/{id:[0-9]+}",
		authMiddleware.OptionalAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if authenticated
				scopes := a.GetScopes(r)
				if scopes != nil {
					// Check for read:member scope
					hasScope := false
					for _, s := range scopes {
						if s == a.ReadMemberScope {
							hasScope = true
							break
						}
					}
					if hasScope {
						dbMh.readMemberHandler(w, r)
						return
					}
				}
				// Not authenticated or no scope - return limited data
				dbMh.readMemberUnauthedHandler(w, r)
			}),
		),
	).Methods("GET", "OPTIONS")
	r.Handle("/db/members/{id:[0-9]+}",
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.WriteMemberScope)(
				http.HandlerFunc(dbMh.updateMemberHandler),
			),
		),
	).Methods("PUT", "OPTIONS")
	r.Handle("/db/members/{id:[0-9]+}",
		authMiddleware.RequireAuth(
			authMiddleware.RequireScope(a.WriteMemberScope)(
				http.HandlerFunc(dbMh.deleteMemberHandler),
			),
		),
	).Methods("DELETE", "OPTIONS")
	r.Handle("/db/members",
		authMiddleware.OptionalAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if authenticated
				scopes := a.GetScopes(r)
				if scopes != nil {
					// Check for read:member scope
					hasScope := false
					for _, s := range scopes {
						if s == a.ReadMemberScope {
							hasScope = true
							break
						}
					}
					if hasScope {
						dbMh.readAllMemberHandler(w, r)
						return
					}
				}
				// Not authenticated or no scope - return limited data
				dbMh.readAllMemberUnauthedHandler(w, r)
			}),
		),
	).Methods("GET", "OPTIONS")

	// r.HandleFunc("/db", defaultHandler)

	return corsHeaders(ru.Tracing(nextRequestId)(LogHTTP(r)))
}
