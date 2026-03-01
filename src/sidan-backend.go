package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/fdroid"
	"github.com/sebastiw/sidan-backend/src/logger"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	logger.SetupLogging()

	config.Init()

	// Ensure F-Droid repo directories exist
	fdroidCfg := config.GetFDroid()
	os.MkdirAll(fdroidCfg.RepoPath+"/icons", 0755)

	// Always generate the signed index on startup so the F-Droid client can
	// reach the repo even before any APKs have been uploaded.
	if err := fdroid.GenerateIndex(fdroidCfg.RepoPath, fdroidCfg); err != nil {
		slog.Warn("F-Droid index generation on startup failed (is jarsigner on PATH and keystore present?)", "error", err)
	}

	db, err := data.NewDatabase()
	if err != nil {
		slog.Error(err.Error())
	}

	address := fmt.Sprintf(":%v", config.GetServer().Port)
	slog.Info("Starting backend service", slog.String("address", address))

	mux := r.Mux(db)

	http.ListenAndServe(address, mux)
}
