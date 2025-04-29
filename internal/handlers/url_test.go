package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"tiny/internal/config"
	"tiny/internal/handlers"
	"tiny/internal/handlers/mocks"
	"tiny/internal/models"
	"tiny/internal/storage"
	slogdiscard "tiny/internal/utils/logger/handlers"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		url, alias     string
		expectedCode   int
		expectedErrMsg string
		mockError      error
		requestBody    []byte
		wantErr        bool
	}{
		{
			name:         "Success with alias",
			url:          "http://example.com",
			alias:        "test_alias",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Success with empty alias",
			url:          "http://example.com",
			alias:        "",
			expectedCode: http.StatusCreated,
		},
		{
			name:           "Empty url",
			url:            "",
			alias:          "test_alias",
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Invalid URL",
			wantErr:        true,
		},
		{
			name:           "Invalid URL",
			url:            "some invalid URL",
			alias:          "some_alias",
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Invalid URL",
			wantErr:        true,
		},
		{
			name:           "SaveURL Error",
			alias:          "test_alias",
			url:            "https://google.com",
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
			mockStorage := mocks.NewURLStorage(t)

			if tt.expectedErrMsg == "" || tt.mockError != nil {
				mockStorage.On("SaveURL", tt.url, tt.alias).
					Return(tt.mockError).
					Once()
			}

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tt.url, tt.alias)
			if len(tt.requestBody) != 0 {
				input = string(tt.requestBody)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			h := handlers.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockStorage)
			err := h.SaveURL(c)

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

func TestGetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		alias        string
		mockReturn   *models.URL
		mockError    error
		expectedCode int
		expectedResp string
		wantErr      bool
	}{
		{
			name:  "Success - URL found",
			alias: "test_alias",
			mockReturn: &models.URL{
				ID:        1,
				URL:       "http://example.com",
				Alias:     "test_alias",
				CreatedAt: time.Now(),
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not Found - URL does not exist",
			alias:        "missing_alias",
			mockReturn:   nil,
			mockError:    storage.ErrURLNotFound,
			expectedCode: http.StatusNotFound,
			expectedResp: `{"message":"Url could not be found"}`,
			wantErr:      true,
		},
		{
			name:         "Internal Error - Unexpected error",
			alias:        "error_alias",
			mockReturn:   nil,
			mockError:    errors.New("database error"),
			expectedCode: http.StatusInternalServerError,
			expectedResp: `{"message":"Unexpected error"}`,
			wantErr:      true,
		},
		{
			name:         "Empty alias",
			alias:        "",
			mockReturn:   nil,
			expectedCode: http.StatusBadRequest,
			expectedResp: `{"message":"Alias cannot be empty"}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewURLStorage(t)

			if tt.alias != "" {
				mockStorage.On("GetURL", tt.alias).
					Return(tt.mockReturn, tt.mockError).
					Once()
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("alias")
			c.SetParamValues(tt.alias)

			h := handlers.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockStorage)

			err := h.GetURL(c)
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
				assert.Equal(t, tt.mockReturn.URL, resp["url"])
				assert.Equal(t, tt.mockReturn.Alias, resp["alias"])

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
		alias          string
		mockError      error
		expectedCode   int
		expectedErrMsg string
	}{
		{
			name:         "Success URL deleted",
			alias:        "test_alias",
			expectedCode: http.StatusNoContent,
		},
		{
			name:           "URL does not exist",
			alias:          "missing_alias",
			mockError:      storage.ErrURLNotFound,
			expectedCode:   http.StatusNotFound,
			expectedErrMsg: "Not Found",
		},
		{
			name:           "Unexpected error from db",
			alias:          "error_alias",
			mockError:      errors.New("database error"),
			expectedCode:   http.StatusInternalServerError,
			expectedErrMsg: "Url could not be deleted",
		},
		{
			name:           "Empty path alias",
			alias:          "",
			expectedCode:   http.StatusBadRequest,
			expectedErrMsg: "Alias cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewURLStorage(t)

			if tt.alias != "" {
				mockStorage.On("DeleteURL", tt.alias).
					Return(tt.mockError).
					Once()
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("alias")
			c.SetParamValues(tt.alias)

			h := handlers.New(&config.Config{}, slogdiscard.NewDiscardLogger(), mockStorage)

			err := h.DeleteURL(c)

			if err != nil {
				he, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, he.Code)
				assert.Equal(t, tt.expectedErrMsg, he.Message)
			} else {
				assert.Empty(t, rec.Body.String())
				assert.Equal(t, tt.expectedCode, rec.Code)
			}
		})
	}
}
