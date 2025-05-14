package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"urlshortener/internal/config"
	"urlshortener/internal/models"
	"urlshortener/internal/repository"
	httpserver "urlshortener/internal/transport/http"
	"urlshortener/internal/transport/http/mocks"
	slogdiscard "urlshortener/internal/utils/logger/handlers"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		url, shortUrl  string
		expectedCode   int
		expectedErrMsg string
		mockError      error
		requestBody    []byte
		wantErr        bool
	}{
		{
			name:         "Success with alias",
			url:          "http://example.com",
			shortUrl:     "test_alias",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Success with empty alias",
			url:          "http://example.com",
			shortUrl:     "",
			expectedCode: http.StatusCreated,
		},
		{
			name:           "Empty url",
			url:            "",
			shortUrl:       "test_alias",
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Invalid URL",
			wantErr:        true,
		},
		{
			name:           "Invalid URL",
			url:            "some invalid URL",
			shortUrl:       "some_alias",
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Invalid URL",
			wantErr:        true,
		},
		{
			name:           "SaveURL Error",
			shortUrl:       "test_alias",
			url:            "http://example.com",
			expectedErrMsg: "Failed to add URL",
			expectedCode:   http.StatusInternalServerError,
			mockError:      errors.New("unexpected error"),
			wantErr:        true,
		},
		{
			name:           "Empty request body",
			expectedErrMsg: "Request.url:url is a required field",
			expectedCode:   http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewURLService(t)

			if tt.expectedErrMsg == "" || tt.mockError != nil {
				mockSvc.On("SaveURL", context.Background(), tt.url, tt.shortUrl).
					Return(tt.mockError).
					Once()
			}

			input := fmt.Sprintf(`{"url": "%s", "short_url": "%s"}`, tt.url, tt.shortUrl)
			if len(tt.requestBody) != 0 {
				input = string(tt.requestBody)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			s := httpserver.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockSvc)
			err := s.HandleURLSave(c)

			if tt.wantErr {
				var he *echo.HTTPError
				errors.As(err, &he)
				assert.Equal(t, tt.expectedCode, he.Code)
				return
			}

			if assert.NoError(t, err) {
				var resp echo.Map
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedCode, rec.Code)
			}
		})
	}
}

func TestURLRedirect(t *testing.T) {
	cases := []struct {
		name           string
		shortUrl       string
		expectedUrl    string
		mockReturn     *models.URL
		mockError      error
		expectedCode   int
		expectedErrMsg string
		wantErr        bool
	}{
		{
			name:        "Success",
			shortUrl:    "test_alias",
			expectedUrl: "http://example.com",
			mockReturn: &models.URL{
				ID:          1,
				OriginalURL: "http://example.com",
				ShortURL:    "test_alias",
				CreatedAt:   time.Now(),
			},
			expectedCode: http.StatusFound,
		},
		{
			name:           "Empty alias",
			shortUrl:       "",
			mockReturn:     nil,
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Short URL cannot be empty",
			wantErr:        true,
		},
		{
			name:           "URL not found",
			shortUrl:       "missing_url",
			mockReturn:     nil,
			mockError:      repository.ErrURLNotFound,
			expectedCode:   http.StatusNotFound,
			expectedErrMsg: "URL not found",
			wantErr:        true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewURLService(t)

			if tt.expectedErrMsg == "" || tt.mockError != nil {
				mockSvc.On("GetURL", mock.Anything, tt.shortUrl).
					Return(tt.mockReturn, tt.mockError).
					Once()

				if !tt.wantErr {
					mockSvc.On("Visit", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				}
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("short_url")
			c.SetParamValues(tt.shortUrl)

			s := httpserver.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockSvc)
			err := s.HandleURLRedirect(c)
			if tt.wantErr {
				respErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, respErr.Code)
				assert.Equal(t, tt.expectedErrMsg, respErr.Message.(httpserver.Response).Message)
				return
			}

			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectedUrl, rec.Header().Get("Location"))
		})
	}
}

func TestGetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		shortUrl     string
		mockReturn   *models.URL
		mockError    error
		expectedCode int
		expectedResp string
		wantErr      bool
	}{
		{
			name:     "Success - URL found",
			shortUrl: "test_alias",
			mockReturn: &models.URL{
				ID:          1,
				OriginalURL: "http://example.com",
				ShortURL:    "test_alias",
				CreatedAt:   time.Now(),
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not Found - URL does not exist",
			shortUrl:     "missing_alias",
			mockReturn:   nil,
			mockError:    repository.ErrURLNotFound,
			expectedCode: http.StatusNotFound,
			expectedResp: `{"message":"Url could not be found"}`,
			wantErr:      true,
		},
		{
			name:         "Internal Error - Unexpected error",
			shortUrl:     "error_alias",
			mockReturn:   nil,
			mockError:    errors.New("database error"),
			expectedCode: http.StatusInternalServerError,
			expectedResp: `{"message":"Unexpected error"}`,
			wantErr:      true,
		},
		{
			name:         "Empty alias",
			shortUrl:     "",
			mockReturn:   nil,
			expectedCode: http.StatusBadRequest,
			expectedResp: `{"message":"Url cannot be empty"}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewURLService(t)

			if tt.shortUrl != "" {
				mockSvc.On("GetURL", mock.Anything, tt.shortUrl).
					Return(tt.mockReturn, tt.mockError).
					Once()
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("short_url")
			c.SetParamValues(tt.shortUrl)

			s := httpserver.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockSvc)

			err := s.HandleURLGet(c)
			if tt.wantErr {
				var he *echo.HTTPError
				errors.As(err, &he)
				assert.Equal(t, tt.expectedCode, he.Code)
				return
			}

			if tt.mockReturn != nil {
				require.NoError(t, err)

				var resp echo.Map

				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedCode, rec.Code)
				id, ok := resp["id"].(float64)
				if ok {
					resp["id"] = int(id)
				}

				assert.Equal(t, tt.mockReturn.ID, resp["id"])
				assert.Equal(t, tt.mockReturn.OriginalURL, resp["original_url"])
				assert.Equal(t, tt.mockReturn.ShortURL, resp["short_url"])

				// Validate CreatedAt separately
				parsedTime, err := time.Parse(time.RFC3339Nano, resp["created_at"].(string))
				assert.NoError(t, err)
				assert.WithinDuration(t, tt.mockReturn.CreatedAt, parsedTime, time.Second)

			} else {
				assert.JSONEq(t, tt.expectedResp, rec.Body.String())
			}
		})
	}

}

func TestDeleteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		shortUrl       string
		mockError      error
		expectedCode   int
		expectedErrMsg string
	}{
		{
			name:         "Success URL deleted",
			shortUrl:     "test_alias",
			expectedCode: http.StatusNoContent,
		},
		{
			name:           "URL does not exist",
			shortUrl:       "missing_alias",
			mockError:      repository.ErrURLNotFound,
			expectedCode:   http.StatusNotFound,
			expectedErrMsg: "URL not found",
		},
		{
			name:           "Unexpected error from db",
			shortUrl:       "error_alias",
			mockError:      errors.New("database error"),
			expectedCode:   http.StatusInternalServerError,
			expectedErrMsg: "Url could not be deleted",
		},
		{
			name:           "Empty path alias",
			shortUrl:       "",
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Short url cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewURLService(t)

			if tt.shortUrl != "" {
				mockSvc.On("DeleteURL", mock.Anything, tt.shortUrl).
					Return(tt.mockError).
					Once()
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("short_url")
			c.SetParamValues(tt.shortUrl)

			s := httpserver.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockSvc)

			err := s.HandleURLDelete(c)

			if err != nil {
				sErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, sErr.Code)
				assert.Equal(t, tt.expectedErrMsg, sErr.Message.(httpserver.Response).Message)
			} else {
				assert.Empty(t, rec.Body.String())
				assert.Equal(t, tt.expectedCode, rec.Code)
			}
		})
	}
}
