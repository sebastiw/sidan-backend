package main

import (
	"fmt"
	"net/http"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"

	l "github.com/sebastiw/sidan-backend/src/logger"
	c "github.com/sebastiw/sidan-backend/src/config"
	d "github.com/sebastiw/sidan-backend/src/database"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	l.SetupLogging()

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
	slog.Info("Starting backend service", slog.String("address", address))

	mux := r.Mux(db, configuration.Server.StaticPath, configuration.Mail, configuration.OAuth2)

	http.ListenAndServe(address, mux)
}
