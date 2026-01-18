package commondb

import (
	"fmt"
	"time"

	rsql "github.com/sebastiw/go-rsql-mysql"
	"github.com/sebastiw/sidan-backend/src/models"
)

// RSQL configuration for entries filtering
var (
	// Virtual field mappings: user-facing names -> SQL expressions
	entryVirtualMap = map[string]string{
		"likes":    "COUNT(DISTINCT LikeRecords.sig, LikeRecords.host)", // Count unique signatures to avoid SideKick multiplication
		"kumpaner": "SideKicks.number",                                  // Field from cl2003_msgs_kumpaner
	}

	// Allowed keys after transformation (security allowlist)
	entryAllowedKeys = []string{
		"cl2003_msgs.datetime",
		"cl2003_msgs.msg",
		"cl2003_msgs.sig",
		"cl2003_msgs.lat",
		"cl2003_msgs.lon",
		"COUNT(DISTINCT LikeRecords.sig)", // Virtual: likes
		"SideKicks.number",                // Virtual: kumpaner
	}
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

func (d *CommonDatabase) ReadEntries(take int, skip int, rsqlFilter string) ([]models.Entry, error) {
	var entries []models.Entry

	// Start with base query
	query := d.DB.Model(&models.Entry{})

	// If RSQL filtering requested, parse and apply
	if rsqlFilter != "" {
		// Create parser with key transformer
		parser, err := rsql.NewParser(
			rsql.MySQL(),
			rsql.WithKeyTransformers(func(key string) string {
				// Map virtual fields to SQL expressions
				if sqlExpr, ok := entryVirtualMap[key]; ok {
					return sqlExpr
				}
				// Prefix regular fields with table name for JOIN clarity
				return "cl2003_msgs." + key
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("RSQL parser creation failed: %w", err)
		}

		// Parse RSQL query string to SQL
		sqlHaving, err := parser.Process(rsqlFilter, rsql.SetAllowedKeys(entryAllowedKeys))
		if err != nil {
			return nil, fmt.Errorf("RSQL parse error: %w", err)
		}

		// Apply joins and filtering
		// Using aliases that match entryVirtualMap
		query = query.
			Select("cl2003_msgs.*").
			Joins("LEFT JOIN `2003_likes` LikeRecords ON LikeRecords.id = cl2003_msgs.id").
			Joins("LEFT JOIN `cl2003_msgs_kumpaner` SideKicks ON SideKicks.id = cl2003_msgs.id").
			Group("cl2003_msgs.id").
			Having(sqlHaving)
	}

	// Execute query with ordering and pagination
	result := query.
		Order("cl2003_msgs.id DESC"). // Explicit table prefix to avoid ambiguity
		Limit(take).
		Offset(skip).
		Preload("SideKicks").
		Preload("LikeRecords").
		Preload("Permissions").
		Find(&entries)

	if result.Error != nil {
		return nil, result.Error
	}

	// Post-process: compute virtual fields from loaded relationships
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
	// Check if like already exists for this entryId and sig combination
	var count int64
	result := d.DB.Table("2003_likes").
		Where("id = ? AND sig = ?", entryId, sig).
		Count(&count)
	if result.Error != nil {
		return result.Error
	}
	if count > 0 {
		return nil // Like already exists, return success (idempotent)
	}

	now := time.Now()
	like := map[string]interface{}{
		"date": now.Format("2006-01-02"),
		"time": now.Format("15:04:05"),
		"id":   entryId,
		"sig":  sig,
		"host": host,
	}
	result = d.DB.Table("2003_likes").Create(like)
	return result.Error
}
