package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

// GetMemberFromContext retrieves member from auth context or returns nil
func GetMemberFromContext(r *http.Request) *models.Member {
	return auth.GetMember(r)
}

func NewEntryHandler(db data.Database) EntryHandler {
	return EntryHandler{db}
}

type EntryHandler struct {
	db data.Database
}

func (eh EntryHandler) createEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	slog.Debug(ru.GetRequestId(r), "entry", e)
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
		return
	}

	// Get viewer member ID from auth context (nil if unauthenticated)
	var viewerMemberID *int64
	member := GetMemberFromContext(r)
	if member != nil {
		viewerMemberID = &member.Number
	}

	// Apply message filtering based on permissions
	FilterEntryMessage(entry, viewerMemberID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (eh EntryHandler) updateEntryHandler(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	_ = json.NewDecoder(r.Body).Decode(&e)

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	slog.Debug(ru.GetRequestId(r), "entry", e)
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

	slog.Debug(ru.GetRequestId(r), "entry", e)
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
	rsqlQuery := r.URL.Query().Get("q")
	
	// Security check: filtering requires 'filtering' scope
	if rsqlQuery != "" {
		scopes := auth.GetScopes(r)
		hasFilteringScope := false
		if scopes != nil {
			for _, scope := range scopes {
				if scope == auth.FilteringScope {
					hasFilteringScope = true
					break
				}
			}
		}
		
		if !hasFilteringScope {
			w.WriteHeader(http.StatusForbidden)
			http.Error(w, "filtering requires 'filtering' scope", http.StatusForbidden)
			return
		}
	}
	
	// Pass raw RSQL query to database layer
	entries, err := eh.db.ReadEntries(take, skip, rsqlQuery)
	if err != nil {
		// Check if it's an RSQL parsing error (400) vs database error (500)
		if strings.Contains(err.Error(), "RSQL") || strings.Contains(err.Error(), "not allowed") {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, fmt.Sprintf("invalid RSQL query: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	// Get viewer member ID from auth context (nil if unauthenticated)
	var viewerMemberID *int64
	member := GetMemberFromContext(r)
	if member != nil {
		viewerMemberID = &member.Number
	}

	// Apply message filtering to all entries
	FilterEntriesMessages(entries, viewerMemberID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (eh EntryHandler) likeEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)

	member := auth.GetMember(r)
	if member == nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sig := strconv.FormatInt(member.Number, 10)
	host := r.RemoteAddr
	err := eh.db.LikeEntry(id, sig, host)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
