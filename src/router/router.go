package router

import(
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	RequestId string
	Host string
	RemoteAddr string
	Method string
	RequestURI string
	Proto string
	Status int
	ContentLen int
	UserAgent string
	Duration time.Duration
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

func Mux(db *sql.DB, staticPath string) http.Handler {
	r := mux.NewRouter()

	// r.HandleFunc("/auth", defaultHandler)
	// r.HandleFunc("/mail", defaultHandler)
	// r.HandleFunc("/notify", defaultHandler)

	fh := FileHandler{}
	fileServer := http.FileServer(http.Dir(staticPath))
	r.HandleFunc("/file/image", fh.createImageHandler).Methods("PUT")
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer)).Methods("GET")

	mh := MemberHandler{db: db}

	r.HandleFunc("/db/member", mh.createMemberHandler).Methods("PUT")
	r.HandleFunc("/db/member/{id:[0-9]+}", mh.readMemberHandler).Methods("GET")
	r.HandleFunc("/db/member/{id:[0-9]+}", mh.updateMemberHandler).Methods("POST")
	r.HandleFunc("/db/member/{id:[0-9]+}", mh.deleteMemberHandler).Methods("DELETE")
	r.HandleFunc("/db/members", mh.readAllMemberHandler).Methods("GET")

	// r.HandleFunc("/db", defaultHandler)

	return tracing(next_request_id)(LogHTTP(r))
}

type FileHandler struct {
}

type File struct {
	Filename string `json:filename`
}

func file_extension(image_type string) string {
	switch image_type {
	case "image/gif":
		return "gif"
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpeg"
	}
	return ""
}

func (fh FileHandler) createImageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files. (bitshift 10 in decimal 20 times)
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("data")
	if err != nil {
		log.Println(get_request_id(r), "Error Retrieving the File", err)
		return
	}
	defer file.Close()

	// DetectContentType take only first 512 bytes into consideration
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		fmt.Println(err)
		return
	}

	contentType := http.DetectContentType(buff)
	fileExt := file_extension(contentType)
	if "" == fileExt {
		log.Println(get_request_id(r), "Not an image file")
		return
	}

	tempFilename := fmt.Sprintf("upload-*.%s", fileExt)
	tempFile, err := ioutil.TempFile("static", tempFilename)
	if err != nil {
		log.Println(get_request_id(r), err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(get_request_id(r), err)
	}
	tempFile.Write(fileBytes)
	bareFilename := strings.TrimPrefix(tempFile.Name(), "static/")
	size := fmt.Sprintf("%+vb", handler.Size)
	log.Println(get_request_id(r), "Uploaded", handler.Filename, size, bareFilename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(File{Filename: bareFilename})
}

type MemberHandler struct {
	db *sql.DB
}

func (mh MemberHandler) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	log.Println(get_request_id(r), m.Fmt())
	member := d.Create(mh.db, m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) readMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	member := d.Read(mh.db, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) updateMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Println(get_request_id(r), m.Fmt())
	m.Id = int64(id)
	member := d.Update(mh.db, m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) deleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Println(get_request_id(r), m.Fmt())
	m.Id = int64(id)
	member := d.Delete(mh.db, m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) readAllMemberHandler(w http.ResponseWriter, r *http.Request) {
	members := d.ReadAll(mh.db)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

