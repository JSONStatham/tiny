package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"tiny/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Handler struct {
	cfg     *config.Config
	logger  *slog.Logger
	storage URLStorage
}

func New(cfg *config.Config, l *slog.Logger, st URLStorage) Handler {
	return Handler{
		cfg:     cfg,
		logger:  l,
		storage: st,
	}
}

// @title           		   Tiny URL API
// @version        			   1.0
// @description     		   This is a sample server celler server.
// @host      				   localhost:8080
// @BasePath  				   /
// @securityDefinitions.basic  BasicAuth
func (h Handler) RegisterRoutes(e *echo.Echo) {
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				h.logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				h.logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
	}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/up", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	e.POST("/url", h.SaveURL)
	e.GET("/url/:alias", h.GetURL)
	e.GET("/:alias", h.RedirectToURL)
	e.DELETE("/url/:alias", h.DeleteURL)
	e.GET("/url/all", h.GetAllUrls)
}
