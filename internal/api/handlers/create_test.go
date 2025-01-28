package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIdHandler_CreateId(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "Positive test",
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				response:    `http://example.com/`,
				contentType: "text/plain",
			},
		},
		{
			name:   "Test GET method not allowed",
			method: http.MethodGet,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://example.com/`,
				contentType: "",
			},
		},
		{
			name:   "Test PUT method not allowed",
			method: http.MethodPut,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://example.com/`,
				contentType: "",
			},
		},
		{
			name:   "Test DELETE method not allowed",
			method: http.MethodDelete,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://example.com/`,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "https://practicum.yandex.ru/"
			request := httptest.NewRequest(test.method, "/", bytes.NewReader([]byte(url)))
			// создаём новый Recorder
			w := httptest.NewRecorder()
			repo := storage.NewSimpleRepository()
			handler := NewCreateIdHandler(repo, "127.0.0.1")
			handler.CreateId(w, request)

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

			id, err := repo.RetrieveId(url)
			require.NoError(t, err)
			assert.Equal(t, test.want.response+id, string(resBody))
		})
	}
}
