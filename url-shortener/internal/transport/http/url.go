package httpserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"urlshortener/internal/models"
	"urlshortener/internal/repository"
	random "urlshortener/internal/utils"

	"github.com/labstack/echo/v4"
)

type Request struct {
	URL      string `json:"url" validate:"required,url"`
	ShortURL string `json:"short_url,omitempty"`
}

type Response struct {
	Message any `json:"message"`
}

type UrlList struct {
	TotalCount uint64        `json:"total_count"`
	TotalPages uint64        `json:"total_pages"`
	Page       uint64        `json:"page"`
	Size       uint64        `json:"size"`
	Urls       []*models.URL `json:"urls"`
}

//go:generate mockery --name=URLService --output=mocks --case=underscore
type URLService interface {
	SaveURL(ctx context.Context, original_url, short_url string) error
	GetURL(ctx context.Context, short_url string) (*models.URL, error)
	GetAll(ctx context.Context) ([]*models.URL, error)
	Visit(ctx context.Context, url *models.URL, r *http.Request) error
	DeleteURL(ctx context.Context, short_url string) error
}

// SaveURL godoc
// @Summary      Save URL
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        body body Request true "URL"
// @Success      200  {object}  Response
// @Failure		 400  {object}  Response
// @Failure		 404  {object}  Response
// @Failure		 500  {object}  Response
// @Router       /url [post]
func (s server) HandleURLSave(c echo.Context) error {
	var req *Request

	if err := c.Bind(&req); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		return echo.ErrBadRequest
	}

	s.logger.Info("request body decoded", "request", req)

	if errs := validateWithTrans(req); errs != nil {
		return echo.NewHTTPError(http.StatusBadRequest, Response{errs})
	}

	short_url := req.ShortURL
	if short_url == "" {
		short_url = random.RandomString(s.cfg.AliasLength)
	}

	ctx := c.Request().Context()
	if err := s.urlService.SaveURL(ctx, req.URL, short_url); err != nil {
		if errors.Is(err, repository.ErrURLExists) {
			return echo.NewHTTPError(http.StatusBadRequest, Response{"This URL already exists"})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, Response{"Failed to add URL"})
	}

	return c.JSON(http.StatusCreated, Response{short_url})
}

// RedirectToURL godoc
// @Summary      Redirect URL
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        short_url path string true "Shortened URL"
// @Success      200  {string}  string "Found"
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /{short_url} [get]
func (s server) HandleURLRedirect(c echo.Context) error {
	short_url := c.Param("short_url")
	if strings.TrimSpace(short_url) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, Response{"Short URL cannot be empty"})
	}

	ctx := c.Request().Context()
	url, err := s.urlService.GetURL(ctx, short_url)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, Response{"URL not found"})
		}

		return echo.ErrInternalServerError
	}

	s.urlService.Visit(ctx, url, c.Request())

	return c.Redirect(http.StatusFound, url.OriginalURL)
}

// GetUrl godoc
// @Summary      Get All URLs
// @Description  get urls
// @Tags         URL
// @Accept       json
// @Produce      json
// @Success      200  {object}  UrlList
// @Failure		 400  {object}  Response
// @Router       /url/all [get]
func (s server) HandleURLGetAll(c echo.Context) error {
	ctx := c.Request().Context()
	urls, err := s.urlService.GetAll(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, Response{"URL not found"})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, Response{"Failed to get URL"})
	}

	return c.JSON(http.StatusOK, urls)
}

// GetUrl godoc
// @Summary      Get URL
// @Description  get url string by short url
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        short_url path string true "Short of the URL"
// @Success      200  {object}  models.URL
// @Failure		 400  {object}  Response
// @Router       /url/{short_url} [get]
func (s server) HandleURLGet(c echo.Context) error {
	short_url := c.Param("short_url")
	if strings.TrimSpace(short_url) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, Response{"Short URL cannot be empty"})
	}

	ctx := c.Request().Context()
	url, err := s.urlService.GetURL(ctx, short_url)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, Response{"URL not found"})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, Response{"Failed to get URL"})
	}

	return c.JSON(http.StatusOK, url)
}

// DeleteUrl godoc
// @Summary      Delete URL
// @Description  delete url by short url
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        short_url path string true "Short of the URL"
// @Success      200  {object}  models.URL
// @Failure		 400  {object}  Response
// @Router       /url/{short_url} [delete]
func (s server) HandleURLDelete(c echo.Context) error {
	short_url := c.Param("short_url")
	if strings.TrimSpace(short_url) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, Response{"Short URL cannot be empty"})
	}

	ctx := c.Request().Context()
	err := s.urlService.DeleteURL(ctx, short_url)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, Response{"URL not found"})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, Response{"Url could not be deleted"})
	}

	return c.NoContent(http.StatusNoContent)
}
