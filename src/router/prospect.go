package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

func NewProspectHandler(db data.Database) ProspectHandler {
	return ProspectHandler{db}
}

type ProspectHandler struct {
	db data.Database
}

func (ph ProspectHandler) createProspectHandler(w http.ResponseWriter, r *http.Request) {
	var p models.Prospect
	_ = json.NewDecoder(r.Body).Decode(&p)

	slog.Info(ru.GetRequestId(r), "prospect", p.Fmt())
	prospect, err := ph.db.CreateProspect(&p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prospect)
}

func (ph ProspectHandler) readProspectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	prospect, err := ph.db.ReadProspect(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prospect)
}

func (ph ProspectHandler) updateProspectHandler(w http.ResponseWriter, r *http.Request) {
	var p models.Prospect
	_ = json.NewDecoder(r.Body).Decode(&p)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), "prospect", p.Fmt())
	p.Id = int64(id)
	prospect, err := ph.db.UpdateProspect(&p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prospect)
}

func (ph ProspectHandler) deleteProspectHandler(w http.ResponseWriter, r *http.Request) {
	var p models.Prospect
	_ = json.NewDecoder(r.Body).Decode(&p)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), "prospect", p.Fmt())
	p.Id = int64(id)
	prospect, err := ph.db.DeleteProspect(&p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prospect)
}

func (ph ProspectHandler) readAllProspectHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	prospects, err := ph.db.ReadProspects(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prospects)
}
