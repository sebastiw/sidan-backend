package main

import (
	"fmt"
	"log"
	"net/http"

	"database/sql"
	"time"
	_ "github.com/go-sql-driver/mysql"

	c "github.com/sebastiw/sidan-backend/src/config"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	var configuration c.Configurations

	c.ReadConfig(&configuration)

	// sql.connect(connect_config)
	connectString := fmt.Sprintf(
		"%s:%s@tcp(%s:%v)/%s",
		configuration.Database.User,
		configuration.Database.Password,
		configuration.Database.Host,
		configuration.Database.Port,
		configuration.Database.Schema)
	log.Printf("Connect St%s", connectString)
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	rows, err := db.Query(`SHOW TABLES;`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	log.Printf("ROES: %s", rows)

	address := fmt.Sprintf(":%v", configuration.Server.Port)
	log.Printf("Starting backend service at %v", address)

	log.Fatal(http.ListenAndServe(address, r.Mux()))
}
