package mysqldb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) GetUserFromEmails(emails []string) (*models.User, error) {
	return d.CommonDB.GetUserFromEmails(emails)
}

func (d *MySQLDatabase) GetUserFromLogin(username string, password string) (*models.User, error) {
	return d.CommonDB.GetUserFromLogin(username, password)
}
