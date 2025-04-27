package handlers_test

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
	"tiny/internal/config"
	"tiny/internal/handlers"
	"tiny/internal/http-server/handlers/mocks"
	"tiny/internal/models"
	"tiny/internal/storage"
	response "tiny/internal/utils/api"
	slogdiscard "tiny/internal/utils/logger/handlers"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url, alias  string
		respStatus  int
		respError   string
		mockError   error
		requestBody []byte
	}{
		{
			name:       "Success with alias",
			url:        "http://example.com",
			alias:      "test_alias",
			respStatus: 201,
		},
		{
			name:       "Success with empty alias",
			url:        "http://example.com",
			alias:      "",
			respStatus: 201,
		},
		{
			name:       "Empty url",
			url:        "",
			alias:      "test_alias",
			respStatus: 400,
			respError:  "Invalid URL",
		},
		{
			name:       "Invalid URL",
			url:        "some invalid URL",
			alias:      "some_alias",
			respStatus: 400,
			respError:  "Invalid URL",
		},
		{
			name:       "SaveURL Error",
			alias:      "test_alias",
			url:        "https://google.com",
			respError:  "Failed to add URL",
			respStatus: 400,
			mockError:  errors.New("unexpected error"),
		},
		{
			name:        "Empty request body",
			respError:   "Invalid URL",
			respStatus:  400,
			requestBody: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewURLStorage(t)

			if tt.respError == "" || tt.mockError != nil {
				mockStorage.On("SaveURL", tt.url, tt.alias).
					Return(tt.mockError).
					Once()
			}

			// Create handler
			handler := handlers.SaveURL(config.Config{}, slogdiscard.NewDiscardLogger(), mockStorage)

			// Format request body
			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tt.url, tt.alias)
			if len(tt.requestBody) != 0 {
				input = string(tt.requestBody)
			}

			req, err := http.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tt.respStatus)

			body := rr.Body.String()

			var resp urlhandler.Response
			var errResponse response.ErrorResponse

			// Unmarshal response and error response
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.NoError(t, json.Unmarshal([]byte(body), &errResponse))

			require.Equal(t, tt.respError, errResponse.Message)
		})
	}
}

func TestGetURL(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		alias      string
		mockReturn *models.URL
		mockError  error
		respStatus int
		respBody   string
	}{
		{
			name:  "Success - URL found",
			alias: "test_alias",
			mockReturn: &models.URL{
				ID:        1,
				URL:       "http://example.com",
				Alias:     "test_alias",
				CreatedAt: now,
			},
			respStatus: http.StatusOK,
		},
		{
			name:       "Not Found - URL does not exist",
			alias:      "missing_alias",
			mockReturn: nil,
			mockError:  storage.ErrURLNotFound,
			respStatus: http.StatusNotFound,
			respBody:   `{"message":"Url could not be found"}`,
		},
		{
			name:       "Internal Error - Unexpected error",
			alias:      "error_alias",
			mockReturn: nil,
			mockError:  errors.New("database error"),
			respStatus: http.StatusInternalServerError,
			respBody:   `{"message":"Unexpected error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewURLStorage(t)

			mockStorage.On("GetURL", tt.alias).
				Return(tt.mockReturn, tt.mockError).
				Once()

			// Create handler
			handler := urlhandler.GetURL(config.Config{}, slogdiscard.NewDiscardLogger(), mockStorage)

			url := fmt.Sprintf("/url/%s", tt.alias)
			req, err := http.NewRequest(http.MethodGet, url, nil)

			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("alias", tt.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.respStatus, rr.Code)

			if tt.mockReturn != nil {
				var resp struct {
					Data models.URL `json:"data"`
				}
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

				assert.Equal(t, tt.mockReturn.ID, resp.Data.ID)
				assert.Equal(t, tt.mockReturn.URL, resp.Data.URL)
				assert.Equal(t, tt.mockReturn.Alias, resp.Data.Alias)

				// Validate CreatedAt separately
				assert.WithinDuration(t, tt.mockReturn.CreatedAt, resp.Data.CreatedAt, time.Second)
			} else {
				assert.JSONEq(t, tt.respBody, rr.Body.String())
			}
		})
	}

}

func TestDeleteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alias      string
		mockError  error
		respStatus int
		respError  string
	}{
		{
			name:       "Success - URL deleted",
			alias:      "test_alias",
			respStatus: http.StatusNoContent,
		},
		{
			name:       "Not Found - URL does not exist",
			alias:      "missing_alias",
			mockError:  storage.ErrURLNotFound,
			respStatus: http.StatusNotFound,
			respError:  `{"message": "Url cannot be found"}`,
		},
		{
			name:       "Internal Error - Unexpected error",
			alias:      "error_alias",
			mockError:  errors.New("database error"),
			respStatus: http.StatusInternalServerError,
			respError:  `{"message": "Url could not be deleted"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewURLStorage(t)

			mockStorage.On("DeleteURL", tt.alias).
				Return(tt.mockError).
				Once()

			// Create handler
			handler := urlhandler.DeleteURL(config.Config{}, slogdiscard.NewDiscardLogger(), mockStorage)

			url := fmt.Sprintf("/url/%s", tt.alias)
			req, err := http.NewRequest(http.MethodDelete, url, nil)

			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("alias", tt.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.respStatus, rr.Code)

			if tt.mockError != nil {
				assert.JSONEq(t, tt.respError, rr.Body.String())
			} else {
				assert.Empty(t, rr.Body.String())
			}
		})
	}

}
