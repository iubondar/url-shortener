package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteUrlsHandler_DeleteUserURLs(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name        string
		method      string
		records     []storage.Record
		body        string
		userID      uuid.UUID
		wantCode    int
		wantRecords []storage.Record
	}{
		{
			name:   "Positive test",
			method: http.MethodDelete,
			records: []storage.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
			},
			body:     "[\"123\"]",
			userID:   userID,
			wantCode: http.StatusAccepted,
			wantRecords: []storage.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   true,
				},
			},
		},
		{
			name:        "Invalid JSON",
			method:      http.MethodDelete,
			records:     []storage.Record{},
			body:        "[\"123\", 456,]",
			userID:      userID,
			wantCode:    http.StatusBadRequest,
			wantRecords: []storage.Record{},
		},
		{
			name:        "GET method not allowed",
			method:      http.MethodGet,
			records:     []storage.Record{},
			body:        "[\"123\"]",
			userID:      userID,
			wantCode:    http.StatusMethodNotAllowed,
			wantRecords: []storage.Record{},
		},
		{
			name:        "POST method not allowed",
			method:      http.MethodPost,
			records:     []storage.Record{},
			body:        "[\"123\"]",
			userID:      userID,
			wantCode:    http.StatusMethodNotAllowed,
			wantRecords: []storage.Record{},
		},
		{
			name:        "PUT method not allowed",
			method:      http.MethodPut,
			records:     []storage.Record{},
			body:        "[\"123\"]",
			userID:      userID,
			wantCode:    http.StatusMethodNotAllowed,
			wantRecords: []storage.Record{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			request := httptest.NewRequest(test.method, "/api/user/urls", bytes.NewReader([]byte(test.body)))
			authCookie, err := auth.NewAuthCookie(test.userID)
			require.NoError(t, err)
			request.AddCookie(authCookie)

			w := httptest.NewRecorder()
			repo := storage.SimpleRepository{
				Records: test.records,
			}
			handler := NewDeleteUrlsHandler(&repo)

			handler.DeleteUserURLs(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.wantCode, res.StatusCode)
			assert.ElementsMatch(t, test.wantRecords, repo.Records)
		})
	}
}
