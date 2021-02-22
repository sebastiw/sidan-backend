package database

import (
	"database/sql"

	. "github.com/sebastiw/sidan-backend/src/database"
	. "github.com/sebastiw/sidan-backend/src/database/models"
)

func Create(db *sql.DB, name string) Member {
	res, err := db.Exec("INSERT INTO cl2007_members (`name`) VALUES (?)", name)
	ErrorCheck(err)

	id, e := res.LastInsertId()
	ErrorCheck(e)

	var m = Member{Id: id}
	return m
}

func Read(db *sql.DB, id int) Member {
	var m = Member{}

	err := db.QueryRow("SELECT * FROM cl2007_members WHERE id = ?", id).Scan(&m.Id, &m.Number, &m.Name)
	ErrorCheck(err)

	return m
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
