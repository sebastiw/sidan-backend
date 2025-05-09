package main

import (
	"fmt"
	"net/http"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sebastiw/sidan-backend/src/logger"
	"github.com/sebastiw/sidan-backend/src/config"
	d "github.com/sebastiw/sidan-backend/src/database"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	logger.SetupLogging()

	config.Init()

	db := d.Connect(
		config.GetDatabase().User,
		config.GetDatabase().Password,
		config.GetDatabase().Host,
		config.GetDatabase().Port,
		config.GetDatabase().Schema)
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	d.Ping(db)
	d.Configure(db)
	d.ConfigureSession(db)

	address := fmt.Sprintf(":%v", config.GetServer().Port)
	slog.Info("Starting backend service", slog.String("address", address))

	mux := r.Mux(db)

	http.ListenAndServe(address, mux)
}
