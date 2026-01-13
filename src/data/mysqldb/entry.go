package mysqldb

import (
       "github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) CreateEntry(entry *models.Entry) (*models.Entry, error) {
	return d.CommonDB.CreateEntry(entry)
}

func (d *MySQLDatabase) ReadEntry(id int64) (*models.Entry, error) {
	return d.CommonDB.ReadEntry(id)
}

func (d *MySQLDatabase) ReadEntries(take int, skip int) ([]models.Entry, error) {
	return d.CommonDB.ReadEntries(take, skip)
}

func (d *MySQLDatabase) UpdateEntry(entry *models.Entry) (*models.Entry, error) {
	return d.CommonDB.UpdateEntry(entry)
}

func (d *MySQLDatabase) DeleteEntry(entry *models.Entry) (*models.Entry, error) {
	return d.CommonDB.DeleteEntry(entry)
}

func (d *MySQLDatabase) LikeEntry(entryId int64, sig string) error {
	return d.CommonDB.LikeEntry(entryId, sig)
}
