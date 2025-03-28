package database

import (
	"database/sql"
	"errors"
	"strings"
	"strconv"

	. "github.com/sebastiw/sidan-backend/src/database"
	. "github.com/sebastiw/sidan-backend/src/database/models"
)

func NewUserOperation(db *sql.DB) UserOperation {
	return UserOperation{db}
}

type UserOperation struct {
	db *sql.DB
}

func (o UserOperation) GetUserFromEmails(emails []string) (User, error) {
	var u = User{}

	qms := strings.Repeat("?,", len(emails))
	qms = qms[:len(qms)-1] // remove the trailing ","

	q := `
SELECT
 "#" AS type, number, email, password_classic
FROM cl2007_members
WHERE email in (` + qms + `) AND isvalid = true

ORDER BY number DESC
LIMIT 1
`

	y := make([]interface{}, len(emails))
	for i, v := range emails {
		y[i] = v
	}

	err := o.db.QueryRow(q, y...).Scan(
		&u.Type,
		&u.Number,
		&u.Email,
		&u.FulHaxPass,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func contains(s []byte, str byte) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (o UserOperation) GetUserFromLogin(username string, password string) (User, error) {
	var u = User{}

	q := `
SELECT
 "#" AS type, number, email, password_classic
FROM cl2007_members
WHERE number=? AND password_classic=? AND isvalid = true

ORDER BY number DESC
LIMIT 1
`
	types := []byte{'#', 'P', 'S', 'p', 's'}
	if !contains(types, username[0])  {
		return u, errors.New("Username not starting with '#', 'P', 'S'")
	}
	number, err := strconv.Atoi(username[1:])
	ErrorCheck(err)

	err2 := o.db.QueryRow(q, number, password).Scan(
		&u.Type,
		&u.Number,
		&u.Email,
		&u.FulHaxPass,
	)

	return u, err2
}
