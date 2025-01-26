package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iubondar/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrieveURLHandler_RetrieveURL(t *testing.T) {
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "Positive test",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name:   "Test POST method not allowed",
			method: http.MethodPost,
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
			},
		},
		{
			name:   "Test PUT method not allowed",
			method: http.MethodPut,
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
			},
		},
		{
			name:   "Test DELETE method not allowed",
			method: http.MethodDelete,
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := storage.SimpleRepository{
				UrlsToIds: map[string]string{"https://practicum.yandex.ru/": "123"},
				IdsToURLs: map[string]string{"123": "https://practicum.yandex.ru/"},
			}
			handler := NewRetrieveURLHandler(repo)

			request := httptest.NewRequest(test.method, "/", nil)
			request.SetPathValue("id", "123")

			// создаём новый Recorder
			w := httptest.NewRecorder()

			handler.RetrieveURL(w, request)

			res := w.Result()
			// проверяем код ответа, выходим если он ошибочный
			require.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
				return
			}

			// проверяем нужный заголовок
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}

func TestRetrieveURLHandler_WithNoIdParameter(t *testing.T) {
	repo := storage.SimpleRepository{
		UrlsToIds: map[string]string{"https://practicum.yandex.ru/": "123"},
		IdsToURLs: map[string]string{"123": "https://practicum.yandex.ru/"},
	}
	handler := NewRetrieveURLHandler(repo)
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	// создаём новый Recorder
	w := httptest.NewRecorder()

	handler.RetrieveURL(w, request)

	res := w.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRetrieveURLHandler_WithNoURL(t *testing.T) {
	handler := NewRetrieveURLHandler(storage.NewSimpleRepository())
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("id", "123")

	// создаём новый Recorder
	w := httptest.NewRecorder()

	handler.RetrieveURL(w, request)

	res := w.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// {
// 	name:   "URL not found test",
// 	method: http.MethodGet,
// 	want: want{
// 		code:     http.StatusTemporaryRedirect,
// 		location: ``,
// 	},
// },
