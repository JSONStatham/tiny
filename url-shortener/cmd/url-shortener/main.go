package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urlshortener/internal/config"
	"urlshortener/internal/repository/postgres"
	"urlshortener/internal/repository/redis"
	"urlshortener/internal/services/url"
	httpserver "urlshortener/internal/transport/http"
	"urlshortener/internal/transport/kafka"
	"urlshortener/pkg/logging"

	_ "urlshortener/api"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfg := config.MustLoad()

	logger := logging.SetupLogger(cfg.Env)

	logger.Info("Init repository")
	repository, err := postgres.New(cfg)
	if err != nil {
		logger.Error("failed to init repository", "error", err)
		os.Exit(1)
	}
	defer repository.DB.Close()

	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		logger.Error("failed to init kafka producer", "error", err)
		os.Exit(1)
	}
	logger.Info("kafka producer created", "details", producer.Details())
	defer producer.Close()

	cache := redis.New(cfg)
	logger.Info("redis connected", "addr", cfg.Cache.Addr)
	defer cache.Close()

	urlService := url.New(cfg, logger, repository, producer, cache, nil)

	httpServer := httpserver.New(cfg, logger, urlService)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server started", slog.String("address", cfg.HttpServer.Address))

		if err := httpServer.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Warn("server shutdown")
				return
			}

			logger.Error("failed to start server", "error", err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := httpServer.Stop(ctx); err != nil {
		logger.Error("failed to stop server", "error", err)
		return
	}

	logger.Info("server stoped gracefully")
}
