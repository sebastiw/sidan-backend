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

	// Regenerate the signed index on startup if the repo already has content,
	// so a fresh deploy never serves a stale JAR.
	if metas, err := fdroid.LoadAllMeta(fdroidCfg.RepoPath); err == nil && len(metas) > 0 {
		if err := fdroid.GenerateIndex(fdroidCfg.RepoPath, fdroidCfg); err != nil {
			slog.Warn("F-Droid index regeneration on startup failed", "error", err)
		}
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
