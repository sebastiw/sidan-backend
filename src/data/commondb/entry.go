package commondb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateEntry(entry *models.Entry) (*models.Entry, error) {
	result := d.DB.Create(&entry)
	if result.Error != nil {
		return nil, result.Error
	}
	return entry, nil
}

func (d *CommonDatabase) ReadEntry(id int64) (*models.Entry, error) {
	var entry models.Entry

	result := d.DB.Preload("SideKicks").First(&entry, models.Entry{Id: id})

	if result.Error != nil {
		return nil, result.Error
	}
	return &entry, nil
}

func (d *CommonDatabase) ReadEntries(take int, skip int) ([]models.Entry, error) {
	var entries []models.Entry

	result := d.DB.Limit(take).Offset(skip).Find(&entries).Preload("SideKicks")

	if result.Error != nil {
		return nil, result.Error
	}
	return entries, nil
}

func (d *CommonDatabase) UpdateEntry(entry *models.Entry) (*models.Entry, error) {
	result := d.DB.Save(&entry)

	if result.Error != nil {
		return nil, result.Error
	}
	return entry, nil
}

func (d *CommonDatabase) DeleteEntry(entry *models.Entry) (*models.Entry, error) {
	result := d.DB.Delete(&entry)

	if result.Error != nil {
		return nil, result.Error
	}
	return entry, nil
}
