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

// @Summary Create entry
// @Tags entries
// @Accept json
// @Produce json
// @Param entry body models.Entry true "Entry data"
// @Success 200 {object} models.Entry
// @Router /db/entries [post]
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

// @Summary Get entry by ID
// @Tags entries
// @Produce json
// @Param id path int true "Entry ID"
// @Success 200 {object} models.Entry
// @Router /db/entries/{id} [get]
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

// @Summary Update entry
// @Tags entries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Entry ID"
// @Param entry body models.Entry true "Entry data"
// @Success 200 {object} models.Entry
// @Router /db/entries/{id} [put]
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

// @Summary Delete entry
// @Tags entries
// @Security BearerAuth
// @Param id path int true "Entry ID"
// @Success 200 {object} models.Entry
// @Router /db/entries/{id} [delete]
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

// @Summary List all entries
// @Tags entries
// @Produce json
// @Param take query int false "Number of entries to return" default(20)
// @Param skip query int false "Number of entries to skip" default(0)
// @Success 200 {array} models.Entry
// @Router /db/entries [get]
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
