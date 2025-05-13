package commondb

import (
	"log/slog"

	"gorm.io/gorm"
	// "github.com/sebastiw/sidan-backend/src/models"
)

type CommonDatabase struct {
	DB     *gorm.DB
	Flavor string
}

func NewCommonDatabase(db *gorm.DB, flavor string) *CommonDatabase {
	return &CommonDatabase{
		DB:     db,
		Flavor: flavor,
	}
}

func (d *CommonDatabase) Migrate() error {
	return d.DB.AutoMigrate()
}

func (d *CommonDatabase) IsEmpty() (bool, error) {
	settings, err := d.GetSettingsById(1)
	if err != nil {
		slog.Warn("failed to check if database is empty")
		return false, err
	}

	return settings == nil, nil
}

func (d *CommonDatabase) BeginTransaction() *gorm.DB {
	return d.DB.Begin()
}
func (d *CommonDatabase) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}
func (d *CommonDatabase) RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}
