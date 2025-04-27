package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"tiny/internal/models"
	"tiny/internal/storage"
	random "tiny/internal/utils"

	"github.com/labstack/echo/v4"
)

//go:generate mockery --name=URLStorage --output=mocks --case=underscore
type URLStorage interface {
	SaveURL(urlToSave, alias string) error
	GetURL(alias string) (*models.URL, error)
	FetchAll(ctx context.Context) ([]*models.URL, error)
	DeleteURL(alias string) error
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type UrlList struct {
	TotalCount uint64        `json:"total_count"`
	TotalPages uint64        `json:"total_pages"`
	Page       uint64        `json:"page"`
	Size       uint64        `json:"size"`
	Urls       []*models.URL `json:"urls"`
}

// SaveURL godoc
// @Summary      Save URL
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        body body Request true "URL"
// @Success      200  {object}  Response
// @Failure		 400  {object}  response.ErrorResponse
// @Failure		 404  {object}  response.ErrorResponse
// @Failure		 500  {object}  response.ErrorResponse
// @Router       /url [post]
func (h Handler) SaveURL(c echo.Context) error {
	var req *Request

	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		return echo.ErrBadRequest
	}

	h.logger.Info("request body decoded", "request", req)

	if errs := validateWithTrans(req); errs != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errs)
	}

	alias := req.Alias
	if alias == "" {
		alias = random.RandomString(h.cfg.AliasLength)
	}

	if err := h.storage.SaveURL(req.URL, alias); err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			return echo.NewHTTPError(http.StatusBadRequest, "This URL already exists")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add URL")
	}

	return c.JSON(http.StatusCreated, echo.Map{"mesage": alias})
}

// RedirectToURL godoc
// @Summary      Redirect URL
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        alias path string true "Alias of the URL"
// @Success      200  {string}  string "Found"
// @Failure		 400  {object}  response.ErrorResponse
// @Failure		 404  {object}  response.ErrorResponse
// @Failure		 500  {object}  response.ErrorResponse
// @Router       /{alias} [get]
func (h Handler) RedirectToURL(c echo.Context) error {
	alias := c.Param("alias")
	if strings.TrimSpace(alias) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Alias cannot be empty")
	}

	url, err := h.storage.GetURL(alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return echo.ErrNotFound
		}

		return echo.ErrInternalServerError
	}
	return c.Redirect(http.StatusFound, url.URL)
}

// GetUrl godoc
// @Summary      Get All URLs
// @Description  get urls
// @Tags         URL
// @Accept       json
// @Produce      json
// @Success      200  {object}  UrlList
// @Failure		 400  {object}  response.ErrorResponse
// @Router       /url/all [get]
func (h Handler) GetAllUrls(c echo.Context) error {

	urls, err := h.storage.FetchAll(context.Background())
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return echo.ErrNotFound
		}

		h.logger.Error("error getting url from db", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get url")
	}

	return c.JSON(http.StatusOK, urls)
}

// GetUrl godoc
// @Summary      Get URL
// @Description  get url string by alias
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        alias path string true "Alias of the URL"
// @Success      200  {object}  models.URL
// @Failure		 400  {object}  response.ErrorResponse
// @Router       /url/{alias} [get]
func (h Handler) GetURL(c echo.Context) error {
	alias := c.Param("alias")
	if strings.TrimSpace(alias) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Alias cannot be empty")
	}

	url, err := h.storage.GetURL(alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return echo.ErrNotFound
		}

		h.logger.Error("error getting url from db", "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "Failed to get url")
	}

	return c.JSON(http.StatusOK, url)
}

// DeleteUrl godoc
// @Summary      Delete URL
// @Description  delete url by alias
// @Tags         URL
// @Accept       json
// @Produce      json
// @Param        alias path string true "Alias of the URL"
// @Success      200  {object}  models.URL
// @Failure		 400  {object}  response.ErrorResponse
// @Router       /url/{alias} [delete]
func (h Handler) DeleteURL(c echo.Context) error {
	alias := c.Param("alias")
	if strings.TrimSpace(alias) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Alias cannot be empty")
	}

	err := h.storage.DeleteURL(alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return echo.ErrNotFound
		}

		return echo.NewHTTPError(http.StatusBadRequest, "Url could not be deleted")
	}

	return c.NoContent(http.StatusNoContent)
}
