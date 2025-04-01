package logger

import (
	"log/slog"
	"os"
)

func SetupLogging() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		AddSource: true,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}
