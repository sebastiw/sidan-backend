package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	. "github.com/sebastiw/sidan-backend/src/database"
	. "github.com/sebastiw/sidan-backend/src/database/models"
)

func NewMemberOperation(db *sql.DB) MemberOperation {
	return MemberOperation{db}
}

type MemberOperation struct {
	db *sql.DB
}

func (o MemberOperation) Create(m Member) Member {
	q := `
INSERT INTO cl2007_members
SET
 number=(SELECT t.number+1 FROM cl2007_members t ORDER BY number DESC LIMIT 1),
 name=?, email=?, im=?, phone=?, adress=?, adressurl=?, title=?, history=?,
 picture=?, password=?, isvalid=?, password_classic=?,
 password_classic_resetstring=?, password_resetstring=?
`

	res, err := o.db.Exec(q,
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

func (o MemberOperation) Read(id int) Member {
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

	err := o.db.QueryRow(q, id).Scan(
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

func (o MemberOperation) ReadAll() []Member {
	l := make([]Member, 0)

	q := `
SELECT
 id, number, name, email, im, phone, adress, adressurl, title, history, picture,
 password, isvalid, password_classic,
 password_classic_resetstring, password_resetstring
FROM cl2007_members
ORDER BY number DESC, id DESC
`

	rows, err := o.db.Query(q)
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

func (o MemberOperation) Update(m Member) Member {
	q := `
UPDATE cl2007_members
SET
 name=?, email=?, im=?, phone=?, adress=?, adressurl=?, title=?, history=?,
 picture=?, password=?, isvalid=?, password_classic=?,
 password_classic_resetstring=?, password_resetstring=?
WHERE id=? AND number=?
LIMIT 1
`

	if 0 == m.Id || nil == m.Number {
		// Raise error
		ErrorCheck(errors.New("id and/or Number is not set"))
	}

	res, err := o.db.Exec(q,
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

	if i == 0 {
		log.Println(fmt.Sprintf("0 rows affected (id: %d, number: %s)", m.Id, *m.Number))
	}

	return m
}

func (o MemberOperation) Delete(m Member) Member {
	if 0 == m.Id || nil == m.Number {
		// Raise error
		ErrorCheck(errors.New("id and/or Number is not set"))
	}

	_, err := o.db.Exec("DELETE FROM cl2007_members WHERE id=? AND number=?", m.Id, m.Number)
	ErrorCheck(err)

	return m
}
