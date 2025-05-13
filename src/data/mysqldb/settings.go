package mysqldb

import (
       "github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) GetSettingsById(settingsId int64) (*models.Settings, error) {
       return d.CommonDB.GetSettingsById(settingsId)
}
