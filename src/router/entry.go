package router

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	d "github.com/sebastiw/sidan-backend/src/database/operations"
	model "github.com/sebastiw/sidan-backend/src/database/models"
)

func NewEntryHandler(db *sql.DB) EntryHandler {
	return EntryHandler{d.NewEntryOperation(db)}
}

type EntryHandler struct {
	op d.EntryOperation
}

func (eh EntryHandler) createEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e model.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	log.Println(get_request_id(r), e.Fmt())
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
	var e model.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Println(get_request_id(r), e.Fmt())
	e.Id = int64(id)
	entry := eh.op.Update(e)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) deleteEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e model.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Println(get_request_id(r), e.Fmt())
	e.Id = int64(id)
	entry := eh.op.Delete(e)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) readAllEntryHandler(w http.ResponseWriter, r *http.Request) {
	take := MakeDefaultInt(r, "take", "20")
	skip := MakeDefaultInt(r, "skip", "0")
	entries := eh.op.ReadAll(int64(take), int64(skip))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
