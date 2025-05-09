package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	. "github.com/sebastiw/sidan-backend/src/database"
	"github.com/sebastiw/sidan-backend/src/models"
)

func NewEntryOperation(db *sql.DB) EntryOperation {
	return EntryOperation{db}
}

// swagger:model

type EntryOperation struct {
	db *sql.DB
}

func (o EntryOperation) Create(e models.Entry) models.Entry {
	q := `
INSERT INTO cl2003_msgs
SET
  date=CURRENT_DATE, time=CURRENT_TIME,
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

	q = `INSERT INTO cl2003_msgs_kumpaner(id, number) VALUES (?, ?)`
	for _, n := range e.SideKicks {
		if n.Number[0] != '#' {
			panic("SideKick not starting with '#': " + n.Number)
		}
		number, err := strconv.Atoi(n.Number[1:])
		ErrorCheck(err)
		_, err2 := o.db.Exec(q, id, number)
		ErrorCheck(err2)
	}

	return e
}

// swagger:operation

func (o EntryOperation) Read(id int) models.Entry {
	q := `
SELECT
 m.id, m.date, m.time, m.datetime, m.msg, m.status, m.cl, m.sig, m.email, m.place,
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
 m.id
`

	var kumpaner = *new(string)
	var e = models.Entry{SideKicks: []models.SideKick{}}
	err := o.db.QueryRow(q, id).Scan(
		&e.Id,
		&e.Date,
		&e.Time,
		&e.DateTime,
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
	if kumpaner != "" {
		sidekicks := make([]models.SideKick, 0)
		for _, n := range strings.Split(kumpaner, ",") {
			sidekicks = append(sidekicks, models.SideKick{Number: "#" + n})
		}
		e.SideKicks = sidekicks
	}

	return e
}

func (o EntryOperation) ReadAll(take int64, skip int64) []models.Entry {
	l := make([]models.Entry, 0)

	q := `
SELECT
 m.id, m.date, m.time, m.datetime, m.msg, m.status, m.cl, m.sig, m.email, m.place,
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
 m.id
ORDER BY id DESC
LIMIT ?, ?
`

	rows, err := o.db.Query(q, skip, take)
	ErrorCheck(err)
	defer rows.Close()

	for rows.Next() {
		var kumpaner = *new(string)
		var e = models.Entry{SideKicks: []models.SideKick{}}
		err := rows.Scan(
			&e.Id,
			&e.Date,
			&e.Time,
			&e.DateTime,
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
		if kumpaner != "" {
			sidekicks := make([]models.SideKick, 0)
			for _, n := range strings.Split(kumpaner, ",") {
				sidekicks = append(sidekicks, models.SideKick{Number: "#" + n})
			}
			e.SideKicks = sidekicks
		}
		l = append(l, e)
	}

	return l
}

func (o EntryOperation) Update(e models.Entry) models.Entry {
	q := `
UPDATE cl2003_msgs
SET
  msg=?, status=?, cl=?, sig=?, email=?, place=?, ip=?, host=?,
  olsug=?, enheter=?, lat=?, lon=?, report=?
WHERE id=?
LIMIT 1
`

	if 0 == e.Id {
		// Raise error
		ErrorCheck(errors.New("id is not set"))
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

	if i == 0 {
		slog.Warn(fmt.Sprintf("0 rows affected (id: %d)", e.Id))
	}

	return e
}

func (o EntryOperation) Delete(e models.Entry) models.Entry {
	if 0 == e.Id {
		// Raise error
		ErrorCheck(errors.New("id is not set"))
	}

	_, err := o.db.Exec("DELETE FROM cl2003_msgs WHERE id=?", e.Id)
	ErrorCheck(err)

	return e
}
