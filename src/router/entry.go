package router

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	models "github.com/sebastiw/sidan-backend/src/models"
	d "github.com/sebastiw/sidan-backend/src/database/operations"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

func NewEntryHandler(db *sql.DB) EntryHandler {
	return EntryHandler{d.NewEntryOperation(db)}
}

type EntryHandler struct {
	op d.EntryOperation
}

func (eh EntryHandler) createEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	slog.Debug(ru.GetRequestId(r), e)
	entry := eh.op.Create(e)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) readEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	entry := eh.op.Read(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) updateEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), e)
	e.Id = int64(id)
	entry := eh.op.Update(e)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) deleteEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), e)
	e.Id = int64(id)
	entry := eh.op.Delete(e)

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
	entries := eh.op.ReadAll(int64(take), int64(skip))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
