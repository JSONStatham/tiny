package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"urlshortener/internal/config"
	"urlshortener/internal/services/url"
	"urlshortener/internal/services/userinfo"

	"github.com/labstack/echo/v4"
)

type server struct {
	cfg        *config.Config
	logger     *slog.Logger
	urlService *url.URLService
	userInfo   *userinfo.Service
	srv        *http.Server
}

func New(cfg *config.Config, l *slog.Logger, us *url.URLService) server {
	e := echo.New()
	s := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      e,
		ReadTimeout:  cfg.HttpServer.TimeOut,
		WriteTimeout: cfg.HttpServer.TimeOut,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	server := server{
		cfg:        cfg,
		logger:     l,
		urlService: us,
		srv:        s,
	}

	server.registerRoutes(e)

	return server
}

func (s server) Run() error {
	return s.srv.ListenAndServe()
}

func (s server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
