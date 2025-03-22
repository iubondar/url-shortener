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

			assert.Equal(t, test.wantCode, res.StatusCode)
			assert.ElementsMatch(t, test.wantRecords, repo.Records)
		})
	}
}
