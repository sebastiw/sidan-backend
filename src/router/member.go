package router

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	model "github.com/sebastiw/sidan-backend/src/database/models"
	d "github.com/sebastiw/sidan-backend/src/database/operations"
)

func NewMemberHandler(db *sql.DB) MemberHandler {
	return MemberHandler{d.NewMemberOperation(db)}
}

type MemberHandler struct {
	op d.MemberOperation
}

func (mh MemberHandler) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	log.Println(getRequestId(r), m.Fmt())
	member := mh.op.Create(m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) readMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	member := mh.op.Read(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) updateMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Println(getRequestId(r), m.Fmt())
	m.Id = int64(id)
	member := mh.op.Update(m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) deleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member
	_ = json.NewDecoder(r.Body).Decode(&m)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Println(getRequestId(r), m.Fmt())
	m.Id = int64(id)
	member := mh.op.Delete(m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (mh MemberHandler) readAllMemberHandler(w http.ResponseWriter, r *http.Request) {
	onlyValid := MakeDefaultBool(r, "onlyValid", "false")
	members := mh.op.ReadAll(onlyValid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
