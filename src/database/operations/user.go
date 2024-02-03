package database

import (
	"database/sql"
	"strings"

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
 "#" AS type, number, email
FROM cl2007_members
WHERE email in (` + qms + `) AND isvalid = true

ORDER BY number DESC
LIMIT 1
`

	emailString := strings.Join(emails, ",")

	err := o.db.QueryRow(q, emailString).Scan(
		&u.Type,
		&u.Number,
		&u.Email,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}
