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

func NewMemberHandler(db data.Database) MemberHandler {
	return MemberHandler{db}
}

type MemberHandler struct {
	db data.Database
}

// @Summary Create member
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member body models.Member true "Member data"
// @Success 200 {object} models.Member
// @Router /db/members [post]
func (mh MemberHandler) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	slog.Info(ru.GetRequestId(r), m.Fmt())
	member, err := mh.db.CreateMember(&m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// @Summary Get member by ID
// @Description Returns full member data with read:member scope, limited data otherwise
// @Tags members
// @Produce json
// @Param id path int true "Member ID"
// @Success 200 {object} models.Member
// @Router /db/members/{id} [get]
func (mh MemberHandler) readMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	member, err := mh.db.ReadMember(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) readMemberUnauthedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	member, err := mh.db.ReadMember(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	b, err := json.Marshal(member)
	if err != nil {
		slog.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	liteMemberData := models.MemberLite{}

	err = json.Unmarshal(b, &liteMemberData)
	if err != nil {
		slog.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(liteMemberData)
}

// @Summary Update member
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Member ID"
// @Param member body models.Member true "Member data"
// @Success 200 {object} models.Member
// @Router /db/members/{id} [put]
func (mh MemberHandler) updateMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), m.Fmt())
	m.Id = int64(id)
	member, err := mh.db.UpdateMember(&m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// @Summary Delete member
// @Tags members
// @Security BearerAuth
// @Param id path int true "Member ID"
// @Success 200 {object} models.Member
// @Router /db/members/{id} [delete]
func (mh MemberHandler) deleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), m.Fmt())
	m.Id = int64(id)
	member, err := mh.db.DeleteMember(&m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// @Summary List all members
// @Description Returns full member data with read:member scope, limited data otherwise
// @Tags members
// @Produce json
// @Param onlyValid query bool false "Return only valid members" default(false)
// @Success 200 {array} models.Member
// @Router /db/members [get]
func (mh MemberHandler) readAllMemberHandler(w http.ResponseWriter, r *http.Request) {
	onlyValid := MakeDefaultBool(r, "onlyValid", "false")
	members, err := mh.db.ReadMembers(onlyValid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (mh MemberHandler) readAllMemberUnauthedHandler(w http.ResponseWriter, r *http.Request) {
	onlyValid := MakeDefaultBool(r, "onlyValid", "false")
	members, err := mh.db.ReadMembers(onlyValid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	b, err := json.Marshal(members)
	if err != nil {
		slog.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	liteMemberData := []models.MemberLite{}

	err = json.Unmarshal(b, &liteMemberData)
	if err != nil {
		slog.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(liteMemberData)
}
