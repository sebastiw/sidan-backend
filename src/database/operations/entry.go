package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	. "github.com/sebastiw/sidan-backend/src/database"
	. "github.com/sebastiw/sidan-backend/src/database/models"
)

func NewEntryOperation(db *sql.DB) EntryOperation {
	return EntryOperation{db}
}

type EntryOperation struct {
	db *sql.DB
}

func (o EntryOperation) Create(e Entry) Entry {
	q := `
INSERT INTO cl2003_msgs
SET
  date=CURRENT_DATE, TIME=CURRENT_TIME,
  msg=?, status=?, cl=?, sig=?, email=?, place=?, ip=?, host=?,
  olsug=?, enheter=?, lat=?, lon=?, report=?
`

	res, err := o.db.Exec(q,
		e.Msg,
		e.Status,
		e.Cl,
		e.Sig,
		e.Email,
		e.Place,
		e.Ip,
		e.Host,
		e.Olsug,
		e.Enheter,
		e.Lat,
		e.Lon,
		e.Report)

	ErrorCheck(err)

	id, err := res.LastInsertId()
	ErrorCheck(err)

	e.Id = id
	return e
}

func (o EntryOperation) Read(id int) Entry {
	q := `
SELECT
 m.id, m.date, m.time, m.msg, m.status, m.cl, m.sig, m.email, m.place,
 m.ip, m.host, m.olsug, m.enheter, m.lat, m.lon, m.report,
 count(l.id) as likes,
 p.user_id IS NOT NULL AS secret,
 p.user_id IS NOT NULL && "0" NOT IN (p.user_id) AS personalsecret,
 GROUP_CONCAT(DISTINCT IFNULL(k.number, '') ORDER BY k.number SEPARATOR ",") as kumpaner
FROM cl2003_msgs AS m
LEFT JOIN 2003_likes AS l ON m.id = l.id
LEFT JOIN cl2003_permissions AS p ON m.id = p.id
LEFT JOIN cl2003_msgs_kumpaner AS k ON m.id = k.id
WHERE m.id=?
GROUP BY
 m.id, m.date, m.time, m.msg, m.status, m.cl, m.sig, m.email, m.place,
 m.ip, m.host, m.olsug, m.enheter, m.lat, m.lon, m.report,
 secret, personalsecret
`

	var kumpaner = *new(string)
	var e = Entry{SideKicks: []SideKick{}}
	err := o.db.QueryRow(q, id).Scan(
		&e.Id,
		&e.Date,
		&e.Time,
		&e.Msg,
		&e.Status,
		&e.Cl,
		&e.Sig,
		&e.Email,
		&e.Place,
		&e.Ip,
		&e.Host,
		&e.Olsug,
		&e.Enheter,
		&e.Lat,
		&e.Lon,
		&e.Report,
		&e.Likes,
		&e.Secret,
		&e.PersonalSecret,
		&kumpaner)

	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		ErrorCheck(err)
	default:
	}
	if(kumpaner != "") {
		sidekicks := make([]SideKick, 0)
		for _,n := range strings.Split(kumpaner, ",") {
			sidekicks = append(sidekicks, SideKick{Number: "#"+n})
		}
		e.SideKicks = sidekicks
	}

	return e
}

func (o EntryOperation) ReadAll(take int64, skip int64) []Entry {
	l := make([]Entry, 0)

	q := `
SELECT
 m.id, m.date, m.time, m.msg, m.status, m.cl, m.sig, m.email, m.place,
 m.ip, m.host, m.olsug, m.enheter, m.lat, m.lon, m.report,
 count(l.id) as likes,
 p.user_id IS NOT NULL AS secret,
 p.user_id IS NOT NULL && "0" NOT IN (p.user_id) AS personalsecret,
 GROUP_CONCAT(DISTINCT IFNULL(k.number, '') ORDER BY k.number SEPARATOR ",") as kumpaner
FROM cl2003_msgs AS m
LEFT JOIN 2003_likes AS l ON m.id = l.id
LEFT JOIN cl2003_permissions AS p ON m.id = p.id
LEFT JOIN cl2003_msgs_kumpaner AS k ON m.id = k.id
GROUP BY
 m.id, m.date, m.time, m.msg, m.status, m.cl, m.sig, m.email, m.place,
 m.ip, m.host, m.olsug, m.enheter, m.lat, m.lon, m.report,
 secret, personalsecret
ORDER BY id DESC
LIMIT ?, ?
`

	rows, err := o.db.Query(q, skip, take)
	ErrorCheck(err)
	defer rows.Close()

	for rows.Next() {
		var kumpaner = *new(string)
		var e = Entry{SideKicks: []SideKick{}}
		err := rows.Scan(
			&e.Id,
			&e.Date,
			&e.Time,
			&e.Msg,
			&e.Status,
			&e.Cl,
			&e.Sig,
			&e.Email,
			&e.Place,
			&e.Ip,
			&e.Host,
			&e.Olsug,
			&e.Enheter,
			&e.Lat,
			&e.Lon,
			&e.Report,
			&e.Likes,
			&e.Secret,
			&e.PersonalSecret,
			&kumpaner)
		switch {
		case err == sql.ErrNoRows:
		case err != nil:
			ErrorCheck(err)
		default:
		}
		if(kumpaner != "") {
			sidekicks := make([]SideKick, 0)
			for _,n := range strings.Split(kumpaner, ",") {
				sidekicks = append(sidekicks, SideKick{Number: "#"+n})
			}
			e.SideKicks = sidekicks
		}
		l = append(l, e)
	}

	return l
}

func (o EntryOperation) Update(e Entry) Entry {
	q := `
UPDATE cl2003_msgs
SET
  msg=?, status=?, cl=?, sig=?, email=?, place=?, ip=?, host=?,
  olsug=?, enheter=?, lat=?, lon=?, report=?
WHERE id=?
LIMIT 1
`

	if(0 == e.Id) {
		// Raise error
		ErrorCheck(errors.New("Id is not set"))
	}

	res, err := o.db.Exec(q,
		e.Msg,
		e.Status,
		e.Cl,
		e.Sig,
		e.Email,
		e.Place,
		e.Ip,
		e.Host,
		e.Olsug,
		e.Enheter,
		e.Lat,
		e.Lon,
		e.Report,
		e.Id)
	ErrorCheck(err)

	i, err := res.RowsAffected()
	ErrorCheck(err)

	if(i == 0) {
		log.Println(fmt.Sprintf("0 rows affected (id: %d)", e.Id))
	}

	return e
}

func (o EntryOperation) Delete(e Entry) Entry {
	if(0 == e.Id) {
		// Raise error
		ErrorCheck(errors.New("Id is not set"))
	}

	_, err := o.db.Exec("DELETE FROM cl2003_msgs WHERE id=?", e.Id)
	ErrorCheck(err)

	return e
}
