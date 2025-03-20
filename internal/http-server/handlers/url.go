package urlhandler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"tiny/internal/config"
	"tiny/internal/models"
	"tiny/internal/storage"
	random "tiny/internal/utils"
	resp "tiny/internal/utils/api"
	"tiny/internal/utils/logger"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
)

//go:generate mockery --name=URLStorage --output=mocks --case=underscore
type URLStorage interface {
	SaveURL(urlToSave, alias string) error
	GetURL(alias string) (*models.URL, error)
	DeleteURL(alias string) error
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Message string `json:"message"`
	Alias   string `json:"alias,omitempty"`
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
func SaveURL(cfg config.Config, log *slog.Logger, st URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.SaveURL"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// If empty body requested
			log.Error("request body is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("The request body is empty"))
			return
		}

		if err != nil {
			log.Error("failed to decode request body", logger.Err(err))
			render.JSON(w, r, resp.Error("Failed do decode body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		en := en.New()
		uni = ut.New(en, en)

		validate = validator.New()
		trans, ok := uni.GetTranslator("en")
		if ok {
			en_translations.RegisterDefaultTranslations(validate, trans)
		}

		if err := validate.Struct(req); err != nil {
			errs := err.(validator.ValidationErrors)
			log.Error("invalid request", logger.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Invalid URL", errs.Translate(trans)))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.RandomString(cfg.AliasLength)
		}

		if err = st.SaveURL(req.URL, alias); err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("This URL already exists"))
				return
			}

			log.Error("failed to add url", logger.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to add URL"))
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Alias: alias,
		})
	}
}

// TODO: work porper with responses and error responses
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
func RedirectToURL(cfg config.Config, log *slog.Logger, st URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.RedirectToURL"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias, ok := getAliasFromRequest(w, r, log)
		if !ok {
			return
		}

		url, err := st.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url could not be found", logger.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("Url could not be found"))
				return
			}

			log.Info("url could not be found", logger.Err(err))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("Url could not be found"))
			return
		}

		http.Redirect(w, r, url.URL, http.StatusFound)
		return
	}
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
func GetURL(cfg config.Config, log *slog.Logger, st URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.GetURL"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias, ok := getAliasFromRequest(w, r, log)
		if !ok {
			return
		}

		url, err := st.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url could not be found", logger.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("Url could not be found"))
				return
			}

			log.Info("error getting url from db", logger.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Unexpected error"))
			return
		}

		render.JSON(w, r, render.M{"data": url})
	}
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
func DeleteURL(cfg config.Config, log *slog.Logger, st URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.DeleteURL"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias, ok := getAliasFromRequest(w, r, log)
		if !ok {
			return
		}

		err := st.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url could not be found", logger.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("Url cannot be found"))
				return
			}

			log.Info("url could not be deleted", logger.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Url could not be deleted"))
			return
		}

		render.NoContent(w, r)
	}
}

func getAliasFromRequest(w http.ResponseWriter, r *http.Request, log *slog.Logger) (string, bool) {
	alias := chi.URLParam(r, "alias")
	if alias == "" {
		log.Info("alias is not specified in the request")

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, resp.Error("ALias is required in the request"))
		return "", false
	}

	return alias, true
}
