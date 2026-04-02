package mysqldb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) CreateProspect(prospect *models.Prospect) (*models.Prospect, error) {
	return d.CommonDB.CreateProspect(prospect)
}

func (d *MySQLDatabase) ReadProspect(id int64) (*models.Prospect, error) {
	return d.CommonDB.ReadProspect(id)
}

func (d *MySQLDatabase) ReadProspects(status string) ([]models.Prospect, error) {
	return d.CommonDB.ReadProspects(status)
}

func (d *MySQLDatabase) UpdateProspect(prospect *models.Prospect) (*models.Prospect, error) {
	return d.CommonDB.UpdateProspect(prospect)
}

func (d *MySQLDatabase) DeleteProspect(prospect *models.Prospect) (*models.Prospect, error) {
	return d.CommonDB.DeleteProspect(prospect)
}
