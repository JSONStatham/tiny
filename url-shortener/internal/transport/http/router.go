package httpserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title           		   Tiny URL API
// @version        			   1.0
// @description     		   This is a sample server celler server.
// @host      				   localhost:8080
// @BasePath  				   /
// @securityDefinitions.basic  BasicAuth
func (s server) registerRoutes(e *echo.Echo) {
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				s.logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				s.logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
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
		return c.NoContent(http.StatusOK)
	})
	e.POST("/url", s.HandleURLSave)
	e.GET("/:short_url", s.HandleURLRedirect)
	e.GET("/url/:short_url", s.HandleURLGet)
	e.GET("/url/all", s.HandleURLGetAll)
	e.DELETE("/url/:short_url", s.HandleURLDelete)
}
