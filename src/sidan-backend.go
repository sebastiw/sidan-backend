package main

import (
	"fmt"
	"net/http"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sebastiw/sidan-backend/src/logger"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	logger.SetupLogging()

	config.Init()

	db, err := data.NewDatabase()
	if err != nil {
		slog.Error(err.Error())
	}

	address := fmt.Sprintf(":%v", config.GetServer().Port)
	slog.Info("Starting backend service", slog.String("address", address))

	mux := r.Mux(db)

	http.ListenAndServe(address, mux)
}
