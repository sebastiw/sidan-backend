package database

import (
	"database/sql"

	. "github.com/sebastiw/sidan-backend/src/database"
	. "github.com/sebastiw/sidan-backend/src/database/models"
)

func Create(db *sql.DB, m Member) Member {
	res, err := db.Exec("INSERT INTO cl2007_members (`name`, `im`) VALUES (?, ?)", m.Name, m.Im)
	ErrorCheck(err)

	id, e := res.LastInsertId()
	ErrorCheck(e)

	m.Id = id
	return m
}

func Read(db *sql.DB, id int) Member {
	var m = Member{}

	err := db.QueryRow("SELECT * FROM cl2007_members WHERE id = ?", id).Scan(
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

	rows, err := db.Query("SELECT * FROM cl2007_members")
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

func Update(db *sql.DB, name string) Member {
	res, err := db.Exec("UPDATE cl2007_members SET name = ?", name)
	ErrorCheck(err)

	id, err := res.LastInsertId()
	ErrorCheck(err)

	var m = Member{Id: id}
	return m
}

func Delete(db *sql.DB, id int64) Member {
	_, err := db.Exec("DELETE FROM cl2007_members WHERE id = ?", id)
	ErrorCheck(err)

	var m = Member{Id: id}
	return m
}
