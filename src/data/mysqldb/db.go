package mysqldb

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/driver/mysql"

	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data/commondb"
	"github.com/sebastiw/sidan-backend/src/models"
)

type MySQLDatabase struct {
	DB       *gorm.DB
	CommonDB *commondb.CommonDatabase
}

func dsnWithoutSchema(user string, pw string, host string, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=UTC",
		user, pw, host, port)
}
func dsnWithSchema(user string, pw string, host string, port int, schema string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC&multiStatements=true",
		user, pw, host, port, schema)
}

func createSchema(db *gorm.DB, schema string) error {
	createDatabaseCommand := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;", schema)
	result := db.Exec(createDatabaseCommand)
	return result.Error
}

func Configure(db *gorm.DB) {
	//db.SetConnMaxLifetime(time.Minute * 3)
	//db.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(10)
}

func ConfigureSession(db *gorm.DB) {
	q := `SET SESSION sql_mode = 'STRICT_TRANS_TABLES,STRICT_ALL_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,TRADITIONAL,NO_ENGINE_SUBSTITUTION'`
	result := db.Exec(q)
	if result.Error != nil {
		slog.Warn("unable to configure session")
	}
}

func NewMySQLDatabase() (*MySQLDatabase, error) {
	slog.Info("using database mysql")
	username := config.GetDatabase().User
	host := config.GetDatabase().Host
	port := config.GetDatabase().Port
	schema := config.GetDatabase().Schema

	slog.Info("db params",
		slog.String("username", username),
		slog.String("host", host),
		slog.Int("port", port),
		slog.String("schema", schema))

	// dsn := dsnWithoutSchema(
	//      username, config.GetDatabase().Password, host, port)
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// // db, err := sql.Open("mysql", connStr)
	// if err != nil {
	//      slog.Error("unable to open database")
	//      return nil, err
	// }

	// // create the database if it does not exist
	// err = createSchema(db, schema)
	// if err != nil {
	//      slog.Error("unable to create database")
	//      return nil, err
	// }

	dsn := dsnWithSchema(
		username, config.GetDatabase().Password, host, port, schema)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// db, err = sql.Open("mysql", connStr)
	if err != nil {
		slog.Error("unable to open database with schema")
		return nil, err
	}

	// Configure(db)
	ConfigureSession(db)

	commonDb := commondb.NewCommonDatabase(db, "MySQL")

	mysqlDb := MySQLDatabase{
		DB:       db,
		CommonDB: commonDb,
	}
	return &mysqlDb, nil
}

func (d *MySQLDatabase) IsEmpty() (bool, error) {
	return d.CommonDB.IsEmpty()
}

// Auth operations (Phase 1) - delegated to CommonDB
func (d *MySQLDatabase) CreateAuthState(state *models.AuthState) error {
	return d.CommonDB.CreateAuthState(state)
}

func (d *MySQLDatabase) GetAuthState(id string) (*models.AuthState, error) {
	return d.CommonDB.GetAuthState(id)
}

func (d *MySQLDatabase) DeleteAuthState(id string) error {
	return d.CommonDB.DeleteAuthState(id)
}

func (d *MySQLDatabase) CleanupExpiredAuthStates() error {
	return d.CommonDB.CleanupExpiredAuthStates()
}

func (d *MySQLDatabase) CreateAuthToken(token *models.AuthToken) error {
	return d.CommonDB.CreateAuthToken(token)
}

func (d *MySQLDatabase) GetAuthToken(memberID int64, provider string) (*models.AuthToken, error) {
	return d.CommonDB.GetAuthToken(memberID, provider)
}

func (d *MySQLDatabase) GetAuthTokenByMemberID(memberID int64) ([]models.AuthToken, error) {
	return d.CommonDB.GetAuthTokenByMemberID(memberID)
}

func (d *MySQLDatabase) UpdateAuthToken(token *models.AuthToken) error {
	return d.CommonDB.UpdateAuthToken(token)
}

func (d *MySQLDatabase) DeleteAuthToken(memberID int64, provider string) error {
	return d.CommonDB.DeleteAuthToken(memberID, provider)
}

func (d *MySQLDatabase) DeleteAllAuthTokens(memberID int64) error {
	return d.CommonDB.DeleteAllAuthTokens(memberID)
}

func (d *MySQLDatabase) CreateAuthProviderLink(link *models.AuthProviderLink) error {
	return d.CommonDB.CreateAuthProviderLink(link)
}

func (d *MySQLDatabase) GetAuthProviderLink(provider, providerUserID string) (*models.AuthProviderLink, error) {
	return d.CommonDB.GetAuthProviderLink(provider, providerUserID)
}

func (d *MySQLDatabase) GetAuthProviderLinksByMemberID(memberID int64) ([]models.AuthProviderLink, error) {
	return d.CommonDB.GetAuthProviderLinksByMemberID(memberID)
}

func (d *MySQLDatabase) GetMemberByProviderEmail(provider, email string) (*models.Member, error) {
	return d.CommonDB.GetMemberByProviderEmail(provider, email)
}

func (d *MySQLDatabase) DeleteAuthProviderLink(provider, providerUserID string) error {
	return d.CommonDB.DeleteAuthProviderLink(provider, providerUserID)
}

func (d *MySQLDatabase) CreateAuthSession(session *models.AuthSession) error {
	return d.CommonDB.CreateAuthSession(session)
}

func (d *MySQLDatabase) GetAuthSession(id string) (*models.AuthSession, error) {
	return d.CommonDB.GetAuthSession(id)
}

func (d *MySQLDatabase) UpdateAuthSession(session *models.AuthSession) error {
	return d.CommonDB.UpdateAuthSession(session)
}

func (d *MySQLDatabase) DeleteAuthSession(id string) error {
	return d.CommonDB.DeleteAuthSession(id)
}

func (d *MySQLDatabase) DeleteAllAuthSessions(memberID int64) error {
	return d.CommonDB.DeleteAllAuthSessions(memberID)
}

func (d *MySQLDatabase) CleanupExpiredAuthSessions() error {
	return d.CommonDB.CleanupExpiredAuthSessions()
}

func (d *MySQLDatabase) TouchAuthSession(id string) error {
	return d.CommonDB.TouchAuthSession(id)
}
