package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tiny/internal/config"
	urlhandler "tiny/internal/http-server/handlers"
	"tiny/internal/storage/postgres"
	"tiny/internal/utils/logger"
	pretty_logger "tiny/pkg/logger"

	_ "tiny/docs"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// @title           		   Tiny URL API
// @version        			   1.0
// @description     		   This is a sample server celler server.
// @host      				   localhost:8080
// @BasePath  				   /
// @securityDefinitions.basic  BasicAuth
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg)

	log.Info("Init storage")

	storage, err := postgres.New(cfg, log)
	if err != nil {
		log.Error("failed to init storage", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer storage.DB.Close()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(logger.New(log))
	router.Use(middleware.URLFormat)

	router.Get("/up", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	basicAuth := middleware.BasicAuth("tiny", map[string]string{
		cfg.HttpServer.User: cfg.HttpServer.Password,
	})

	router.With(basicAuth).Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/url", func(r chi.Router) {
		r.Use(basicAuth)

		r.Post("/", urlhandler.SaveURL(cfg, log, storage))
		r.Get("/{alias}", urlhandler.GetURL(cfg, log, storage))
		r.Delete("/{alias}", urlhandler.DeleteURL(cfg, log, storage))
		// router.Get("/url/all", urlhandler.GetAllURL(cfg, log, storage))
	})

	router.Get("/{alias}", urlhandler.RedirectToURL(cfg, log, storage))

	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.TimeOut,
		WriteTimeout: cfg.HttpServer.TimeOut,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	// Gracefull shtdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server started", slog.String("address", cfg.HttpServer.Address))

		if err := server.ListenAndServe(); err != nil {
			log.Error("failed to start server", logger.Err(err))
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", logger.Err(err))
		return
	}

	log.Info("server stoped gracefully")
}

func setupLogger(cfg config.Config) *slog.Logger {
	var logger *slog.Logger

	switch cfg.Env {
	case envLocal:
		logger = setupPrettyLogger()
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return logger
}

func setupPrettyLogger() *slog.Logger {
	opts := pretty_logger.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
