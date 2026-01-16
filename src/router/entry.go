package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"fmt"

	"github.com/gorilla/mux"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/data/mysqldb"
	"github.com/sebastiw/sidan-backend/src/filter"
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

	// Get advanced filter query parameter (requires use:advanced_filter scope)
	queryString := r.URL.Query().Get("q")
	sortString := r.URL.Query().Get("sort")

	// Get viewer member from auth context (nil if unauthenticated)
	member := GetMemberFromContext(r)
	var viewerMemberID *int64
	if member != nil {
		viewerMemberID = &member.Number
	}

	// Get user scopes
	scopes := auth.GetScopes(r)

	// Check if using advanced filtering - require scope
	if queryString != "" || sortString != "" {
		hasFilterScope := false
		if scopes != nil {
			for _, scope := range scopes {
				if scope == auth.UseAdvancedFilterScope {
					hasFilterScope = true
					break
				}
			}
		}
		if !hasFilterScope {
			http.Error(w, `{"error":"advanced filtering requires use:advanced_filter scope"}`, http.StatusForbidden)
			return
		}
	}

	var entries []models.Entry
	var err error

	if queryString != "" || sortString != "" {
		// Use advanced filtering with ACL and masking
		// Get the underlying GORM DB from data layer
		gormDB, ok := eh.db.(*mysqldb.MySQLDatabase)
		if !ok {
			http.Error(w, "database type not supported for filtering", http.StatusInternalServerError)
			return
		}

		// Start building query
		db := gormDB.GetDB().Model(&models.Entry{})

		// Apply filters and ACL constraints
		db, err = filter.QueryWithFiltersAndACL(
			db,
			&filter.EntrySchema,
			queryString,
			sortString,
			scopes,
			viewerMemberID,
			"entry",
		)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid filter: %v"}`, err), http.StatusBadRequest)
			return
		}

		// Apply conditional masking at database level
		db = filter.ApplyConditionalMasking(db, "entry", viewerMemberID)

		// Apply pagination
		db = db.Limit(take).Offset(skip)

		// Execute query
		result := db.Find(&entries)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("database error: %v", result.Error), http.StatusInternalServerError)
			return
		}
	} else {
		// Legacy path: no advanced filtering
		entries, err = eh.db.ReadEntries(take, skip)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		// Apply message filtering to all entries (legacy method)
		FilterEntriesMessages(entries, viewerMemberID)
	}

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
