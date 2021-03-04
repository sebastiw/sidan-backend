package database

import (
	"fmt"
	"log"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func connect_string(user string, pw string, host string, port int, schema string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", user, pw, host, port, schema)
}

func Connect(user string, pw string, host string, port int, schema string) *sql.DB {
	connectString := connect_string(user, pw, host, port, schema)
	log.Printf("Connecting to %s", host)

	db, err := sql.Open("mysql", connectString)
	ErrorCheck(err)

	return db
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