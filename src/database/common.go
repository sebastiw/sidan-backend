package database

import (
	"fmt"
	"log"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func connectString(user string, pw string, host string, port int, schema string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", user, pw, host, port, schema)
}

func Connect(user string, pw string, host string, port int, schema string) *sql.DB {
	connectString := connectString(user, pw, host, port, schema)
	log.Printf("Connecting to %s", host)

	db, err := sql.Open("mysql", connectString)
	ErrorCheck(err)

	return db
}

func ConfigureSession(db *sql.DB) {
	q := `SET SESSION sql_mode = 'STRICT_TRANS_TABLES,STRICT_ALL_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,TRADITIONAL,NO_ENGINE_SUBSTITUTION'`
	_, err := db.Exec(q)
	ErrorCheck(err)
}

func Configure(db *sql.DB) {
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}

func Ping(db *sql.DB) {
	log.Printf("Test ping DB")
	err := db.Ping()
	ErrorCheck(err)
}

func ErrorCheck(err error) {
	if err != nil {
		panic(err.Error())
	}
}
