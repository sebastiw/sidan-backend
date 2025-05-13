package commondb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) GetSettingsById(settingsId int64) (*models.Settings, error) {
	var settings models.Settings
	result := d.DB.First(&settings, "id = ?", settingsId)

	if result.Error != nil {
		return nil,result.Error
	}
	return &settings,nil
}
