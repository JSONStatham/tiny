package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tiny/internal/config"
	"tiny/internal/handlers"

	"tiny/internal/storage/postgres"
	"tiny/pkg/logging"

	_ "tiny/docs"

	"github.com/labstack/echo/v4"
)

func main() {
	cfg := config.MustLoad()

	logger := logging.SetupLogger(cfg)

	logger.Info("Init storage")

	storage, err := postgres.New(cfg)
	if err != nil {
		logger.Error("failed to init storage", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer storage.DB.Close()

	echo := echo.New()
	h := handlers.New(cfg, logger, storage)
	h.RegisterRoutes(echo)

	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      echo,
		ReadTimeout:  cfg.HttpServer.TimeOut,
		WriteTimeout: cfg.HttpServer.TimeOut,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	// Gracefull shtdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server started", slog.String("address", cfg.HttpServer.Address))

		if err := server.ListenAndServe(); err != nil {
			logger.Error("failed to start server", "error", err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to stop server", "error", err)
		return
	}

	logger.Info("server stoped gracefully")
}
