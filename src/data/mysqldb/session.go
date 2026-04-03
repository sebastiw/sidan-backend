package mysqldb

import "github.com/sebastiw/sidan-backend/src/models"

func (d *MySQLDatabase) CreateSession(session *models.Session) error {
	return d.CommonDB.CreateSession(session)
}

func (d *MySQLDatabase) GetSession(token string) (*models.Session, error) {
	return d.CommonDB.GetSession(token)
}

func (d *MySQLDatabase) DeleteSession(token string) error {
	return d.CommonDB.DeleteSession(token)
}

func (d *MySQLDatabase) CleanupExpiredSessions() error {
	return d.CommonDB.CleanupExpiredSessions()
}
