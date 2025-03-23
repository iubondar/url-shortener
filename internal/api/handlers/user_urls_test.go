package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserUrlsHandler_RetrieveUserURLs(t *testing.T) {
	userID := uuid.New()
	baseURL := "http://127.0.0.1"
	tests := []struct {
		name     string
		method   string
		records  []storage.Record
		userID   uuid.UUID
		wantCode int
		wantOut  []UserUrlsOut
	}{
		{
			name:   "Positive test",
			method: http.MethodGet,
			records: []storage.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
				{
					ShortURL:    "456",
					OriginalURL: "http://ya.ru",
					UserID:      userID,
				},
			},
			userID:   userID,
			wantCode: http.StatusOK,
			wantOut: []UserUrlsOut{
				{
					ShortURL:    baseURL + "/123",
					OriginalURL: "http://example.com",
				},
				{
					ShortURL:    baseURL + "/456",
					OriginalURL: "http://ya.ru",
				},
			},
		},
		{
			name:   "Only with correct userID",
			method: http.MethodGet,
			records: []storage.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
				{
					ShortURL:    "456",
					OriginalURL: "http://ya.ru",
					UserID:      uuid.New(),
				},
			},
			userID:   userID,
			wantCode: http.StatusOK,
			wantOut: []UserUrlsOut{
				{
					ShortURL:    baseURL + "/123",
					OriginalURL: "http://example.com",
				},
			},
		},
		{
			name:     "POST method not allowed",
			method:   http.MethodPost,
			records:  []storage.Record{},
			userID:   userID,
			wantCode: http.StatusMethodNotAllowed,
			wantOut:  []UserUrlsOut{},
		},
		{
			name:     "PUT method not allowed",
			method:   http.MethodPut,
			records:  []storage.Record{},
			userID:   userID,
			wantCode: http.StatusMethodNotAllowed,
			wantOut:  []UserUrlsOut{},
		},
		{
			name:     "DELETE method not allowed",
			method:   http.MethodDelete,
			records:  []storage.Record{},
			userID:   userID,
			wantCode: http.StatusMethodNotAllowed,
			wantOut:  []UserUrlsOut{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			request := httptest.NewRequest(test.method, "/api/user/urls", nil)
			authCookie, err := auth.NewAuthCookie(test.userID)
			require.NoError(t, err)
			request.AddCookie(authCookie)

			w := httptest.NewRecorder()
			repo := storage.SimpleRepository{
				Records: test.records,
			}
			handler := NewUserUrlsHandler(&repo, baseURL)

			handler.RetrieveUserURLs(w, request)

			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, test.wantCode, res.StatusCode)
			if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
				return
			}

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			var buf bytes.Buffer
			_, err = buf.ReadFrom(res.Body)
			require.NoError(t, err)

			var out []UserUrlsOut
			err = json.Unmarshal(buf.Bytes(), &out)
			require.NoError(t, err)

			assert.ElementsMatch(t, test.wantOut, out)
		})
	}
}
