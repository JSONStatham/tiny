package main

import (
	"analytics/internal/config"
	"analytics/internal/repository/postgres"
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfg := config.MustLoad()

	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		}),
	)

	ctx := context.Background()

	repository, err := postgres.New(ctx, cfg)
	if err != nil {
		logger.Error("failed to init repository", "err", err)
		os.Exit(1)
	}
	defer repository.Close()

	logger.Info("repository created")
}
