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

func NewArrHandler(db data.Database) ArrHandler {
	return ArrHandler{db}
}

type ArrHandler struct {
	db data.Database
}

func (ah ArrHandler) createArrHandler(w http.ResponseWriter, r *http.Request) {
	var a models.Arr
	_ = json.NewDecoder(r.Body).Decode(&a)

	slog.Info(ru.GetRequestId(r), "arr", a.Fmt())
	arr, err := ah.db.CreateArr(&a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arr)
}

func (ah ArrHandler) readArrHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	arr, err := ah.db.ReadArr(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arr)
}

func (ah ArrHandler) updateArrHandler(w http.ResponseWriter, r *http.Request) {
	var a models.Arr
	_ = json.NewDecoder(r.Body).Decode(&a)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), "arr", a.Fmt())
	a.Id = int64(id)
	arr, err := ah.db.UpdateArr(&a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arr)
}

func (ah ArrHandler) deleteArrHandler(w http.ResponseWriter, r *http.Request) {
	var a models.Arr
	_ = json.NewDecoder(r.Body).Decode(&a)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), "arr", a.Fmt())
	a.Id = int64(id)
	arr, err := ah.db.DeleteArr(&a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arr)
}

func (ah ArrHandler) readAllArrHandler(w http.ResponseWriter, r *http.Request) {
	take := MakeDefaultInt(r, "take", "20")
	skip := MakeDefaultInt(r, "skip", "0")
	arrs, err := ah.db.ReadArrs(take, skip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arrs)
}
