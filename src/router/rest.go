package router

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/sebastiw/sidan-backend/src/database"
)

type RestHandler struct {
	version int
	db *sql.DB
	user_id string
}

type sideKick struct {
  Number string `json:""`
}

type Entry struct {
  Id int64 `json:""`
  Sig string `json:""`
  Place string `json:""`
  Lat *float64 `json:""`
  Lon *float64 `json:""`
  Msg string `json:""`
  Date string `json:""`
  Time string `json:""`
  Status int64 `json:""`
  Likes int64 `json:""`
  Enheter int64 `json:""`
  Secret bool `json:""`
  PersonalSecret bool `json:""`
  SideKicks []sideKick `json:[]`
}


func (rh RestHandler) Fmt() string {
	return fmt.Sprintf("Rest{version: %d}", rh.version)
}

func (rh RestHandler) getEntries(w http.ResponseWriter, r *http.Request) {
	take := MakeDefaultInt(r, "Take", "20")
	skip := MakeDefaultInt(r, "Skip", "0")

	entries := make([]Entry, 0)

	

	q := `CALL ReadEntries(?, ?, ?)`

	rows, err := rh.db.Query(q, skip, take, rh.user_id)
	ErrorCheck(err)
	defer rows.Close()

	for rows.Next() {
		var e = Entry{}
		err := rows.Scan(
			&e.Id,
			&e.Date,
			&e.Time,
			&e.Msg,
			&e.Status,
			&e.Sig,
			&e.Place,
			&e.Enheter,
			&e.Lat,
			&e.Lon,
			&e.Likes,
			&e.Secret,
			&e.PersonalSecret)
		switch {
		case err == sql.ErrNoRows:
		case err != nil:
			ErrorCheck(err)
		default:
		}

		kumpaner := make([]sideKick, 0)
		q2 := `SELECT number FROM cl2003_msgs_kumpaner WHERE id=?`
		rows2, err2 := rh.db.Query(q2, e.Id)
		ErrorCheck(err2)
		defer rows2.Close()

		for rows2.Next() {
			var s = sideKick{}
			err2 := rows2.Scan(&s.Number)
			switch {
			case err2 == sql.ErrNoRows:
			case err2 != nil:
				ErrorCheck(err2)
			default:
			}
			kumpaner = append(kumpaner, s)
		}
		e.SideKicks = kumpaner
		entries = append(entries, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
