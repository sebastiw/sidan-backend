package data

import (
	"fmt"
	"log/slog"
	"errors"

	// "gorm.io/gorm"

	"github.com/sebastiw/sidan-backend/src/models"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data/mysqldb"
)

type Database interface {
	// Migrate() error
	IsEmpty() (bool, error)
	// BeginTransaction() *gorm.DB
	// CommitTransaction(*gorm.DB) error
	// RollbackTransaction(*gorm.DB) error

	GetSettingsById(settingsId int64) (*models.Settings, error)

	GetUserFromEmails(emails []string) (*models.User, error)
	GetUserFromLogin(username string, password string) (*models.User, error)

	CreateEntry(entry *models.Entry) (*models.Entry, error)
	ReadEntry(id int64) (*models.Entry, error)
	ReadEntries(take int, skip int) ([]models.Entry, error)
	UpdateEntry(entry *models.Entry) (*models.Entry, error)
	DeleteEntry(entry *models.Entry) (*models.Entry, error)

	CreateMember(member *models.Member) (*models.Member, error)
	ReadMember(id int64) (*models.Member, error)
	ReadMembers(onlyValid bool) ([]models.Member, error)
	UpdateMember(member *models.Member) (*models.Member, error)
	DeleteMember(member *models.Member) (*models.Member, error)
}

func NewDatabase() (Database, error) {
	var database Database
	var err error

	switch config.GetDatabase().Type {
	case "mysql":
		slog.Info("creating mysql database")
		database, err = mysqldb.NewMySQLDatabase()
	default:
		msg := fmt.Sprintf("unsupported database type: '%s'. supported types are: mysql, sqlite, postgres, mssql", config.GetDatabase().Type)
		return nil, errors.New(msg)
	}

	if err != nil {
		return nil, err
	}

	// err = database.Migrate()
	// if err != nil {
	//         return nil, err
	// }

	return database, nil
}
