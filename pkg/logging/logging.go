package logging

import (
	"log/slog"
	"os"
	"tiny/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func SetupLogger(cfg *config.Config) *slog.Logger {
	var logger *slog.Logger

	switch cfg.Env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: false,
				Level:     slog.LevelDebug,
			}),
		)
	case envProd:
		logger = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelInfo,
			}),
		)
	}

	return logger
}
