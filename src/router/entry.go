package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"fmt"

	"github.com/gorilla/mux"

	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

func NewEntryHandler(db data.Database) EntryHandler {
	return EntryHandler{db}
}

type EntryHandler struct {
	db data.Database
}

func (eh EntryHandler) createEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	slog.Debug(ru.GetRequestId(r), e)
	entry, err := eh.db.CreateEntry(&e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) readEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	entry, err := eh.db.ReadEntry(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) updateEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	slog.Debug(ru.GetRequestId(r), e)
	e.Id = int64(id)
	entry, err := eh.db.UpdateEntry(&e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) deleteEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	slog.Debug(ru.GetRequestId(r), e)
	e.Id = int64(id)
	entry, err := eh.db.DeleteEntry(&e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// Responses:
//
//	default: []Entry
//	200: [Entry]
//
//swagger:route GET /db/entries entry readAllEntry
func (eh EntryHandler) readAllEntryHandler(w http.ResponseWriter, r *http.Request) {
	take := MakeDefaultInt(r, "take", "20")
	skip := MakeDefaultInt(r, "skip", "0")
	entries, err := eh.db.ReadEntries(take, skip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
