package main

import (
	"fmt"
	"net/http"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/sebastiw/sidan-backend/docs"

	"github.com/sebastiw/sidan-backend/src/logger"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	r "github.com/sebastiw/sidan-backend/src/router"
)

// @title Sidan API
// @version 3.0
// @description Backend for sidan. Authentication: Visit /swagger-auth to login with Google and get your JWT token, then paste it in the Authorize dialog.
// @license.name MIT
// @license.url http://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Visit /swagger-auth to obtain a token via Google OAuth2.

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
