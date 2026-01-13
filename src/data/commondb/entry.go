package commondb

import (
	"time"
	
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateEntry(entry *models.Entry) (*models.Entry, error) {
	// Set current date and time if not provided
	now := time.Now()
	if entry.Date == "" {
		entry.Date = now.Format("2006-01-02")
	}
	if entry.Time == "" {
		entry.Time = now.Format("15:04:05")
	}
	// Set DateTime field (required for database)
	if entry.DateTime.IsZero() {
		entry.DateTime = now
	}

	result := d.DB.Create(&entry)
	if result.Error != nil {
		return nil, result.Error
	}
	return entry, nil
}

func (d *CommonDatabase) ReadEntry(id int64) (*models.Entry, error) {
	var entry models.Entry

	// Load entry with related data
	result := d.DB.Preload("SideKicks").
		Preload("LikeRecords").
		Preload("Permissions").
		First(&entry, models.Entry{Id: id})

	if result.Error != nil {
		return nil, result.Error
	}
	
	// Compute virtual fields
	entry.Likes = int64(len(entry.LikeRecords))
	entry.Secret = len(entry.Permissions) > 0
	entry.PersonalSecret = false
	for _, perm := range entry.Permissions {
		if perm.UserId != 0 {
			entry.PersonalSecret = true
			break
		}
	}
	
	return &entry, nil
}

func (d *CommonDatabase) ReadEntries(take int, skip int) ([]models.Entry, error) {
	var entries []models.Entry

	result := d.DB.Order("id DESC").
		Limit(take).
		Offset(skip).
		Preload("SideKicks").
		Preload("LikeRecords").
		Preload("Permissions").
		Find(&entries)

	if result.Error != nil {
		return nil, result.Error
	}
	
	// Compute virtual fields for each entry
	for i := range entries {
		entries[i].Likes = int64(len(entries[i].LikeRecords))
		entries[i].Secret = len(entries[i].Permissions) > 0
		entries[i].PersonalSecret = false
		for _, perm := range entries[i].Permissions {
			if perm.UserId != 0 {
				entries[i].PersonalSecret = true
				break
			}
		}
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

func (d *CommonDatabase) LikeEntry(entryId int64, sig string, host string) error {
	now := time.Now()
	like := map[string]interface{}{
		"date": now.Format("2006-01-02"),
		"time": now.Format("15:04:05"),
		"id":   entryId,
		"sig":  sig,
		"host": host,
	}
	result := d.DB.Table("2003_likes").Create(like)
	return result.Error
}
