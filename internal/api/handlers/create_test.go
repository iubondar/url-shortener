package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIDHandler_CreateID(t *testing.T) {
	userID := uuid.New()
	type fields struct {
		records []storage.Record
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		fields fields
		want   want
	}{
		{
			name:   "Positive test",
			method: http.MethodPost,
			url:    "https://practicum.yandex.ru/",
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusCreated,
				response:    `http://127.0.0.1`,
				contentType: "text/plain",
			},
		},
		{
			name:   "Existed record test",
			method: http.MethodPost,
			url:    testURL,
			fields: fields{
				records: []storage.Record{
					{
						ShortURL:    "123",
						OriginalURL: testURL,
						UserID:      userID,
					},
				},
			},
			want: want{
				code:        http.StatusConflict,
				response:    `http://127.0.0.1`,
				contentType: "text/plain",
			},
		},
		{
			name:   "Test invalid URL",
			method: http.MethodPost,
			url:    "https/practicum.yandex.ru/",
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    `http://127.0.0.1`,
				contentType: "text/plain",
			},
		},
		{
			name:   "Test GET method not allowed",
			method: http.MethodGet,
			url:    "https://practicum.yandex.ru/",
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:   "Test PUT method not allowed",
			method: http.MethodPut,
			url:    "https://practicum.yandex.ru/",
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:   "Test DELETE method not allowed",
			method: http.MethodDelete,
			url:    "https://practicum.yandex.ru/",
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", bytes.NewReader([]byte(test.url)))
			// создаём новый Recorder
			w := httptest.NewRecorder()
			repo := storage.SimpleRepository{
				Records: test.fields.records,
			}
			handler := NewCreateIDHandler(repo, "127.0.0.1")
			handler.CreateID(w, request)

			res := w.Result()
			// проверяем код ответа, выходим если он ошибочный
			require.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
				return
			}

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			id, err := repo.RetrieveID(test.url)
			require.NoError(t, err)
			assert.Equal(t, test.want.response+"/"+id, string(resBody))
		})
	}
}
