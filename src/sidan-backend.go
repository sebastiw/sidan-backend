package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	c "github.com/sebastiw/sidan-backend/src/config"
	d "github.com/sebastiw/sidan-backend/src/database"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	// to change the flags on the default logger; Llongfile or Lshortfile
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var configuration c.Configuration

	c.ReadConfig(&configuration)

	db := d.Connect(
		configuration.Database.User,
		configuration.Database.Password,
		configuration.Database.Host,
		configuration.Database.Port,
		configuration.Database.Schema)
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	d.Ping(db)
	d.Configure(db)
	d.ConfigureSession(db)

	address := fmt.Sprintf(":%v", configuration.Server.Port)
	log.Printf("Starting backend service at %v", address)

	mux := r.Mux(db, configuration.Server.StaticPath, configuration.Mail, configuration.OAuth2)

	log.Fatal(http.ListenAndServe(address, mux))
}
