package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	. "github.com/sebastiw/sidan-backend/src/database"
	. "github.com/sebastiw/sidan-backend/src/database/models"
)

func Create(db *sql.DB, m Member) Member {
	q := `
INSERT INTO cl2007_members
SET
 number=(SELECT t.number+1 FROM cl2007_members t ORDER BY number DESC LIMIT 1),
 name=?, email=?, im=?, phone=?, adress=?, adressurl=?, title=?, history=?,
 picture=?, password=?, isvalid=?, password_classic=?,
 password_classic_resetstring=?, password_resetstring=?
`

	res, err := db.Exec(q,
		m.Name,
		m.Email,
		m.Im,
		m.Phone,
		m.Adress,
		m.Adressurl,
		m.Title,
		m.History,
		m.Picture,
		m.Password,
		m.Isvalid,
		m.Password_classic,
		m.Password_classic_resetstring,
		m.Password_resetstring)

	ErrorCheck(err)

	id, e := res.LastInsertId()
	ErrorCheck(e)

	m.Id = id
	return m
}

func Read(db *sql.DB, id int) Member {
	var m = Member{}

	q := `
SELECT
 id, number, name, email, im, phone, adress, adressurl, title, history, picture,
 password, isvalid, password_classic,
 password_classic_resetstring, password_resetstring
FROM cl2007_members
WHERE id=?
ORDER BY number DESC,id DESC
LIMIT 1
`

	err := db.QueryRow(q, id).Scan(
		&m.Id,
		&m.Number,
		&m.Name,
		&m.Email,
		&m.Im,
		&m.Phone,
		&m.Adress,
		&m.Adressurl,
		&m.Title,
		&m.History,
		&m.Picture,
		&m.Password,
		&m.Isvalid,
		&m.Password_classic,
		&m.Password_classic_resetstring,
		&m.Password_resetstring)

	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		ErrorCheck(err)
	default:
	}
	return m
}

func ReadAll(db *sql.DB) []Member {
	l := make([]Member, 0)

	q := `
SELECT
 id, number, name, email, im, phone, adress, adressurl, title, history, picture,
 password, isvalid, password_classic,
 password_classic_resetstring, password_resetstring
FROM cl2007_members
ORDER BY number DESC, id DESC
`

	rows, err := db.Query(q)
	ErrorCheck(err)
	defer rows.Close()

	for rows.Next() {
		var m = Member{}
		err := rows.Scan(
			&m.Id,
			&m.Number,
			&m.Name,
			&m.Email,
			&m.Im,
			&m.Phone,
			&m.Adress,
			&m.Adressurl,
			&m.Title,
			&m.History,
			&m.Picture,
			&m.Password,
			&m.Isvalid,
			&m.Password_classic,
			&m.Password_classic_resetstring,
			&m.Password_resetstring)
		switch {
		case err == sql.ErrNoRows:
		case err != nil:
			ErrorCheck(err)
		default:
		}
		l = append(l, m)
	}

	return l
}

func Update(db *sql.DB, m Member) Member {
	q := `
UPDATE cl2007_members
SET
 name=?, email=?, im=?, phone=?, adress=?, adressurl=?, title=?, history=?,
 picture=?, password=?, isvalid=?, password_classic=?,
 password_classic_resetstring=?, password_resetstring=?
WHERE id=? AND number=?
LIMIT 1
`

	if(0 == m.Id || nil == m.Number) {
		// Raise error
		ErrorCheck(errors.New("Id and/or Number is not set"))
	}

	res, err := db.Exec(q,
		m.Name,
		m.Email,
		m.Im,
		m.Phone,
		m.Adress,
		m.Adressurl,
		m.Title,
		m.History,
		m.Picture,
		m.Password,
		m.Isvalid,
		m.Password_classic,
		m.Password_classic_resetstring,
		m.Password_resetstring,
		m.Id,
		m.Number)
	ErrorCheck(err)

	i, err := res.RowsAffected()
	ErrorCheck(err)

	if(i == 0) {
		log.Println(fmt.Sprintf("0 rows affected (id: %d, number: %s)", m.Id, *m.Number))
	}

	return m
}

func Delete(db *sql.DB, m Member) Member {
	if(0 == m.Id || nil == m.Number) {
		// Raise error
		ErrorCheck(errors.New("Id and/or Number is not set"))
	}

	_, err := db.Exec("DELETE FROM cl2007_members WHERE id=? AND number=?", m.Id, m.Number)
	ErrorCheck(err)

	return m
}
